package db_client

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type DbConnectionCallback func(context.Context, *pgx.Conn) error

type dbClientConnectionConfig struct {
	recycleConnections bool
	connectionCallback DbConnectionCallback
}

func newDbClientConnectionConfig() *dbClientConnectionConfig {
	c := &dbClientConnectionConfig{
		recycleConnections: true,
	}
	return c
}

type DbClientConnectionOption func(*dbClientConnectionConfig)

func WithConnectionRecycleDisabled() DbClientConnectionOption {
	return func(dcc *dbClientConnectionConfig) {
		dcc.recycleConnections = false
	}
}

func WithConnectionCallback(cb DbConnectionCallback) DbClientConnectionOption {
	return func(dcc *dbClientConnectionConfig) {
		dcc.connectionCallback = cb
	}
}
