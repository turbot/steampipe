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

	// this will yield a false negative when connecting to a local service using a network IP
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
	// disable automatic prepared statements
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

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
	c.pool = dbPool

	return c.establishSystemConnectionPool(ctx, config)
}

// establishSystemConnectionPool creates a connection pool that can be used to support execution of user-initiated
// queries (loading of connection state etc.)
// unlike the other connection pool, this doesn't wait for the pool to completely start
// this is because, the other pool will have established and verified a connection with the service
func (c *DbClient) establishSystemConnectionPool(ctx context.Context, config *pgxpool.Config) error {
	utils.LogTime("db_client.establishSystemConnectionPool start")
	defer utils.LogTime("db_client.establishSystemConnectionPool end")

	// create a copy of the config
	copiedConfig := config.Copy()
	copiedConfig.ConnConfig.Config.RuntimeParams = map[string]string{
		"application_name": runtime.ClientSystemConnectionAppName,
	}

	// this returns connection pool
	dbPool, err := pgxpool.NewWithConfig(context.Background(), copiedConfig)
	if err != nil {
		return err
	}
	c.sysPool = dbPool
	return nil
}
