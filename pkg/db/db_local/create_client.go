package db_local

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/constants/runtime"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
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
		// Connect to the database using the first listen address, which is usually localhost
		"host":   info.Listen[0],
		"port":   fmt.Sprintf("%d", info.Port),
		"user":   opts.Username,
		"dbname": opts.DatabaseName,
	}
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

// createLocalDbClient connects and returns a connection to the given database using
// the provided username
// if the database is not provided (empty), it connects to the default database in the service
// that was created during installation.
// NOTE: no session data callback is used - no sesison data will be present
func createLocalDbClient(ctx context.Context, opts *CreateDbOptions) (*pgx.Conn, error) {
	utils.LogTime("db.createLocalDbClient start")
	defer utils.LogTime("db.createLocalDbClient end")

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
		"application_name": runtime.PgClientAppName,
	}
	err = db_common.AddRootCertToConfig(&connConfig.Config, getRootCertLocation())
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

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(viper.GetInt(constants.ArgDatabaseStartTimeout))*time.Second)
	defer cancel()

	connStr := fmt.Sprintf("host=localhost port=%d user=%s dbname=postgres sslmode=disable", port, constants.DatabaseSuperUser)
	conn, err := db_common.WaitForConnection(
		timeoutCtx,
		connStr,
		db_common.WithRetryInterval(constants.DBRecoveryRetryBackoff),
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
		return nil, err
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
		db_common.WithTimeout(constants.DBRecoveryTimeout),
	)
	if err != nil {
		conn.Close(ctx)
		log.Println("[TRACE] WaitForRecovery timed out")
		return nil, sperr.Wrap(err, sperr.WithMessage("timed out waiting for recovery to complete"))
	}

	return conn, nil
}
