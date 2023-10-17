package db_client

import (
	"database/sql"
	"time"
)

type PoolOverrides struct {
	Size        int
	MaxLifeTime time.Duration
	MaxIdleTime time.Duration
}

// applies the values in the given config if they are non-zero in PoolOverrides
func (c PoolOverrides) apply(db *sql.DB) {
	if c.Size > 0 {
		db.SetMaxOpenConns(c.Size)
	}
	if c.MaxLifeTime > 0 {
		db.SetConnMaxLifetime(c.MaxLifeTime)
	}
	if c.MaxIdleTime > 0 {
		db.SetConnMaxIdleTime(c.MaxIdleTime)
	}
}

type clientConfig struct {
	userPoolSettings       PoolOverrides
	managementPoolSettings PoolOverrides
}

type ClientOption func(*clientConfig)

func WithUserPoolOverride(s PoolOverrides) ClientOption {
	return func(cc *clientConfig) {
		cc.userPoolSettings = s
	}
}

func WithManagementPoolOverride(s PoolOverrides) ClientOption {
	return func(cc *clientConfig) {
		cc.managementPoolSettings = s
	}
}
