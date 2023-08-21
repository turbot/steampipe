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
		if dcc.connectionCallback != nil {
			originalCallback := dcc.connectionCallback
			wrappedCallback := func(ctx context.Context, c *pgx.Conn) error {
				// call the original one first
				if err := originalCallback(ctx, c); err != nil {
					return err
				}
				// then this one
				if err := cb(ctx, c); err != nil {
					return err
				}
				return nil
			}
			dcc.connectionCallback = wrappedCallback
			return
		}
		dcc.connectionCallback = cb
	}
}
