package pluginmanager_service

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe/v2/pkg/connection"
)

// Test helpers for rate limiter tests

func newTestRateLimiter(pluginName, name string, source string) *plugin.RateLimiter {
	return &plugin.RateLimiter{
		Plugin: pluginName,
		Name:   name,
		Source: source,
		Status: plugin.LimiterStatusActive,
	}
}

// Test 1: ShouldFetchRateLimiterDefs

func TestPluginManager_ShouldFetchRateLimiterDefs_Nil(t *testing.T) {
	pm := newTestPluginManager(t)
	pm.pluginLimiters = nil

	should := pm.ShouldFetchRateLimiterDefs()

	assert.True(t, should, "Should fetch when pluginLimiters is nil")
}

func TestPluginManager_ShouldFetchRateLimiterDefs_NotNil(t *testing.T) {
	pm := newTestPluginManager(t)
	pm.pluginLimiters = make(connection.PluginLimiterMap)

	should := pm.ShouldFetchRateLimiterDefs()

	assert.False(t, should, "Should not fetch when pluginLimiters is initialized")
}

// Test 2: GetPluginsWithChangedLimiters

func TestPluginManager_GetPluginsWithChangedLimiters_NoChanges(t *testing.T) {
	pm := newTestPluginManager(t)

	limiter1 := newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig)
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": limiter1,
		},
	}

	newLimiters := connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": limiter1,
		},
	}

	changed := pm.getPluginsWithChangedLimiters(newLimiters)

	assert.Len(t, changed, 0, "No plugins should have changed limiters")
}

func TestPluginManager_GetPluginsWithChangedLimiters_NewPlugin(t *testing.T) {
	pm := newTestPluginManager(t)
	pm.userLimiters = connection.PluginLimiterMap{}

	newLimiters := connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig),
		},
	}

	changed := pm.getPluginsWithChangedLimiters(newLimiters)

	assert.Len(t, changed, 1, "Should detect new plugin")
	assert.Contains(t, changed, "plugin1")
}

func TestPluginManager_GetPluginsWithChangedLimiters_RemovedPlugin(t *testing.T) {
	pm := newTestPluginManager(t)
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig),
		},
	}

	newLimiters := connection.PluginLimiterMap{}

	changed := pm.getPluginsWithChangedLimiters(newLimiters)

	assert.Len(t, changed, 1, "Should detect removed plugin")
	assert.Contains(t, changed, "plugin1")
}

// Test 3: UpdateRateLimiterStatus

func TestPluginManager_UpdateRateLimiterStatus_NoOverride(t *testing.T) {
	pm := newTestPluginManager(t)

	pluginLimiter := newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourcePlugin)
	pluginLimiter.Status = plugin.LimiterStatusActive

	pm.pluginLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": pluginLimiter,
		},
	}
	pm.userLimiters = connection.PluginLimiterMap{}

	pm.updateRateLimiterStatus()

	assert.Equal(t, plugin.LimiterStatusActive, pluginLimiter.Status)
}

func TestPluginManager_UpdateRateLimiterStatus_WithOverride(t *testing.T) {
	pm := newTestPluginManager(t)

	pluginLimiter := newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourcePlugin)
	pluginLimiter.Status = plugin.LimiterStatusActive

	userLimiter := newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig)

	pm.pluginLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": pluginLimiter,
		},
	}
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": userLimiter,
		},
	}

	pm.updateRateLimiterStatus()

	assert.Equal(t, plugin.LimiterStatusOverridden, pluginLimiter.Status)
}

func TestPluginManager_UpdateRateLimiterStatus_MultiplePlugins(t *testing.T) {
	pm := newTestPluginManager(t)

	plugin1Limiter1 := newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourcePlugin)
	plugin1Limiter2 := newTestRateLimiter("plugin1", "limiter2", plugin.LimiterSourcePlugin)
	plugin2Limiter1 := newTestRateLimiter("plugin2", "limiter1", plugin.LimiterSourcePlugin)

	pm.pluginLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": plugin1Limiter1,
			"limiter2": plugin1Limiter2,
		},
		"plugin2": connection.LimiterMap{
			"limiter1": plugin2Limiter1,
		},
	}

	// Only override plugin1/limiter1
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig),
		},
	}

	pm.updateRateLimiterStatus()

	assert.Equal(t, plugin.LimiterStatusOverridden, plugin1Limiter1.Status)
	assert.Equal(t, plugin.LimiterStatusActive, plugin1Limiter2.Status)
	assert.Equal(t, plugin.LimiterStatusActive, plugin2Limiter1.Status)
}

// Test 4: GetUserDefinedLimitersForPlugin

func TestPluginManager_GetUserDefinedLimitersForPlugin_Exists(t *testing.T) {
	pm := newTestPluginManager(t)

	limiter := newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig)
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": limiter,
		},
	}

	result := pm.getUserDefinedLimitersForPlugin("plugin1")

	assert.Len(t, result, 1)
	assert.Equal(t, limiter, result["limiter1"])
}

func TestPluginManager_GetUserDefinedLimitersForPlugin_NotExists(t *testing.T) {
	pm := newTestPluginManager(t)
	pm.userLimiters = connection.PluginLimiterMap{}

	result := pm.getUserDefinedLimitersForPlugin("plugin1")

	assert.NotNil(t, result, "Should return empty map, not nil")
	assert.Len(t, result, 0)
}

// Test 5: GetUserAndPluginLimitersFromTableResult

func TestPluginManager_GetUserAndPluginLimitersFromTableResult(t *testing.T) {
	pm := newTestPluginManager(t)

	rateLimiters := []*plugin.RateLimiter{
		newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourcePlugin),
		newTestRateLimiter("plugin1", "limiter2", plugin.LimiterSourceConfig),
		newTestRateLimiter("plugin2", "limiter1", plugin.LimiterSourcePlugin),
	}

	pluginLimiters, userLimiters := pm.getUserAndPluginLimitersFromTableResult(rateLimiters)

	// Check plugin limiters
	assert.Len(t, pluginLimiters, 2)
	assert.NotNil(t, pluginLimiters["plugin1"]["limiter1"])
	assert.NotNil(t, pluginLimiters["plugin2"]["limiter1"])

	// Check user limiters
	assert.Len(t, userLimiters, 1)
	assert.NotNil(t, userLimiters["plugin1"]["limiter2"])
}

func TestPluginManager_GetUserAndPluginLimitersFromTableResult_Empty(t *testing.T) {
	pm := newTestPluginManager(t)

	rateLimiters := []*plugin.RateLimiter{}

	pluginLimiters, userLimiters := pm.getUserAndPluginLimitersFromTableResult(rateLimiters)

	assert.NotNil(t, pluginLimiters)
	assert.NotNil(t, userLimiters)
	assert.Len(t, pluginLimiters, 0)
	assert.Len(t, userLimiters, 0)
}

// Test 6: GetPluginsWithChangedLimiters Concurrent

func TestPluginManager_GetPluginsWithChangedLimiters_Concurrent(t *testing.T) {
	pm := newTestPluginManager(t)
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig),
		},
	}

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			newLimiters := connection.PluginLimiterMap{
				"plugin1": connection.LimiterMap{
					"limiter1": newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig),
				},
			}

			if idx%2 == 0 {
				// Add a new limiter
				newLimiters["plugin1"]["limiter2"] = newTestRateLimiter("plugin1", "limiter2", plugin.LimiterSourceConfig)
			}

			_ = pm.getPluginsWithChangedLimiters(newLimiters)
		}(i)
	}

	wg.Wait()
}

// Test 7: UpdateRateLimiterStatus with Multiple Limiters

func TestPluginManager_UpdateRateLimiterStatus_MultipleLimiters(t *testing.T) {
	pm := newTestPluginManager(t)

	limiter1 := newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourcePlugin)
	limiter2 := newTestRateLimiter("plugin1", "limiter2", plugin.LimiterSourcePlugin)
	limiter3 := newTestRateLimiter("plugin1", "limiter3", plugin.LimiterSourcePlugin)

	pm.pluginLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": limiter1,
			"limiter2": limiter2,
			"limiter3": limiter3,
		},
	}

	// Override only limiter2
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter2": newTestRateLimiter("plugin1", "limiter2", plugin.LimiterSourceConfig),
		},
	}

	pm.updateRateLimiterStatus()

	assert.Equal(t, plugin.LimiterStatusActive, limiter1.Status)
	assert.Equal(t, plugin.LimiterStatusOverridden, limiter2.Status)
	assert.Equal(t, plugin.LimiterStatusActive, limiter3.Status)
}

// Test 8: GetUserAndPluginLimitersFromTableResult with Duplicate Names

func TestPluginManager_GetUserAndPluginLimitersFromTableResult_DuplicateNames(t *testing.T) {
	pm := newTestPluginManager(t)

	// Same limiter name, different sources
	rateLimiters := []*plugin.RateLimiter{
		newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourcePlugin),
		newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig),
	}

	pluginLimiters, userLimiters := pm.getUserAndPluginLimitersFromTableResult(rateLimiters)

	assert.NotNil(t, pluginLimiters["plugin1"]["limiter1"])
	assert.NotNil(t, userLimiters["plugin1"]["limiter1"])
	assert.NotEqual(t, pluginLimiters["plugin1"]["limiter1"], userLimiters["plugin1"]["limiter1"])
}

// Test 9: UpdateRateLimiterStatus with Empty Maps

func TestPluginManager_UpdateRateLimiterStatus_EmptyMaps(t *testing.T) {
	pm := newTestPluginManager(t)
	pm.pluginLimiters = connection.PluginLimiterMap{}
	pm.userLimiters = connection.PluginLimiterMap{}

	// Should not panic
	pm.updateRateLimiterStatus()
}

// Test 10: GetPluginsWithChangedLimiters with Nil Comparison

func TestPluginManager_GetPluginsWithChangedLimiters_NilComparison(t *testing.T) {
	pm := newTestPluginManager(t)

	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": nil,
	}

	newLimiters := connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig),
		},
	}

	changed := pm.getPluginsWithChangedLimiters(newLimiters)

	assert.Contains(t, changed, "plugin1", "Should detect change from nil to non-nil")
}

// Test 11: ShouldFetchRateLimiterDefs Concurrent

func TestPluginManager_ShouldFetchRateLimiterDefs_Concurrent(t *testing.T) {
	pm := newTestPluginManager(t)
	pm.pluginLimiters = nil

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pm.ShouldFetchRateLimiterDefs()
		}()
	}

	wg.Wait()
}

// Test 12: GetUserDefinedLimitersForPlugin Concurrent

func TestPluginManager_GetUserDefinedLimitersForPlugin_Concurrent(t *testing.T) {
	pm := newTestPluginManager(t)
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": connection.LimiterMap{
			"limiter1": newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourceConfig),
		},
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := pm.getUserDefinedLimitersForPlugin("plugin1")
			assert.NotNil(t, result)
		}()
	}

	wg.Wait()
}

// Test 13: GetUserAndPluginLimitersFromTableResult Concurrent

func TestPluginManager_GetUserAndPluginLimitersFromTableResult_Concurrent(t *testing.T) {
	pm := newTestPluginManager(t)

	rateLimiters := []*plugin.RateLimiter{
		newTestRateLimiter("plugin1", "limiter1", plugin.LimiterSourcePlugin),
		newTestRateLimiter("plugin1", "limiter2", plugin.LimiterSourceConfig),
		newTestRateLimiter("plugin2", "limiter1", plugin.LimiterSourcePlugin),
	}

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pluginLimiters, userLimiters := pm.getUserAndPluginLimitersFromTableResult(rateLimiters)
			assert.NotNil(t, pluginLimiters)
			assert.NotNil(t, userLimiters)
		}()
	}

	wg.Wait()
}
