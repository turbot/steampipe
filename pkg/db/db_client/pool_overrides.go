package db_client

import "time"

type poolConfigOverrides struct {
	userPoolSize          int
	managementPoolSize    int
	maxConnectionLifeTime time.Duration
	maxConnectionIdleTime time.Duration
}

type PoolOverride func(*poolConfigOverrides)

func WithUserPoolSize(size int) PoolOverride {
	return func(po *poolConfigOverrides) {
		po.userPoolSize = size
	}
}

func WithManagementPoolSize(size int) PoolOverride {
	return func(po *poolConfigOverrides) {
		po.managementPoolSize = size
	}
}

func WithMaxLife(t time.Duration) PoolOverride {
	return func(po *poolConfigOverrides) {
		po.maxConnectionLifeTime = t
	}
}

func WithMaxIdle(t time.Duration) PoolOverride {
	return func(po *poolConfigOverrides) {
		po.maxConnectionIdleTime = t
	}
}
