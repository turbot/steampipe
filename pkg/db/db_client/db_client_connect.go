package db_client

import (
	"context"
	"database/sql"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
)

const (
	MaxConnLifeTime = 10 * time.Minute
	MaxConnIdleTime = 1 * time.Minute
)

func getDriverNameFromConnectionString(connStr string) string {
	return connStr
}

type DbConnectionCallback func(context.Context, *sql.Conn) error

func (c *DbClient) establishConnectionPool(ctx context.Context, overrides clientConfig) error {
	utils.LogTime("db_client.establishConnectionPool start")
	defer utils.LogTime("db_client.establishConnectionPool end")

	pool, err := establishConnectionPool(ctx, c.connectionString)
	if err != nil {
		return err
	}

	// TODO - how do we apply the AfterConnect hook here?
	// the after connect hook used to create and populate the introspection tables

	// apply any overrides
	// this is used to set the pool size and lifetimes of the connections from up top
	overrides.userPoolSettings.apply(pool)

	err = db_common.WaitForPool(
		ctx,
		pool,
		db_common.WithRetryInterval(constants.DBConnectionRetryBackoff),
		db_common.WithTimeout(time.Duration(viper.GetInt(constants.ArgDatabaseStartTimeout))*time.Second),
	)
	if err != nil {
		return err
	}
	c.userPool = pool

	return c.establishManagementConnectionPool(ctx, overrides)
}

// establishSystemConnectionPool creates a connection pool to use to execute
// system-initiated queries (loading of connection state etc.)
// unlike establishConnectionPool, which is run first to create the user-query pool
// this doesn't wait for the pool to completely start, as establishConnectionPool will have established and verified a connection with the service
func (c *DbClient) establishManagementConnectionPool(ctx context.Context, overrides clientConfig) error {
	utils.LogTime("db_client.establishManagementConnectionPool start")
	defer utils.LogTime("db_client.establishManagementConnectionPool end")

	pool, err := establishConnectionPool(ctx, c.connectionString)

	// apply any overrides
	// this is used to set the pool size and lifetimes of the connections from up top
	overrides.managementPoolSettings.apply(pool)

	err = db_common.WaitForPool(
		ctx,
		pool,
		db_common.WithRetryInterval(constants.DBConnectionRetryBackoff),
		db_common.WithTimeout(time.Duration(viper.GetInt(constants.ArgDatabaseStartTimeout))*time.Second),
	)
	if err != nil {
		return err
	}
	c.userPool = pool

	return c.establishManagementConnectionPool(ctx, overrides)
}

func establishConnectionPool(ctx context.Context, connectionString string) (*sql.DB, error) {
	pool, err := sql.Open(getDriverNameFromConnectionString(connectionString), connectionString)
	if err != nil {
		return nil, err
	}
	pool.SetConnMaxIdleTime(MaxConnIdleTime)
	pool.SetConnMaxLifetime(MaxConnLifeTime)
	pool.SetMaxOpenConns(db_common.MaxDbConnections())
	return pool, nil
}
