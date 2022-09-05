package db_client

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/constants/runtime"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
	"time"
)

func EstablishConnection(ctx context.Context, connStr string, maxConnections int) (*pgxpool.Pool, error) {
	utils.LogTime("db_client.EstablishConnection start")
	defer utils.LogTime("db_client.EstablishConnection end")

	const (
		connMaxIdleTime = 1 * time.Minute
		connMaxLifetime = 10 * time.Minute
	)
	connConfig, _ := pgx.ParseConfig(connStr)
	connConfig.RuntimeParams = map[string]string{
		// set an app name so that we can track connections from this execution
		"application_name":        runtime.PgClientAppName,
		"pool_max_conns":          fmt.Sprintf("%d", maxConnections),
		"pool_max_conn_lifetime":  connMaxIdleTime.String(),
		"pool_max_conn_idle_time": connMaxLifetime.String(),
	}
	connStr = stdlib.RegisterConnConfig(connConfig)

	// this returns connection pool
	dbPool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	//
	//maxParallel := constants.DefaultMaxConnections
	//if viper.IsSet(constants.ArgMaxParallel) {
	//	maxParallel = viper.GetInt(constants.ArgMaxParallel)
	//}
	//
	//// set max open connections to the max connections argument
	//db.SetMaxOpenConns(maxParallel)
	//// NOTE: leave max idle connections at default of 2
	//// close idle connections after 1 minute
	//db.SetConnMaxIdleTime(1 * time.Minute)
	//// do not re-use a connection more than 10 minutes old - force a refresh
	//db.SetConnMaxLifetime(10 * time.Minute)

	if err := db_common.WaitForConnection(ctx, dbPool); err != nil {
		return nil, err
	}
	return dbPool, nil
}

func maxDbConnections() int {
	maxParallel := constants.DefaultMaxConnections
	if viper.IsSet(constants.ArgMaxParallel) {
		maxParallel = viper.GetInt(constants.ArgMaxParallel)
	}
	return maxParallel
}
