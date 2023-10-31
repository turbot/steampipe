package db_local

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/constants/runtime"
	"github.com/turbot/pipe-fittings/db_common"
	"github.com/turbot/pipe-fittings/statushooks"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

func getLocalSteampipeConnectionString(opts *DbOptions) (string, error) {
	if opts == nil {
		opts = &DbOptions{}
	}
	utils.LogTime("db.createDbClient start")
	defer utils.LogTime("db.createDbClient end")

	// load the db status
	info, err := GetState()
	if err != nil {
		return "", err
	}
	if info == nil {
		return "", fmt.Errorf("steampipe service is not running")
	}
	if info.ResolvedListenAddresses == nil {
		return "", fmt.Errorf("steampipe service is in unknown state")
	}

	// if no database name is passed, use constants.DatabaseUser
	if len(opts.Username) == 0 {
		opts.Username = constants.DatabaseUser
	}
	// if no username name is passed, deduce it from the db status
	if len(opts.DatabaseName) == 0 {
		opts.DatabaseName = info.Database
	}
	// if we still don't have it, fallback to default "postgres"
	if len(opts.DatabaseName) == 0 {
		opts.DatabaseName = "postgres"
	}

	psqlInfoMap := map[string]string{
		"host":   utils.GetFirstListenAddress(info.ResolvedListenAddresses),
		"port":   fmt.Sprintf("%d", info.Port),
		"user":   opts.Username,
		"dbname": opts.DatabaseName,
	}
	log.Println("[TRACE] SQLInfoMap >>>", psqlInfoMap)
	psqlInfoMap = utils.MergeMaps(psqlInfoMap, dsnSSLParams())
	log.Println("[TRACE] SQLInfoMap >>>", psqlInfoMap)

	psqlInfo := []string{}
	for k, v := range psqlInfoMap {
		psqlInfo = append(psqlInfo, fmt.Sprintf("%s=%s", k, v))
	}
	log.Println("[TRACE] PSQLInfo >>>", psqlInfo)

	return strings.Join(psqlInfo, " "), nil
}

type DbOptions struct {
	DatabaseName   string
	Username       string
	MaxConnections int
}

// CreateLocalDbConnectionPool connects and returns a connection to the given database using
// the provided username
// if the database is not provided (empty), it connects to the default database in the service
// that was created during installation.
// NOTE: no session data callback is used - no session data will be present
func CreateLocalDbConnectionPool(ctx context.Context, opts *DbOptions) (*sql.DB, error) {
	utils.LogTime("db.CreateLocalDbConnection start")
	defer utils.LogTime("db.CreateLocalDbConnection end")

	psqlInfo, err := getLocalSteampipeConnectionString(opts)
	if err != nil {
		return nil, err
	}
	return CreateDbConnectionPool(ctx, psqlInfo, opts)
}

// CreateDbConnectionPool connects and returns a connection to the database using the given connection string
func CreateDbConnectionPool(ctx context.Context, connectionString string, opts *DbOptions) (*sql.DB, error) {
	utils.LogTime("db.CreateDbConnectionPool start")
	defer utils.LogTime("db.CreateDbConnectionPool end")

	// TODO KAI is this needed

	// err = db_common.AddRootCertToConfig(&connConfig.Config, localfilepaths.GetRootCertLocation())
	// if err != nil {
	// 	return nil, err
	// }

	pool, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	const (
		connMaxIdleTime = 1 * time.Minute
		connMaxLifetime = 10 * time.Minute
	)
	pool.SetMaxOpenConns(opts.MaxConnections)
	pool.SetConnMaxLifetime(connMaxLifetime)
	pool.SetConnMaxIdleTime(connMaxIdleTime)

	err = db_common.WaitForPool(
		ctx,
		pool,
		db_common.WithRetryInterval(constants.DBConnectionRetryBackoff),
		db_common.WithTimeout(time.Duration(viper.GetInt(constants.ArgDatabaseStartTimeout))*time.Second),
	)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

// createMaintenanceClient connects to the postgres server using the
// maintenance database (postgres) and superuser
// this is used in a couple of places
//  1. During installation to setup the DBMS with foreign_server, extension et.al.
//  2. During service start and stop to query the DBMS for parameters (connected clients, database name etc.)
//
// this is called immediately after the service process is started and hence
// all special handling related to service startup failures SHOULD be handled here
func createMaintenanceClient(ctx context.Context, port int) (*sql.Conn, error) {
	utils.LogTime("db_local.createMaintenanceClient start")
	defer utils.LogTime("db_local.createMaintenanceClient end")

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(viper.GetInt(constants.ArgDatabaseStartTimeout))*time.Second)
	defer cancel()

	statushooks.SetStatus(ctx, "Waiting for connection")
	// TODO kai move connection string logic somewhere central
	connStr := fmt.Sprintf("host=127.0.0.1 port=%d user=%s dbname=postgres sslmode=disable application_name=%s",
		port,
		constants.DatabaseSuperUser,
		runtime.ServiceConnectionAppName)
	opts := &DbOptions{
		Username:       constants.DatabaseSuperUser,
		MaxConnections: 1,
	}
	tempPool, err := CreateDbConnectionPool(ctx, connStr, opts)
	if err != nil {
		log.Println("[TRACE] could not connect to service")
		return nil, sperr.Wrap(err, sperr.WithMessage("connection setup failed"))
	}
	conn, err := tempPool.Conn(timeoutCtx)
	if err != nil {
		log.Println("[TRACE] could not connect to service")
		return nil, sperr.Wrap(err, sperr.WithMessage("connection setup failed"))
	}

	// wait for recovery to complete
	// the database may enter recovery mode if it detects that
	// it wasn't shutdown gracefully.
	// For large databases, this can take long
	// We want to wait for a LONG time for this to complete
	// Use the context that was given - since that is tied to os.Signal
	// and can be interrupted
	err = db_common.WaitForRecovery(
		ctx,
		conn,
		db_common.WithRetryInterval(constants.DBRecoveryRetryBackoff),
	)
	if err != nil {
		conn.Close()
		log.Println("[TRACE] WaitForRecovery timed out")
		return nil, sperr.Wrap(err, sperr.WithMessage("could not complete recovery"))
	}

	return conn, nil
}
