package db_client

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type DbConnectionCallback func(context.Context, *pgx.Conn) error

type dbClientConnectionConfig struct {
	maxIdleTime        time.Duration
	maxLifeTime        time.Duration
	connectionCallback DbConnectionCallback
}

func newDbClientConnectionConfig() *dbClientConnectionConfig {
	c := &dbClientConnectionConfig{
		maxIdleTime: 1 * time.Minute,
		maxLifeTime: 10 * time.Minute,
	}
	return c
}

type DbClientConnectionOption func(*dbClientConnectionConfig)

func WithMaxLifeTime(d time.Duration) DbClientConnectionOption {
	return func(dcc *dbClientConnectionConfig) {
		dcc.maxLifeTime = d
	}
}

func WithMaxIdleTime(d time.Duration) DbClientConnectionOption {
	return func(dcc *dbClientConnectionConfig) {
		dcc.maxIdleTime = d
	}
}

func WithConnectionCallback(cb DbConnectionCallback) DbClientConnectionOption {
	return func(dcc *dbClientConnectionConfig) {
		dcc.connectionCallback = cb
	}
}
