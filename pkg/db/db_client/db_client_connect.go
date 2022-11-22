package db_client

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
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
	minConnections := 2
	maxConnections := maxDbConnections()
	if minConnections > maxConnections {
		minConnections = maxConnections
	}

	config, err := pgxpool.ParseConfig(c.connectionString)
	if err != nil {
		return err
	}
	config.MaxConns = int32(maxConnections)
	config.MinConns = int32(minConnections)
	config.MaxConnLifetime = connMaxLifetime
	config.MaxConnIdleTime = connMaxIdleTime
	if c.onConnectionCallback != nil {
		config.AfterConnect = c.onConnectionCallback
	}
	// set an app name so that we can track database connections from this Steampipe execution
	// this is used to determine whether the database can safely be closed
	config.ConnConfig.Config.RuntimeParams = map[string]string{
		"application_name": runtime.PgClientAppName,
	}

	// this returns connection pool
	dbPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}

	if err := db_common.WaitForPool(ctx, dbPool); err != nil {
		return err
	}
	c.pool = dbPool
	return nil
}

func maxDbConnections() int {
	maxParallel := constants.DefaultMaxConnections
	if viper.IsSet(constants.ArgMaxParallel) {
		maxParallel = viper.GetInt(constants.ArgMaxParallel)
	}
	return maxParallel
}
