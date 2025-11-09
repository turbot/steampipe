package pluginmanager_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe/v2/pkg/connection"
)

// TestShouldFetchRateLimiterDefs tests checking if rate limiter defs should be fetched
func TestShouldFetchRateLimiterDefs(t *testing.T) {
	pm := createTestPluginManager(t)

	tests := []struct {
		name     string
		setup    func()
		expected bool
	}{
		{
			name: "plugin limiters not initialized",
			setup: func() {
				pm.pluginLimiters = nil
			},
			expected: true,
		},
		{
			name: "plugin limiters initialized",
			setup: func() {
				pm.pluginLimiters = make(connection.PluginLimiterMap)
			},
			expected: false,
		},
		{
			name: "plugin limiters with data",
			setup: func() {
				pm.pluginLimiters = connection.PluginLimiterMap{
					"test": make(connection.LimiterMap),
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := pm.ShouldFetchRateLimiterDefs()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetUserDefinedLimitersForPlugin tests getting user-defined limiters for a plugin
func TestGetUserDefinedLimitersForPlugin(t *testing.T) {
	pm := createTestPluginManager(t)

	fillRate := float32(10)
	bucketSize := int64(100)

	// Set up user limiters
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": {
			"limiter1": &plugin.RateLimiter{
				Name:       "limiter1",
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
		},
	}

	tests := []struct {
		name           string
		pluginName     string
		expectedCount  int
		expectedExists bool
	}{
		{
			name:           "plugin with limiters",
			pluginName:     "plugin1",
			expectedCount:  1,
			expectedExists: true,
		},
		{
			name:           "plugin without limiters",
			pluginName:     "plugin2",
			expectedCount:  0,
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiters := pm.getUserDefinedLimitersForPlugin(tt.pluginName)

			assert.NotNil(t, limiters)
			assert.Equal(t, tt.expectedCount, len(limiters))

			if tt.expectedExists {
				_, exists := limiters["limiter1"]
				assert.True(t, exists)
			}
		})
	}
}

// TestUpdateRateLimiterStatus tests updating rate limiter status
func TestUpdateRateLimiterStatus(t *testing.T) {
	pm := createTestPluginManager(t)

	fillRate := float32(10)
	bucketSize := int64(100)

	// Set up plugin limiters
	pm.pluginLimiters = connection.PluginLimiterMap{
		"plugin1": {
			"limiter1": &plugin.RateLimiter{
				Name:       "limiter1",
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
				Status:     plugin.LimiterStatusActive,
			},
			"limiter2": &plugin.RateLimiter{
				Name:       "limiter2",
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
				Status:     plugin.LimiterStatusActive,
			},
		},
	}

	// Set up user limiters (overriding limiter1)
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": {
			"limiter1": &plugin.RateLimiter{
				Name:       "limiter1",
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
		},
	}

	// Update status
	pm.updateRateLimiterStatus()

	// limiter1 should be marked as overridden
	assert.Equal(t, plugin.LimiterStatusOverridden, pm.pluginLimiters["plugin1"]["limiter1"].Status)

	// limiter2 should remain active
	assert.Equal(t, plugin.LimiterStatusActive, pm.pluginLimiters["plugin1"]["limiter2"].Status)
}

// TestGetPluginsWithChangedLimiters tests detecting plugins with changed limiters
func TestGetPluginsWithChangedLimiters(t *testing.T) {
	pm := createTestPluginManager(t)

	fillRate1 := float32(10)
	bucketSize1 := int64(100)
	fillRate2 := float32(20)
	bucketSize2 := int64(200)

	// Set up current user limiters
	pm.userLimiters = connection.PluginLimiterMap{
		"plugin1": {
			"limiter1": &plugin.RateLimiter{
				Name:       "limiter1",
				FillRate:   &fillRate1,
				BucketSize: &bucketSize1,
			},
		},
		"plugin2": {
			"limiter1": &plugin.RateLimiter{
				Name:       "limiter1",
				FillRate:   &fillRate1,
				BucketSize: &bucketSize1,
			},
		},
	}

	tests := []struct {
		name           string
		newLimiters    connection.PluginLimiterMap
		expectedCount  int
		expectedPlugin string
	}{
		{
			name: "no changes",
			newLimiters: connection.PluginLimiterMap{
				"plugin1": {
					"limiter1": &plugin.RateLimiter{
						Name:       "limiter1",
						FillRate:   &fillRate1,
						BucketSize: &bucketSize1,
					},
				},
				"plugin2": {
					"limiter1": &plugin.RateLimiter{
						Name:       "limiter1",
						FillRate:   &fillRate1,
						BucketSize: &bucketSize1,
					},
				},
			},
			expectedCount: 0,
		},
		{
			name: "one plugin changed",
			newLimiters: connection.PluginLimiterMap{
				"plugin1": {
					"limiter1": &plugin.RateLimiter{
						Name:       "limiter1",
						FillRate:   &fillRate2,
						BucketSize: &bucketSize2,
					},
				},
				"plugin2": {
					"limiter1": &plugin.RateLimiter{
						Name:       "limiter1",
						FillRate:   &fillRate1,
						BucketSize: &bucketSize1,
					},
				},
			},
			expectedCount:  1,
			expectedPlugin: "plugin1",
		},
		{
			name: "new plugin added",
			newLimiters: connection.PluginLimiterMap{
				"plugin1": {
					"limiter1": &plugin.RateLimiter{
						Name:       "limiter1",
						FillRate:   &fillRate1,
						BucketSize: &bucketSize1,
					},
				},
				"plugin2": {
					"limiter1": &plugin.RateLimiter{
						Name:       "limiter1",
						FillRate:   &fillRate1,
						BucketSize: &bucketSize1,
					},
				},
				"plugin3": {
					"limiter1": &plugin.RateLimiter{
						Name:       "limiter1",
						FillRate:   &fillRate1,
						BucketSize: &bucketSize1,
					},
				},
			},
			expectedCount:  1,
			expectedPlugin: "plugin3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := pm.getPluginsWithChangedLimiters(tt.newLimiters)

			assert.Equal(t, tt.expectedCount, len(changed))

			if tt.expectedPlugin != "" {
				_, exists := changed[tt.expectedPlugin]
				assert.True(t, exists, "expected plugin %s to be in changed set", tt.expectedPlugin)
			}
		})
	}
}

// TestGetUserAndPluginLimitersFromTableResult tests splitting limiters by source
func TestGetUserAndPluginLimitersFromTableResult(t *testing.T) {
	pm := createTestPluginManager(t)

	fillRate := float32(10)
	bucketSize := int64(100)

	rateLimiters := []*plugin.RateLimiter{
		{
			Name:       "plugin_limiter1",
			Plugin:     "plugin1",
			FillRate:   &fillRate,
			BucketSize: &bucketSize,
			Source:     plugin.LimiterSourcePlugin,
		},
		{
			Name:       "plugin_limiter2",
			Plugin:     "plugin1",
			FillRate:   &fillRate,
			BucketSize: &bucketSize,
			Source:     plugin.LimiterSourcePlugin,
		},
		{
			Name:       "user_limiter1",
			Plugin:     "plugin1",
			FillRate:   &fillRate,
			BucketSize: &bucketSize,
			Source:     plugin.LimiterSourceConfig,
		},
		{
			Name:       "plugin_limiter1",
			Plugin:     "plugin2",
			FillRate:   &fillRate,
			BucketSize: &bucketSize,
			Source:     plugin.LimiterSourcePlugin,
		},
	}

	pluginLimiters, userLimiters := pm.getUserAndPluginLimitersFromTableResult(rateLimiters)

	// Verify plugin limiters
	assert.Equal(t, 2, len(pluginLimiters))
	assert.NotNil(t, pluginLimiters["plugin1"])
	assert.Equal(t, 2, len(pluginLimiters["plugin1"]))
	assert.NotNil(t, pluginLimiters["plugin2"])
	assert.Equal(t, 1, len(pluginLimiters["plugin2"]))

	// Verify user limiters
	assert.Equal(t, 1, len(userLimiters))
	assert.NotNil(t, userLimiters["plugin1"])
	assert.Equal(t, 1, len(userLimiters["plugin1"]))
	assert.NotNil(t, userLimiters["plugin1"]["user_limiter1"])
}

// TestHandlePluginLimiterChanges tests handling plugin limiter changes
func TestHandlePluginLimiterChanges(t *testing.T) {
	t.Skip("Skipping test that requires database connection")

	// This test requires a full database connection pool which we don't have in unit tests
	// The logic is tested indirectly through other tests that check the plugin limiter maps
}

// TestRateLimiterEquality tests comparing rate limiters for equality
func TestRateLimiterEquality(t *testing.T) {
	fillRate := float32(10)
	bucketSize := int64(100)
	maxConcurrency := int64(5)

	tests := []struct {
		name     string
		limiter1 *plugin.RateLimiter
		limiter2 *plugin.RateLimiter
		expected bool
	}{
		{
			name: "identical limiters",
			limiter1: &plugin.RateLimiter{
				Name:       "test",
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
			limiter2: &plugin.RateLimiter{
				Name:       "test",
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
			expected: true,
		},
		{
			name: "different fill rates",
			limiter1: &plugin.RateLimiter{
				Name:       "test",
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
			limiter2: &plugin.RateLimiter{
				Name:       "test",
				FillRate:   func() *float32 { r := float32(20); return &r }(),
				BucketSize: &bucketSize,
			},
			expected: false,
		},
		{
			name: "different max concurrency",
			limiter1: &plugin.RateLimiter{
				Name:           "test",
				MaxConcurrency: &maxConcurrency,
			},
			limiter2: &plugin.RateLimiter{
				Name:           "test",
				MaxConcurrency: func() *int64 { c := int64(10); return &c }(),
			},
			expected: false,
		},
		{
			name: "one nil max concurrency",
			limiter1: &plugin.RateLimiter{
				Name:           "test",
				MaxConcurrency: &maxConcurrency,
			},
			limiter2: &plugin.RateLimiter{
				Name:           "test",
				MaxConcurrency: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.limiter1.Equals(tt.limiter2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestLimiterMapEquality tests comparing limiter maps for equality
func TestLimiterMapEquality(t *testing.T) {
	fillRate := float32(10)
	bucketSize := int64(100)

	tests := []struct {
		name     string
		map1     connection.LimiterMap
		map2     connection.LimiterMap
		expected bool
	}{
		{
			name: "identical maps",
			map1: connection.LimiterMap{
				"limiter1": &plugin.RateLimiter{
					Name:       "limiter1",
					FillRate:   &fillRate,
					BucketSize: &bucketSize,
				},
			},
			map2: connection.LimiterMap{
				"limiter1": &plugin.RateLimiter{
					Name:       "limiter1",
					FillRate:   &fillRate,
					BucketSize: &bucketSize,
				},
			},
			expected: true,
		},
		{
			name: "different sizes",
			map1: connection.LimiterMap{
				"limiter1": &plugin.RateLimiter{
					Name:       "limiter1",
					FillRate:   &fillRate,
					BucketSize: &bucketSize,
				},
			},
			map2: connection.LimiterMap{
				"limiter1": &plugin.RateLimiter{
					Name:       "limiter1",
					FillRate:   &fillRate,
					BucketSize: &bucketSize,
				},
				"limiter2": &plugin.RateLimiter{
					Name:       "limiter2",
					FillRate:   &fillRate,
					BucketSize: &bucketSize,
				},
			},
			expected: false,
		},
		{
			name: "different keys",
			map1: connection.LimiterMap{
				"limiter1": &plugin.RateLimiter{
					Name:       "limiter1",
					FillRate:   &fillRate,
					BucketSize: &bucketSize,
				},
			},
			map2: connection.LimiterMap{
				"limiter2": &plugin.RateLimiter{
					Name:       "limiter2",
					FillRate:   &fillRate,
					BucketSize: &bucketSize,
				},
			},
			expected: false,
		},
		{
			name: "both nil",
			map1: nil,
			map2: nil,
			expected: true,
		},
		{
			name: "one nil",
			map1: connection.LimiterMap{
				"limiter1": &plugin.RateLimiter{
					Name:       "limiter1",
					FillRate:   &fillRate,
					BucketSize: &bucketSize,
				},
			},
			map2:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.map1.Equals(tt.map2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRateLimiterWithDifferentScopes tests rate limiters with different scopes
func TestRateLimiterWithDifferentScopes(t *testing.T) {
	fillRate := float32(10)
	bucketSize := int64(100)

	tests := []struct {
		name     string
		limiter1 *plugin.RateLimiter
		limiter2 *plugin.RateLimiter
		expected bool
	}{
		{
			name: "same scopes",
			limiter1: &plugin.RateLimiter{
				Name:       "test",
				Scope:      []string{"table1", "table2"},
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
			limiter2: &plugin.RateLimiter{
				Name:       "test",
				Scope:      []string{"table1", "table2"},
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
			expected: true,
		},
		{
			name: "different scopes",
			limiter1: &plugin.RateLimiter{
				Name:       "test",
				Scope:      []string{"table1"},
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
			limiter2: &plugin.RateLimiter{
				Name:       "test",
				Scope:      []string{"table2"},
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
			expected: false,
		},
		{
			name: "different scope lengths",
			limiter1: &plugin.RateLimiter{
				Name:       "test",
				Scope:      []string{"table1"},
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
			limiter2: &plugin.RateLimiter{
				Name:       "test",
				Scope:      []string{"table1", "table2"},
				FillRate:   &fillRate,
				BucketSize: &bucketSize,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.limiter1.Equals(tt.limiter2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRateLimiterProtoConversionWithScope tests proto conversion with scope
func TestRateLimiterProtoConversionWithScope(t *testing.T) {
	fillRate := float32(10)
	bucketSize := int64(100)

	limiter := &plugin.RateLimiter{
		Name:       "test_limiter",
		Scope:      []string{"table1", "table2", "table3"},
		FillRate:   &fillRate,
		BucketSize: &bucketSize,
	}

	// Convert to proto
	protoLimiter := RateLimiterAsProto(limiter)

	require.NotNil(t, protoLimiter)
	assert.Equal(t, 3, len(protoLimiter.Scope))
	assert.Contains(t, protoLimiter.Scope, "table1")
	assert.Contains(t, protoLimiter.Scope, "table2")
	assert.Contains(t, protoLimiter.Scope, "table3")

	// Convert back
	convertedLimiter, err := RateLimiterFromProto(protoLimiter, "test_plugin", "test_instance")

	require.NoError(t, err)
	assert.Equal(t, 3, len(convertedLimiter.Scope))
	assert.Contains(t, convertedLimiter.Scope, "table1")
	assert.Contains(t, convertedLimiter.Scope, "table2")
	assert.Contains(t, convertedLimiter.Scope, "table3")
}
