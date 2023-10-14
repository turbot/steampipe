package db_local

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/constants/runtime"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
)

func getLocalSteampipeConnectionString(opts *CreateDbOptions) (string, error) {
	if opts == nil {
		opts = &CreateDbOptions{}
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
		// HACK
		opts.Username = "root" //constants.DatabaseUser
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

type CreateDbOptions struct {
	DatabaseName, Username string
}

// CreateLocalDbConnection connects and returns a connection to the given database using
// the provided username
// if the database is not provided (empty), it connects to the default database in the service
// that was created during installation.
// NOTE: no session data callback is used - no session data will be present
func CreateLocalDbConnection(ctx context.Context, opts *CreateDbOptions) (*pgx.Conn, error) {
	utils.LogTime("db.CreateLocalDbConnection start")
	defer utils.LogTime("db.CreateLocalDbConnection end")

	psqlInfo, err := getLocalSteampipeConnectionString(opts)
	if err != nil {
		return nil, err
	}

	connConfig, err := pgx.ParseConfig(psqlInfo)
	if err != nil {
		return nil, err
	}

	// set an app name so that we can track database connections from this Steampipe execution
	// this is used to determine whether the database can safely be closed
	connConfig.Config.RuntimeParams = map[string]string{
		constants.RuntimeParamsKeyApplicationName: runtime.ServiceConnectionAppName,
	}
	err = db_common.AddRootCertToConfig(&connConfig.Config, filepaths.GetRootCertLocation())
	if err != nil {
		return nil, err
	}

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, err
	}

	if err := db_common.WaitForConnectionPing(ctx, conn); err != nil {
		return nil, err
	}
	return conn, nil
}

// CreateConnectionPool
func CreateConnectionPool(ctx context.Context, opts *CreateDbOptions, maxConnections int) (*pgxpool.Pool, error) {
	utils.LogTime("db_client.establishConnectionPool start")
	defer utils.LogTime("db_client.establishConnectionPool end")

	psqlInfo, err := getLocalSteampipeConnectionString(opts)
	if err != nil {
		return nil, err
	}

	poolConfig, err := pgxpool.ParseConfig(psqlInfo)
	if err != nil {
		return nil, err
	}

	const (
		connMaxIdleTime = 1 * time.Minute
		connMaxLifetime = 10 * time.Minute
	)

	poolConfig.MinConns = 0
	poolConfig.MaxConns = int32(maxConnections)
	poolConfig.MaxConnLifetime = connMaxLifetime
	poolConfig.MaxConnIdleTime = connMaxIdleTime

	poolConfig.ConnConfig.Config.RuntimeParams = map[string]string{
		constants.RuntimeParamsKeyApplicationName: runtime.ServiceConnectionAppName,
	}

	// this returns connection pool
	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	err = db_common.WaitForPool(
		ctx,
		dbPool,
		db_common.WithRetryInterval(constants.DBConnectionRetryBackoff),
		db_common.WithTimeout(time.Duration(viper.GetInt(constants.ArgDatabaseStartTimeout))*time.Second),
	)
	if err != nil {
		return nil, err
	}
	return dbPool, nil
}

// createMaintenanceClient connects to the postgres server using the
// maintenance database (postgres) and superuser
// this is used in a couple of places
//  1. During installation to setup the DBMS with foreign_server, extension et.al.
//  2. During service start and stop to query the DBMS for parameters (connected clients, database name etc.)
//
// this is called immediately after the service process is started and hence
// all special handling related to service startup failures SHOULD be handled here
func createMaintenanceClient(ctx context.Context, port int) (*pgx.Conn, error) {
	utils.LogTime("db_local.createMaintenanceClient start")
	defer utils.LogTime("db_local.createMaintenanceClient end")

	connStr := fmt.Sprintf("host=127.0.0.1 port=%d user=%s dbname=postgres sslmode=disable application_name=%s", port, constants.DatabaseSuperUser, runtime.ServiceConnectionAppName)

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(viper.GetInt(constants.ArgDatabaseStartTimeout))*time.Second)
	defer cancel()

	statushooks.SetStatus(ctx, "Waiting for connection")
	conn, err := db_common.WaitForConnection(
		timeoutCtx,
		connStr,
		db_common.WithRetryInterval(constants.DBConnectionRetryBackoff),
		db_common.WithTimeout(time.Duration(viper.GetInt(constants.ArgDatabaseStartTimeout))*time.Second),
	)
	if err != nil {
		log.Println("[TRACE] could not connect to service")
		return nil, sperr.Wrap(err, sperr.WithMessage("connection setup failed"))
	}

	// wait for db to start accepting queries on this connection
	err = db_common.WaitForConnectionPing(
		timeoutCtx,
		conn,
		db_common.WithRetryInterval(constants.DBConnectionRetryBackoff),
		db_common.WithTimeout(viper.GetDuration(constants.ArgDatabaseStartTimeout)*time.Second),
	)
	if err != nil {
		conn.Close(ctx)
		log.Println("[TRACE] Ping timed out")
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
		conn.Close(ctx)
		log.Println("[TRACE] WaitForRecovery timed out")
		return nil, sperr.Wrap(err, sperr.WithMessage("could not complete recovery"))
	}

	return conn, nil
}
