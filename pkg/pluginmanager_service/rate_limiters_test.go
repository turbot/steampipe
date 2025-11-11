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
				// Simulate concurrent writes (like in handleUserLimiterChanges line 96)
				newLimiters := make(connection.PluginLimiterMap)
				newLimiters["gcp"] = connection.LimiterMap{
					"gcp-limiter-1": &plugin.RateLimiter{
						Name:   "gcp-limiter-1",
						Plugin: "gcp",
					},
				}
				// This write will race with the reads in getUserDefinedLimitersForPlugin
				pm.userLimiters = newLimiters
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
