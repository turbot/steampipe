package db_client

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/constants/runtime"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
)

type DbConnectionCallback func(context.Context, *pgx.Conn) error

func (c *DbClient) establishConnectionPool(ctx context.Context) error {
	utils.LogTime("db_client.establishConnectionPool start")
	defer utils.LogTime("db_client.establishConnectionPool end")

	const (
		connMaxIdleTime = 1 * time.Minute
		connMaxLifetime = 10 * time.Minute
	)
	maxConnections := db_common.MaxDbConnections()

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
	if helpers.StringSliceContains(locals, config.ConnConfig.Host) {
		c.isLocalService = true
	}

	// MinConns should default to 0, but when not set, it actually get very high values (e.g. 80217984)
	// this leads to a huge number of connections getting created
	// TODO BINAEK dig into this and figure out why this is happening.
	// We need to be sure that it is not an issue with service management
	config.MinConns = 0
	config.MaxConns = int32(maxConnections)
	config.MaxConnLifetime = connMaxLifetime
	config.MaxConnIdleTime = connMaxIdleTime
	if c.onConnectionCallback != nil {
		config.AfterConnect = c.onConnectionCallback
	}
	// set an app name so that we can track database connections from this Steampipe execution
	// this is used to determine whether the database can safely be closed
	config.ConnConfig.Config.RuntimeParams = map[string]string{
		constants.RuntimeParamsKeyApplicationName: runtime.ClientConnectionAppName,
	}

	// this returns connection pool
	dbPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}

	err = db_common.WaitForPool(
		ctx,
		dbPool,
		db_common.WithRetryInterval(constants.DBConnectionRetryBackoff),
		db_common.WithTimeout(time.Duration(viper.GetInt(constants.ArgDatabaseStartTimeout))*time.Second),
	)
	if err != nil {
		return err
	}
	c.userPool = dbPool

	return c.establishManagementConnectionPool(ctx, config)
}

// establishSystemConnectionPool creates a connection pool to use to execute
// system-initiated queries (loading of connection state etc.)
// unlike establishConnectionPool, which is run first to create the user-query pool
// this doesn't wait for the pool to completely start, as establishConnectionPool will have established and verified a connection with the service
func (c *DbClient) establishManagementConnectionPool(ctx context.Context, config *pgxpool.Config) error {
	utils.LogTime("db_client.establishSystemConnectionPool start")
	defer utils.LogTime("db_client.establishSystemConnectionPool end")

	// create a copy of the config which is used to setup the user connection pool
	// we need to modify the config - don't want the original to get modified
	copiedConfig := config.Copy()
	copiedConfig.ConnConfig.Config.RuntimeParams = map[string]string{
		constants.RuntimeParamsKeyApplicationName: runtime.ClientSystemConnectionAppName,
	}

	// remove the afterConnect hook - we don't need the session data in management connections
	copiedConfig.AfterConnect = nil

	if copiedConfig.MaxConns == 1 {
		// we need a few extra connections in the management pool, since this pool
		// is used to LISTEN and search path resolution
		copiedConfig.MaxConns = config.MaxConns + constants.ManagementConnectionDelta
	}

	// this returns connection pool
	dbPool, err := pgxpool.NewWithConfig(context.Background(), copiedConfig)
	if err != nil {
		return err
	}
	c.managementPool = dbPool
	return nil
}
