package db_client

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

// TestCreateManagementPoolConfig tests the createManagementPoolConfig function
func TestCreateManagementPoolConfig(t *testing.T) {
	tests := map[string]struct {
		setupConfig func() *pgxpool.Config
		overrides   clientConfig
	}{
		"basic config": {
			setupConfig: func() *pgxpool.Config {
				config, _ := pgxpool.ParseConfig("postgresql://localhost:5432/test")
				config.MinConns = 1
				config.MaxConns = 10
				config.MaxConnLifetime = 10 * time.Minute
				config.MaxConnIdleTime = 1 * time.Minute
				return config
			},
			overrides: clientConfig{},
		},
		"config with runtime params": {
			setupConfig: func() *pgxpool.Config {
				config, _ := pgxpool.ParseConfig("postgresql://localhost:5432/test")
				config.ConnConfig.Config.RuntimeParams = map[string]string{
					"application_name": "test_app",
				}
				return config
			},
			overrides: clientConfig{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			config := tc.setupConfig()

			result := createManagementPoolConfig(config, tc.overrides)

			// Verify it's a copy, not the same instance
			assert.NotSame(t, config, result)

			// Verify AfterConnect is removed
			assert.Nil(t, result.AfterConnect)

			// Verify runtime params are set correctly
			appName, ok := result.ConnConfig.Config.RuntimeParams["application_name"]
			assert.True(t, ok)
			assert.Contains(t, appName, "system") // Should contain system or similar
		})
	}
}

