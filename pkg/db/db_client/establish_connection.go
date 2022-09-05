package db_client

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
	"time"
)

type DbConnectionCallback func(context.Context, *pgx.Conn) error

func EstablishConnection(ctx context.Context, connStr string, minConnections, maxConnections int, connectionCallback DbConnectionCallback) (*pgxpool.Pool, error) {
	utils.LogTime("db_client.EstablishConnection start")
	defer utils.LogTime("db_client.EstablishConnection end")
	const (
		connMaxIdleTime = 1 * time.Minute
		connMaxLifetime = 10 * time.Minute
	)
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

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	config.MaxConns = int32(maxConnections)
	config.MinConns = int32(minConnections)
	config.MaxConnLifetime = connMaxLifetime
	config.MaxConnIdleTime = connMaxIdleTime
	if connectionCallback != nil {
		config.AfterConnect = connectionCallback
	}

	log.Printf("[WARN] EstablishConnection %v", config)

	// this returns connection pool
	dbPool, err := pgxpool.ConnectConfig(context.Background(), config)
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
