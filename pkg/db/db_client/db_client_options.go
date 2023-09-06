package db_client

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolOverrides struct {
	Size        int
	MaxLifeTime time.Duration
	MaxIdleTime time.Duration
}

// applies the values in the given config if they are non-zero in PoolOverrides
func (c PoolOverrides) apply(config *pgxpool.Config) {
	if c.Size > 0 {
		config.MaxConns = int32(c.Size)
	}
	if c.MaxLifeTime > 0 {
		config.MaxConnLifetime = c.MaxLifeTime
	}
	if c.MaxIdleTime > 0 {
		config.MaxConnIdleTime = c.MaxIdleTime
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
