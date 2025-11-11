package pluginmanager_service

import (
	"testing"

	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe/v2/pkg/connection"
)

// TestPluginManager_HandlePluginLimiterChanges_NilPool tests that HandlePluginLimiterChanges
// does not panic when the pool is nil. This can happen when rate limiter definitions change
// before the database pool is initialized.
// Issue: https://github.com/turbot/steampipe/issues/4785
func TestPluginManager_HandlePluginLimiterChanges_NilPool(t *testing.T) {
	// Create a PluginManager with nil pool
	pm := &PluginManager{
		pool:           nil, // This is the condition that triggers the bug
		pluginLimiters: nil,
		userLimiters:   make(connection.PluginLimiterMap),
	}

	// Create some test rate limiters
	newLimiters := connection.PluginLimiterMap{
		"aws": connection.LimiterMap{
			"default": &plugin.RateLimiter{
				Plugin: "aws",
				Name:   "default",
				Source: plugin.LimiterSourcePlugin,
				Status: plugin.LimiterStatusActive,
			},
		},
	}

	// This should not panic even though pool is nil
	err := pm.HandlePluginLimiterChanges(newLimiters)

	// We expect an error (or nil), but not a panic
	if err != nil {
		t.Logf("HandlePluginLimiterChanges returned error (expected): %v", err)
	}

	// Verify that the limiters were stored even if table refresh failed
	if pm.pluginLimiters == nil {
		t.Fatal("Expected pluginLimiters to be initialized")
	}

	if _, exists := pm.pluginLimiters["aws"]; !exists {
		t.Error("Expected aws plugin limiters to be stored")
	}
}
