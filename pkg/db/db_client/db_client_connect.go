package db_client

import (
	"context"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/constants/runtime"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

const (
	MaxConnLifeTime = 10 * time.Minute
	MaxConnIdleTime = 1 * time.Minute
)

type DbConnectionCallback func(context.Context, *pgx.Conn) error

func (c *DbClient) establishConnectionPool(ctx context.Context, overrides clientConfig) error {
	utils.LogTime("db_client.establishConnectionPool start")
	defer utils.LogTime("db_client.establishConnectionPool end")

	config, err := pgxpool.ParseConfig(c.connectionString)
	if err != nil {
		return err
	}

	locals := []string{
		"127.0.0.1",
		"::1",
		"localhost",
	}

	// when connected to a service which is running a plugin compiled with SDK pre-v5, the plugin
	// will not have the ability to turn off caching (feature introduced in SDKv5)
	//
	// the 'isLocalService' is used to set the client end cache to 'false' if caching is turned off in the local service
	//
	// this is a temporary workaround to make sure
	// that we can turn off caching for plugins compiled with SDK pre-V5
	// worst case scenario is that we don't switch off the cache for pre-V5 plugins
	// refer to: https://github.com/turbot/steampipe/blob/f7f983a552a07e50e526fcadf2ccbfdb7b247cc0/pkg/db/db_client/db_client_session.go#L66
	if slices.Contains(locals, config.ConnConfig.Host) {
		c.isLocalService = true
	}

	// MinConns should default to 0, but when not set, it actually get very high values (e.g. 80217984)
	// this leads to a huge number of connections getting created
	// TODO BINAEK dig into this and figure out why this is happening.
	// We need to be sure that it is not an issue with service management
	config.MinConns = 0
	config.MaxConns = int32(db_common.MaxDbConnections())
	config.MaxConnLifetime = MaxConnLifeTime
	config.MaxConnIdleTime = MaxConnIdleTime
	if c.onConnectionCallback != nil {
		config.AfterConnect = c.onConnectionCallback
	}
	// Clean up session map when connections are closed to prevent memory leak
	// Reference: https://github.com/turbot/steampipe/issues/3737
	config.BeforeClose = func(conn *pgx.Conn) {
		if conn != nil && conn.PgConn() != nil {
			backendPid := conn.PgConn().PID()
			// Best-effort cleanup: do not block pool.Close() if sessions lock is busy.
			if c.sessionsTryLock() {
				// Check if sessions map has been nil'd by Close()
				if c.sessions != nil {
					delete(c.sessions, backendPid)
				}
				c.sessionsUnlock()
			}
		}
	}
	// set an app name so that we can track database connections from this Steampipe execution
	// this is used to determine whether the database can safely be closed
	config.ConnConfig.Config.RuntimeParams = map[string]string{
		constants.RuntimeParamsKeyApplicationName: runtime.ClientConnectionAppName,
	}

	// apply any overrides
	// this is used to set the pool size and lifetimes of the connections from up top
	overrides.userPoolSettings.apply(config)

	// this returns connection pool
	dbPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}

	err = db_common.WaitForPool(
		ctx,
		dbPool,
		db_common.WithRetryInterval(constants.DBConnectionRetryBackoff),
		db_common.WithTimeout(time.Duration(viper.GetInt(pconstants.ArgDatabaseStartTimeout))*time.Second),
	)
	if err != nil {
		return err
	}
	c.userPool = dbPool

	return c.establishManagementConnectionPool(ctx, config, overrides)
}

// establishManagementConnectionPool creates a connection pool to use to execute
// system-initiated queries (loading of connection state etc.)
// unlike establishConnectionPool, which is run first to create the user-query pool
// this doesn't wait for the pool to completely start, as establishConnectionPool will have established and verified a connection with the service
func (c *DbClient) establishManagementConnectionPool(ctx context.Context, config *pgxpool.Config, overrides clientConfig) error {
	utils.LogTime("db_client.establishSystemConnectionPool start")
	defer utils.LogTime("db_client.establishSystemConnectionPool end")

	// create a config from the config of the user pool
	copiedConfig := createManagementPoolConfig(config, overrides)

	// this returns connection pool
	dbPool, err := pgxpool.NewWithConfig(context.Background(), copiedConfig)
	if err != nil {
		return err
	}
	c.managementPool = dbPool
	return nil
}

func createManagementPoolConfig(config *pgxpool.Config, overrides clientConfig) *pgxpool.Config {
	// create a copy - we will be modifying this
	copiedConfig := config.Copy()

	// update the app name of the connection
	copiedConfig.ConnConfig.Config.RuntimeParams = map[string]string{
		constants.RuntimeParamsKeyApplicationName: runtime.ClientSystemConnectionAppName,
	}

	// remove the afterConnect hook - we don't need the session data in management connections
	copiedConfig.AfterConnect = nil

	overrides.managementPoolSettings.apply(copiedConfig)

	return copiedConfig
}
