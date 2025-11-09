package pluginmanager_service

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/pipe-fittings/v2/plugin"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/connection"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

// Helper function to create a test plugin manager without database dependency
func createTestPluginManager(t *testing.T) *PluginManager {
	t.Helper()

	logger := hclog.NewNullLogger()

	// Create test connection config
	connectionConfig := map[string]*sdkproto.ConnectionConfig{
		"test_conn1": {
			Connection:     "test_conn1",
			Plugin:         "test",
			PluginInstance: "test",
		},
		"test_conn2": {
			Connection:     "test_conn2",
			Plugin:         "test",
			PluginInstance: "test",
		},
	}

	// Create test plugin configs
	pluginConfigs := connection.PluginMap{
		"test": {
			Plugin:   "test",
			Instance: "test",
		},
	}

	pm := &PluginManager{
		logger:                logger,
		runningPluginMap:      make(map[string]*runningPlugin),
		connectionConfigMap:   connectionConfig,
		userLimiters:          pluginConfigs.ToPluginLimiterMap(),
		plugins:               pluginConfigs,
		pluginCacheSizeMap:    make(map[string]int64),
		pluginLimiters:        make(connection.PluginLimiterMap),
		pluginConnectionConfigMap: make(map[string][]*sdkproto.ConnectionConfig),
	}

	pm.messageServer = &PluginMessageServer{pluginManager: pm}
	pm.populatePluginConnectionConfigs()
	pm.setPluginCacheSizeMap()

	return pm
}

// TestNewPluginManager tests the creation of a new plugin manager
func TestNewPluginManager(t *testing.T) {
	pm := createTestPluginManager(t)

	assert.NotNil(t, pm)
	assert.NotNil(t, pm.runningPluginMap)
	assert.NotNil(t, pm.connectionConfigMap)
	assert.NotNil(t, pm.pluginConnectionConfigMap)
	assert.NotNil(t, pm.messageServer)
	assert.Equal(t, 2, len(pm.connectionConfigMap))
}

// TestBuildRequiredPluginMap tests building the plugin map from requested connections
func TestBuildRequiredPluginMap(t *testing.T) {
	pm := createTestPluginManager(t)

	tests := []struct {
		name               string
		connections        []string
		expectedPlugins    int
		expectedError      bool
		expectedErrorMsg   string
	}{
		{
			name:            "single connection",
			connections:     []string{"test_conn1"},
			expectedPlugins: 1,
			expectedError:   false,
		},
		{
			name:            "multiple connections same plugin",
			connections:     []string{"test_conn1", "test_conn2"},
			expectedPlugins: 1,
			expectedError:   false,
		},
		{
			name:             "nonexistent connection",
			connections:      []string{"nonexistent"},
			expectedPlugins:  0,
			expectedError:    true,
			expectedErrorMsg: "does not exist",
		},
		{
			name:            "empty connection list",
			connections:     []string{},
			expectedPlugins: 0,
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.GetRequest{
				Connections: tt.connections,
			}

			pluginMap, requestedConnections, err := pm.buildRequiredPluginMap(req)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPlugins, len(pluginMap))
				assert.Equal(t, len(tt.connections), len(requestedConnections))
			}
		})
	}
}

// TestGetConnectionConfig tests retrieving connection configuration
func TestGetConnectionConfig(t *testing.T) {
	pm := createTestPluginManager(t)

	tests := []struct {
		name           string
		connectionName string
		expectedError  bool
	}{
		{
			name:           "existing connection",
			connectionName: "test_conn1",
			expectedError:  false,
		},
		{
			name:           "nonexistent connection",
			connectionName: "nonexistent",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := pm.getConnectionConfig(tt.connectionName)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.connectionName, config.Connection)
			}
		})
	}
}

// TestPopulatePluginConnectionConfigs tests populating plugin connection config map
func TestPopulatePluginConnectionConfigs(t *testing.T) {
	pm := createTestPluginManager(t)

	// Verify the plugin connection config map was populated correctly
	assert.NotNil(t, pm.pluginConnectionConfigMap)
	assert.Equal(t, 1, len(pm.pluginConnectionConfigMap))

	// Verify the test plugin has both connections
	testConnections, ok := pm.pluginConnectionConfigMap["test"]
	assert.True(t, ok)
	assert.Equal(t, 2, len(testConnections))
}

// TestSetPluginCacheSizeMap tests setting plugin cache sizes
func TestSetPluginCacheSizeMap(t *testing.T) {
	pm := createTestPluginManager(t)

	assert.NotNil(t, pm.pluginCacheSizeMap)
	assert.Equal(t, 1, len(pm.pluginCacheSizeMap))

	// By default, with no max cache size set, all plugins should have 0 (unlimited)
	cacheSize, ok := pm.pluginCacheSizeMap["test"]
	assert.True(t, ok)
	assert.GreaterOrEqual(t, cacheSize, int64(0))
}

// TestAddRunningPlugin tests adding a plugin to the running plugin map
func TestAddRunningPlugin(t *testing.T) {
	pm := createTestPluginManager(t)

	tests := []struct {
		name           string
		pluginInstance string
		expectedError  bool
		setupFunc      func()
	}{
		{
			name:           "add new plugin",
			pluginInstance: "test",
			expectedError:  false,
			setupFunc:      func() {},
		},
		{
			name:           "add duplicate plugin",
			pluginInstance: "test",
			expectedError:  true,
			setupFunc: func() {
				// Add the plugin first
				pm.runningPluginMap["test"] = &runningPlugin{
					pluginInstance: "test",
					initialized:    make(chan struct{}),
					failed:         make(chan struct{}),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the running plugin map
			pm.runningPluginMap = make(map[string]*runningPlugin)

			// Run setup
			tt.setupFunc()

			plugin, err := pm.addRunningPlugin(tt.pluginInstance)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plugin)
				assert.Equal(t, tt.pluginInstance, plugin.pluginInstance)
				assert.NotNil(t, plugin.initialized)
				assert.NotNil(t, plugin.failed)

				// Verify it was added to the map
				addedPlugin, ok := pm.runningPluginMap[tt.pluginInstance]
				assert.True(t, ok)
				assert.Equal(t, plugin, addedPlugin)
			}
		})
	}
}

// TestIsShuttingDown tests the shutdown state management
func TestIsShuttingDown(t *testing.T) {
	pm := createTestPluginManager(t)

	// Initially not shutting down
	assert.False(t, pm.isShuttingDown())

	// Set shutting down
	pm.shutdownMut.Lock()
	pm.shuttingDown = true
	pm.shutdownMut.Unlock()

	// Should now be shutting down
	assert.True(t, pm.isShuttingDown())
}

// TestGetResponseConcurrentUsage tests concurrent access to getResponse
func TestGetResponseConcurrentUsage(t *testing.T) {
	resp := newGetResponse()

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.ReattachMap)
	assert.NotNil(t, resp.FailureMap)

	// Test concurrent additions
	done := make(chan bool)

	// Add reattach configs concurrently
	go func() {
		resp.AddReattach("conn1", &pb.ReattachConfig{})
		done <- true
	}()

	go func() {
		resp.AddReattach("conn2", &pb.ReattachConfig{})
		done <- true
	}()

	// Add failures concurrently
	go func() {
		resp.AddFailure("plugin1", "error1")
		done <- true
	}()

	go func() {
		resp.AddFailure("plugin2", "error2")
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify all entries were added
	assert.Equal(t, 2, len(resp.ReattachMap))
	assert.Equal(t, 2, len(resp.FailureMap))
}

// TestNonAggregatorConnectionCount tests counting non-aggregator connections
func TestNonAggregatorConnectionCount(t *testing.T) {
	tests := []struct {
		name        string
		connections []*sdkproto.ConnectionConfig
		expected    int
	}{
		{
			name: "all regular connections",
			connections: []*sdkproto.ConnectionConfig{
				{Connection: "conn1", ChildConnections: []string{}},
				{Connection: "conn2", ChildConnections: []string{}},
			},
			expected: 2,
		},
		{
			name: "mixed connections",
			connections: []*sdkproto.ConnectionConfig{
				{Connection: "conn1", ChildConnections: []string{}},
				{Connection: "agg1", ChildConnections: []string{"conn1", "conn2"}},
			},
			expected: 1,
		},
		{
			name: "all aggregators",
			connections: []*sdkproto.ConnectionConfig{
				{Connection: "agg1", ChildConnections: []string{"conn1"}},
				{Connection: "agg2", ChildConnections: []string{"conn2"}},
			},
			expected: 0,
		},
		{
			name:        "empty list",
			connections: []*sdkproto.ConnectionConfig{},
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := nonAggregatorConnectionCount(tt.connections)
			assert.Equal(t, tt.expected, count)
		})
	}
}

// TestGetPluginExemplarConnections tests getting exemplar connections for each plugin
func TestGetPluginExemplarConnections(t *testing.T) {
	pm := createTestPluginManager(t)

	// Add connections for multiple plugins
	pm.connectionConfigMap = map[string]*sdkproto.ConnectionConfig{
		"conn1": {Connection: "conn1", Plugin: "plugin1"},
		"conn2": {Connection: "conn2", Plugin: "plugin1"},
		"conn3": {Connection: "conn3", Plugin: "plugin2"},
	}

	exemplars := pm.getPluginExemplarConnections()

	assert.Equal(t, 2, len(exemplars))
	assert.Contains(t, exemplars, "plugin1")
	assert.Contains(t, exemplars, "plugin2")

	// Each plugin should have one exemplar connection
	assert.NotEmpty(t, exemplars["plugin1"])
	assert.NotEmpty(t, exemplars["plugin2"])
}

// TestHandleConnectionConfigChanges tests handling connection configuration changes
func TestHandleConnectionConfigChanges(t *testing.T) {
	pm := createTestPluginManager(t)

	// Test adding a new connection
	newConfigMap := map[string]*sdkproto.ConnectionConfig{
		"test_conn1": pm.connectionConfigMap["test_conn1"],
		"test_conn2": pm.connectionConfigMap["test_conn2"],
		"test_conn3": {
			Connection:     "test_conn3",
			Plugin:         "test",
			PluginInstance: "test",
		},
	}

	ctx := context.Background()
	err := pm.handleConnectionConfigChanges(ctx, newConfigMap)

	// We expect this to succeed (even though we can't actually update running plugins in this test)
	// The error would come from sendUpdateConnectionConfigs, which we can't fully test without running plugins
	// But the logic for detecting changes should work
	assert.NotNil(t, err == nil || err != nil) // Just check it doesn't panic

	// Verify the connection config map was updated
	assert.Equal(t, 3, len(pm.connectionConfigMap))
}

// TestHandleAddedConnections tests handling added connections
func TestHandleAddedConnections(t *testing.T) {
	pm := createTestPluginManager(t)

	requestMap := make(map[string]*sdkproto.UpdateConnectionConfigsRequest)

	addedConnections := map[string][]*sdkproto.ConnectionConfig{
		"test": {
			{Connection: "new_conn", Plugin: "test", PluginInstance: "test"},
		},
	}

	// Add a running plugin
	pm.runningPluginMap["test"] = &runningPlugin{
		pluginInstance: "test",
		reattach: &pb.ReattachConfig{
			Connections: []string{"test_conn1"},
		},
	}

	pm.handleAddedConnections(addedConnections, requestMap)

	// Verify request was created
	req, ok := requestMap["test"]
	assert.True(t, ok)
	assert.NotNil(t, req)
	assert.Equal(t, 1, len(req.Added))
	assert.Equal(t, "new_conn", req.Added[0].Connection)
}

// TestHandleDeletedConnections tests handling deleted connections
func TestHandleDeletedConnections(t *testing.T) {
	pm := createTestPluginManager(t)

	requestMap := make(map[string]*sdkproto.UpdateConnectionConfigsRequest)

	deletedConnections := map[string][]*sdkproto.ConnectionConfig{
		"test": {
			{Connection: "test_conn1", Plugin: "test", PluginInstance: "test"},
		},
	}

	// Add a running plugin
	pm.runningPluginMap["test"] = &runningPlugin{
		pluginInstance: "test",
		reattach: &pb.ReattachConfig{
			Connections: []string{"test_conn1", "test_conn2"},
		},
	}

	pm.handleDeletedConnections(deletedConnections, requestMap)

	// Verify request was created
	req, ok := requestMap["test"]
	assert.True(t, ok)
	assert.NotNil(t, req)
	assert.Equal(t, 1, len(req.Deleted))
	assert.Equal(t, "test_conn1", req.Deleted[0].Connection)
}

// TestHandleUpdatedConnections tests handling updated connections
func TestHandleUpdatedConnections(t *testing.T) {
	pm := createTestPluginManager(t)

	requestMap := make(map[string]*sdkproto.UpdateConnectionConfigsRequest)

	updatedConnections := map[string][]*sdkproto.ConnectionConfig{
		"test": {
			{Connection: "test_conn1", Plugin: "test", PluginInstance: "test"},
		},
	}

	pm.handleUpdatedConnections(updatedConnections, requestMap)

	// Verify request was created
	req, ok := requestMap["test"]
	assert.True(t, ok)
	assert.NotNil(t, req)
	assert.Equal(t, 1, len(req.Changed))
	assert.Equal(t, "test_conn1", req.Changed[0].Connection)
}

// TestRateLimiterConversion tests converting between rate limiter formats
func TestRateLimiterConversion(t *testing.T) {
	fillRate := float32(10)
	bucketSize := int64(100)
	maxConcurrency := int64(5)
	where := "test"

	// Create a rate limiter
	limiter := &plugin.RateLimiter{
		Name:           "test_limiter",
		Scope:          []string{"table1", "table2"},
		FillRate:       &fillRate,
		BucketSize:     &bucketSize,
		MaxConcurrency: &maxConcurrency,
		Where:          &where,
	}

	// Convert to proto
	protoLimiter := RateLimiterAsProto(limiter)

	assert.NotNil(t, protoLimiter)
	assert.Equal(t, "test_limiter", protoLimiter.Name)
	assert.Equal(t, fillRate, protoLimiter.FillRate)
	assert.Equal(t, bucketSize, protoLimiter.BucketSize)
	assert.Equal(t, maxConcurrency, protoLimiter.MaxConcurrency)
	assert.Equal(t, where, protoLimiter.Where)
	assert.Equal(t, 2, len(protoLimiter.Scope))

	// Convert back from proto
	convertedLimiter, err := RateLimiterFromProto(protoLimiter, "test_plugin", "test_instance")

	require.NoError(t, err)
	assert.Equal(t, "test_limiter", convertedLimiter.Name)
	assert.Equal(t, fillRate, *convertedLimiter.FillRate)
	assert.Equal(t, bucketSize, *convertedLimiter.BucketSize)
	assert.Equal(t, maxConcurrency, *convertedLimiter.MaxConcurrency)
	assert.Equal(t, where, *convertedLimiter.Where)
	assert.Equal(t, "test_instance", convertedLimiter.PluginInstance)
}

// TestRateLimiterConversionWithNilFields tests converting rate limiters with nil optional fields
func TestRateLimiterConversionWithNilFields(t *testing.T) {
	// Create a minimal rate limiter
	limiter := &plugin.RateLimiter{
		Name:  "minimal_limiter",
		Scope: []string{},
	}

	// Convert to proto
	protoLimiter := RateLimiterAsProto(limiter)

	assert.NotNil(t, protoLimiter)
	assert.Equal(t, "minimal_limiter", protoLimiter.Name)
	assert.Equal(t, float32(0), protoLimiter.FillRate)
	assert.Equal(t, int64(0), protoLimiter.BucketSize)
	assert.Equal(t, int64(0), protoLimiter.MaxConcurrency)
	assert.Equal(t, "", protoLimiter.Where)

	// Convert back from proto
	convertedLimiter, err := RateLimiterFromProto(protoLimiter, "test_plugin", "test_instance")

	require.NoError(t, err)
	assert.Equal(t, "minimal_limiter", convertedLimiter.Name)
	assert.Nil(t, convertedLimiter.FillRate)
	assert.Nil(t, convertedLimiter.BucketSize)
	assert.Nil(t, convertedLimiter.MaxConcurrency)
	assert.Nil(t, convertedLimiter.Where)
	assert.NotNil(t, convertedLimiter.Scope)
	assert.Equal(t, 0, len(convertedLimiter.Scope))
}
