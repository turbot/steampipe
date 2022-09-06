package db_client

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
	"time"
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

	// TODO KAI why application_name???

	//connConfig, _ := pgx.ParseConfig(connStr)
	//connCon
	//fig.RuntimeParams = map[string]string{
	//	// set an app name so that we can track connections from this execution
	//	"application_name": runtime.PgClientAppName,
	//connStr = stdlib.RegisterConnConfig(connConfig)

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

	// this returns connection pool
	dbPool, err := pgxpool.ConnectConfig(context.Background(), config)
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
