package pluginmanager_service

import (
	"sync"
	"testing"

	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe/v2/pkg/connection"
)

// TestPluginManager_ConcurrentRateLimiterMapAccess tests concurrent access to userLimiters map
// This test demonstrates issue #4799 - race condition when reading from userLimiters map
// in getUserDefinedLimitersForPlugin without proper mutex protection.
//
// To run this test with race detection:
//   go test -race -v -run TestPluginManager_ConcurrentRateLimiterMapAccess ./pkg/pluginmanager_service
//
// Expected behavior:
// - Before fix: Race detector reports data race on map access
// - After fix: Test passes cleanly with -race flag
func TestPluginManager_ConcurrentRateLimiterMapAccess(t *testing.T) {
	// Create a PluginManager with initialized userLimiters map
	pm := &PluginManager{
		userLimiters: make(connection.PluginLimiterMap),
		mut:          sync.RWMutex{},
	}

	// Add some initial limiters
	pm.userLimiters["aws"] = connection.LimiterMap{
		"aws-limiter-1": &plugin.RateLimiter{
			Name:   "aws-limiter-1",
			Plugin: "aws",
		},
	}
	pm.userLimiters["azure"] = connection.LimiterMap{
		"azure-limiter-1": &plugin.RateLimiter{
			Name:   "azure-limiter-1",
			Plugin: "azure",
		},
	}

	// Number of concurrent goroutines
	numGoroutines := 10
	numIterations := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Launch goroutines that READ from userLimiters via getUserDefinedLimitersForPlugin
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				// This will trigger a race condition if not protected
				_ = pm.getUserDefinedLimitersForPlugin("aws")
				_ = pm.getUserDefinedLimitersForPlugin("azure")
				_ = pm.getUserDefinedLimitersForPlugin("gcp") // doesn't exist
			}
		}(i)
	}

	// Launch goroutines that WRITE to userLimiters
	// This simulates what happens in handleUserLimiterChanges
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				// Simulate concurrent writes (like in handleUserLimiterChanges line 98-100)
				newLimiters := make(connection.PluginLimiterMap)
				newLimiters["gcp"] = connection.LimiterMap{
					"gcp-limiter-1": &plugin.RateLimiter{
						Name:   "gcp-limiter-1",
						Plugin: "gcp",
					},
				}
				// This write must be protected with mutex (just like in handleUserLimiterChanges)
				pm.mut.Lock()
				pm.userLimiters = newLimiters
				pm.mut.Unlock()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Basic sanity check
	if pm.userLimiters == nil {
		t.Error("Expected userLimiters to be non-nil")
	}
}

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
		mut: sync.RWMutex{},
	}

	// Run concurrent operations to trigger race condition
	var wg sync.WaitGroup
	iterations := 100

	// Writer goroutine - simulates handleUserLimiterChanges modifying userLimiters
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			// Simulate production code behavior - use mutex when writing
			// (see handleUserLimiterChanges lines 98-100)
			pm.mut.Lock()
			pm.userLimiters = connection.PluginLimiterMap{
				"aws": connection.LimiterMap{
					"limiter1": &plugin.RateLimiter{
						Name:   "limiter1",
						Plugin: "aws",
						Status: plugin.LimiterStatusOverridden,
					},
				},
			}
			pm.mut.Unlock()
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

// TestPluginManager_ConcurrentRateLimiterMapAccess2 tests for race condition
// when multiple goroutines access pluginLimiters and userLimiters concurrently
func TestPluginManager_ConcurrentRateLimiterMapAccess2(t *testing.T) {
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

	// Multiple writers - must use mutex protection when writing to maps
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Simulate production code behavior - use mutex when writing
				// (see handleUserLimiterChanges lines 98-100)
				pm.mut.Lock()
				pm.userLimiters["aws"] = connection.LimiterMap{
					"limiter1": &plugin.RateLimiter{
						Name:   "limiter1",
						Plugin: "aws",
						Status: plugin.LimiterStatusOverridden,
					},
				}
				pm.mut.Unlock()
			}
		}()
	}

	wg.Wait()
}
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
