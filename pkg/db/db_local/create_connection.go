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
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	putils "github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe-plugin-sdk/v6/sperr"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/constants/runtime"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
)

func getLocalSteampipeConnectionString(opts *CreateDbOptions) (string, error) {
	if opts == nil {
		opts = &CreateDbOptions{}
	}
	putils.LogTime("db.createDbClient start")
	defer putils.LogTime("db.createDbClient end")

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

	var psqlInfoMap map[string]string
	if opts.Username == constants.DatabaseSuperUser {
		// root connects through the unix socket in the data directory
		if err := checkRootSocket(filepaths.GetDataLocation(), info.Port); err != nil {
			return "", err
		}
		psqlInfoMap = rootConnectionParams(info.Port)
		psqlInfoMap["dbname"] = opts.DatabaseName
	} else {
		psqlInfoMap = map[string]string{
			"host":   putils.GetFirstListenAddress(info.ResolvedListenAddresses),
			"port":   fmt.Sprintf("%d", info.Port),
			"user":   opts.Username,
			"dbname": opts.DatabaseName,
		}
		psqlInfoMap = putils.MergeMaps(psqlInfoMap, dsnSSLParams())
	}
	log.Println("[TRACE] SQLInfoMap >>>", psqlInfoMap)

	psqlInfo := []string{}
	for k, v := range psqlInfoMap {
		psqlInfo = append(psqlInfo, fmt.Sprintf("%s=%s", k, quoteConnStringValue(v)))
	}
	log.Println("[TRACE] PSQLInfo >>>", psqlInfo)

	return strings.Join(psqlInfo, " "), nil
}

// quoteConnStringValue quotes a value for a keyword/value connection string, so that
// spaces, quotes and backslashes (in a path or the SSL passphrase) are kept as is
func quoteConnStringValue(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `'`, `\'`)
	return `'` + v + `'`
}

type CreateDbOptions struct {
	DatabaseName, Username string
}

func isRootConnection(opts *CreateDbOptions) bool {
	return opts != nil && opts.Username == constants.DatabaseSuperUser
}

// CreateLocalDbConnection connects and returns a connection to the given database using
// the provided username
// if the database is not provided (empty), it connects to the default database in the service
// that was created during installation.
// NOTE: this connection will use the ServiceConnectionAppName
func CreateLocalDbConnection(ctx context.Context, opts *CreateDbOptions) (*pgx.Conn, error) {
	putils.LogTime("db.CreateLocalDbConnection start")
	defer putils.LogTime("db.CreateLocalDbConnection end")

	psqlInfo, err := getLocalSteampipeConnectionString(opts)
	if err != nil {
		return nil, err
	}

	connConfig, err := pgx.ParseConfig(psqlInfo)
	if err != nil {
		return nil, err
	}

	if isRootConnection(opts) {
		setRootSocketHost(connConfig, filepaths.GetDataLocation())
	}

	// set an app name so that we can track database connections from this Steampipe execution
	// this is used to determine whether the database can safely be closed
	// and also in pipes to allow accurate usage tracking (it excludes system calls)
	connConfig.Config.RuntimeParams = map[string]string{
		constants.RuntimeParamsKeyApplicationName: runtime.ServiceConnectionAppName,
	}
	// the unix socket that root uses does not use SSL
	if connConfig.User != constants.DatabaseSuperUser {
		err = db_common.AddRootCertToConfig(&connConfig.Config, filepaths.GetRootCertLocation())
		if err != nil {
			return nil, err
		}
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

// CreateConnectionPool creates a connection pool using the provided options
// NOTE: this connection pool will use the ServiceConnectionAppName
func CreateConnectionPool(ctx context.Context, opts *CreateDbOptions, maxConnections int) (*pgxpool.Pool, error) {
	putils.LogTime("db_client.establishConnectionPool start")
	defer putils.LogTime("db_client.establishConnectionPool end")

	psqlInfo, err := getLocalSteampipeConnectionString(opts)
	if err != nil {
		return nil, err
	}

	poolConfig, err := pgxpool.ParseConfig(psqlInfo)
	if err != nil {
		return nil, err
	}

	if isRootConnection(opts) {
		setRootSocketHost(poolConfig.ConnConfig, filepaths.GetDataLocation())
	}

	const (
		connMaxIdleTime = 1 * time.Minute
		connMaxLifetime = 10 * time.Minute
	)

	poolConfig.MinConns = 0
	poolConfig.MaxConns = int32(maxConnections) //nolint:gosec // G115: the sole caller (plugin_manager.go) passes a hardcoded literal (20), never user input
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
		db_common.WithTimeout(time.Duration(viper.GetInt(pconstants.ArgDatabaseStartTimeout))*time.Second),
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
func createMaintenanceClient(ctx context.Context, dataDir string, port int) (*pgx.Conn, error) {
	putils.LogTime("db_local.createMaintenanceClient start")
	defer putils.LogTime("db_local.createMaintenanceClient end")

	// the socket appears in the data directory when postgres starts listening
	if err := waitForRootSocket(ctx, dataDir, port); err != nil {
		return nil, sperr.Wrap(err, sperr.WithMessage("connection setup failed"))
	}

	connParams := rootConnectionParams(port)
	connStr := fmt.Sprintf("port=%d user=%s dbname=postgres sslmode=%s application_name=%s", port, connParams["user"], connParams["sslmode"], runtime.ServiceConnectionAppName)
	connConfig, err := pgx.ParseConfig(connStr)
	if err != nil {
		return nil, sperr.Wrap(err, sperr.WithMessage("connection setup failed"))
	}
	setRootSocketHost(connConfig, dataDir)

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(viper.GetInt(pconstants.ArgDatabaseStartTimeout))*time.Second)
	defer cancel()

	statushooks.SetStatus(ctx, "Waiting for connection")
	conn, err := db_common.WaitForConnection(
		timeoutCtx,
		connConfig,
		db_common.WithRetryInterval(constants.DBConnectionRetryBackoff),
		db_common.WithTimeout(time.Duration(viper.GetInt(pconstants.ArgDatabaseStartTimeout))*time.Second),
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
		db_common.WithTimeout(time.Duration(viper.GetInt(pconstants.ArgDatabaseStartTimeout))*time.Second),
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
