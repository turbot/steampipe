package pluginmanager_service

import (
	"sync"
	"testing"

	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe/v2/pkg/connection"
)

// TestPluginManager_ConcurrentUpdateRateLimiterStatus tests for race condition
// when updateRateLimiterStatus is called concurrently with writes to userLimiters map
// References: https://github.com/turbot/steampipe/issues/4786
func TestPluginManager_ConcurrentUpdateRateLimiterStatus(t *testing.T) {
	// Create a PluginManager with test data
	pm := &PluginManager{
		userLimiters: make(connection.PluginLimiterMap),
		pluginLimiters: connection.PluginLimiterMap{
			"aws": connection.LimiterMap{
				"limiter1": &plugin.RateLimiter{
					Name:   "limiter1",
					Plugin: "aws",
					Status: plugin.LimiterStatusActive,
				},
			},
		},
	}

	// Run concurrent operations to trigger race condition
	var wg sync.WaitGroup
	iterations := 100

	// Writer goroutine - simulates handleUserLimiterChanges modifying userLimiters
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			pm.userLimiters = connection.PluginLimiterMap{
				"aws": connection.LimiterMap{
					"limiter1": &plugin.RateLimiter{
						Name:   "limiter1",
						Plugin: "aws",
						Status: plugin.LimiterStatusOverridden,
					},
				},
			}
		}
	}()

	// Reader goroutine - simulates updateRateLimiterStatus reading userLimiters
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			pm.updateRateLimiterStatus()
		}
	}()

	wg.Wait()
}

// TestPluginManager_ConcurrentRateLimiterMapAccess tests for race condition
// when multiple goroutines access pluginLimiters and userLimiters concurrently
func TestPluginManager_ConcurrentRateLimiterMapAccess(t *testing.T) {
	pm := &PluginManager{
		userLimiters: connection.PluginLimiterMap{
			"aws": connection.LimiterMap{
				"limiter1": &plugin.RateLimiter{
					Name:   "limiter1",
					Plugin: "aws",
					Status: plugin.LimiterStatusOverridden,
				},
			},
		},
		pluginLimiters: connection.PluginLimiterMap{
			"aws": connection.LimiterMap{
				"limiter1": &plugin.RateLimiter{
					Name:   "limiter1",
					Plugin: "aws",
					Status: plugin.LimiterStatusActive,
				},
			},
		},
	}

	var wg sync.WaitGroup
	iterations := 50

	// Multiple readers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				pm.updateRateLimiterStatus()
			}
		}()
	}

	// Multiple writers
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				pm.userLimiters["aws"] = connection.LimiterMap{
					"limiter1": &plugin.RateLimiter{
						Name:   "limiter1",
						Plugin: "aws",
						Status: plugin.LimiterStatusOverridden,
					},
				}
			}
		}()
	}

	wg.Wait()
}
