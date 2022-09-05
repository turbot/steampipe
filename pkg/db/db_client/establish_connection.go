package db_client

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
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

	// TODO KAI CHECK THIS WORKS

	//connConfig, _ := pgx.ParseConfig(connStr)
	//connCon
	//fig.RuntimeParams = map[string]string{
	//	// set an app name so that we can track connections from this execution
	//	"application_name": runtime.PgClientAppName,
	//	//"pool_max_conns":          fmt.Sprintf("%d", maxConnections),
	//	//"pool_max_conn_lifetime":  connMaxIdleTime.String(),
	//	//"pool_max_conn_idle_time": connMaxLifetime.String(),
	//}
	//connStr = stdlib.RegisterConnConfig(connConfig)

	// this returns connection pool
	dbPool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

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
