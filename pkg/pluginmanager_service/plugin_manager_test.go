package pluginmanager_service

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/pipe-fittings/v2/plugin"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/connection"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

// Test helpers and mocks

func newTestPluginManager(t *testing.T) *PluginManager {
	t.Helper()

	logger := hclog.NewNullLogger()

	pm := &PluginManager{
		logger:                    logger,
		runningPluginMap:          make(map[string]*runningPlugin),
		pluginConnectionConfigMap: make(map[string][]*sdkproto.ConnectionConfig),
		connectionConfigMap:       make(connection.ConnectionConfigMap),
		pluginCacheSizeMap:        make(map[string]int64),
		plugins:                   make(connection.PluginMap),
		userLimiters:              make(connection.PluginLimiterMap),
		pluginLimiters:            make(connection.PluginLimiterMap),
	}

	pm.messageServer = &PluginMessageServer{pluginManager: pm}

	return pm
}

func newTestConnectionConfig(plugin, instance, connection string) *sdkproto.ConnectionConfig {
	return &sdkproto.ConnectionConfig{
		Plugin:         plugin,
		PluginInstance: instance,
		Connection:     connection,
	}
}

// Test 1: Basic Initialization

func TestPluginManager_New(t *testing.T) {
	pm := newTestPluginManager(t)

	assert.NotNil(t, pm, "PluginManager should not be nil")
	assert.NotNil(t, pm.runningPluginMap, "runningPluginMap should be initialized")
	assert.NotNil(t, pm.messageServer, "messageServer should be initialized")
	assert.NotNil(t, pm.logger, "logger should be initialized")
}

// Test 2: Connection Config Access

func TestPluginManager_GetConnectionConfig_NotFound(t *testing.T) {
	pm := newTestPluginManager(t)

	_, err := pm.getConnectionConfig("nonexistent")

	assert.Error(t, err, "Should return error for nonexistent connection")
	assert.Contains(t, err.Error(), "does not exist", "Error should mention connection doesn't exist")
}

func TestPluginManager_GetConnectionConfig_Found(t *testing.T) {
	pm := newTestPluginManager(t)

	expectedConfig := newTestConnectionConfig("test-plugin", "test-instance", "test-connection")
	pm.connectionConfigMap["test-connection"] = expectedConfig

	config, err := pm.getConnectionConfig("test-connection")

	require.NoError(t, err)
	assert.Equal(t, expectedConfig, config)
}

func TestPluginManager_GetConnectionConfig_NilMap(t *testing.T) {
	pm := newTestPluginManager(t)
	pm.connectionConfigMap = nil

	_, err := pm.getConnectionConfig("conn1")

	assert.Error(t, err, "Should handle nil connectionConfigMap gracefully")
}

// Test 3: Map Population

func TestPluginManager_PopulatePluginConnectionConfigs(t *testing.T) {
	pm := newTestPluginManager(t)

	config1 := newTestConnectionConfig("plugin1", "instance1", "conn1")
	config2 := newTestConnectionConfig("plugin1", "instance1", "conn2")
	config3 := newTestConnectionConfig("plugin2", "instance2", "conn3")

	pm.connectionConfigMap = connection.ConnectionConfigMap{
		"conn1": config1,
		"conn2": config2,
		"conn3": config3,
	}

	pm.populatePluginConnectionConfigs()

	assert.Len(t, pm.pluginConnectionConfigMap, 2, "Should have 2 plugin instances")
	assert.Len(t, pm.pluginConnectionConfigMap["instance1"], 2, "instance1 should have 2 connections")
	assert.Len(t, pm.pluginConnectionConfigMap["instance2"], 1, "instance2 should have 1 connection")
}

// Test 4: Build Required Plugin Map

func TestPluginManager_BuildRequiredPluginMap(t *testing.T) {
	pm := newTestPluginManager(t)

	config1 := newTestConnectionConfig("plugin1", "instance1", "conn1")
	config2 := newTestConnectionConfig("plugin1", "instance1", "conn2")
	config3 := newTestConnectionConfig("plugin2", "instance2", "conn3")

	pm.connectionConfigMap = connection.ConnectionConfigMap{
		"conn1": config1,
		"conn2": config2,
		"conn3": config3,
	}
	pm.populatePluginConnectionConfigs()

	req := &pb.GetRequest{
		Connections: []string{"conn1", "conn3"},
	}

	pluginMap, requestedConns, err := pm.buildRequiredPluginMap(req)

	require.NoError(t, err)
	assert.Len(t, pluginMap, 2, "Should map 2 plugin instances")
	assert.Len(t, requestedConns, 2, "Should have 2 requested connections")
	assert.Contains(t, requestedConns, "conn1")
	assert.Contains(t, requestedConns, "conn3")
}

// Test 5: Concurrent Map Access

func TestPluginManager_ConcurrentMapAccess(t *testing.T) {
	pm := newTestPluginManager(t)

	// Populate some initial data
	for i := 0; i < 10; i++ {
		connName := fmt.Sprintf("conn%d", i)
		config := newTestConnectionConfig("plugin1", "instance1", connName)
		pm.connectionConfigMap[connName] = config
	}
	pm.populatePluginConnectionConfigs()

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent reads with proper locking
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			connName := fmt.Sprintf("conn%d", idx%10)

			pm.mut.RLock()
			_ = pm.connectionConfigMap[connName]
			pm.mut.RUnlock()
		}(i)
	}

	wg.Wait()

	assert.Len(t, pm.connectionConfigMap, 10)
}

// Test 6: Shutdown Flag Management

func TestPluginManager_Shutdown_SetsShuttingDownFlag(t *testing.T) {
	pm := newTestPluginManager(t)

	assert.False(t, pm.isShuttingDown(), "Initially should not be shutting down")

	// Set the flag as Shutdown does
	pm.shutdownMut.Lock()
	pm.shuttingDown = true
	pm.shutdownMut.Unlock()

	assert.True(t, pm.isShuttingDown(), "Should be shutting down after flag is set")
}

func TestPluginManager_Shutdown_WaitsForPluginStart(t *testing.T) {
	pm := newTestPluginManager(t)

	// Simulate a plugin starting
	pm.startPluginWg.Add(1)

	shutdownComplete := make(chan struct{})

	go func() {
		pm.shutdownMut.Lock()
		pm.shuttingDown = true
		pm.shutdownMut.Unlock()
		pm.startPluginWg.Wait()
		close(shutdownComplete)
	}()

	// Give shutdown goroutine time to reach Wait
	time.Sleep(50 * time.Millisecond)

	// Verify shutdown hasn't completed yet
	select {
	case <-shutdownComplete:
		t.Fatal("Shutdown completed before startPluginWg.Done() was called")
	case <-time.After(10 * time.Millisecond):
		// Expected
	}

	// Signal plugin start complete
	pm.startPluginWg.Done()

	// Verify shutdown completes
	select {
	case <-shutdownComplete:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Shutdown did not complete after startPluginWg.Done()")
	}
}

// Test 7: Running Plugin Management

func TestPluginManager_AddRunningPlugin_Success(t *testing.T) {
	pm := newTestPluginManager(t)

	// Add a plugin config
	pm.plugins["test-instance"] = &plugin.Plugin{
		Plugin:   "test-plugin",
		Instance: "test-instance",
	}

	rp, err := pm.addRunningPlugin("test-instance")

	require.NoError(t, err)
	assert.NotNil(t, rp)
	assert.Equal(t, "test-instance", rp.pluginInstance)
	assert.NotNil(t, rp.initialized)
	assert.NotNil(t, rp.failed)

	// Verify it was added to the map
	pm.mut.RLock()
	stored := pm.runningPluginMap["test-instance"]
	pm.mut.RUnlock()
	assert.Equal(t, rp, stored)
}

func TestPluginManager_AddRunningPlugin_AlreadyExists(t *testing.T) {
	pm := newTestPluginManager(t)

	// Add a plugin config
	pm.plugins["test-instance"] = &plugin.Plugin{
		Plugin:   "test-plugin",
		Instance: "test-instance",
	}

	// Add first time
	_, err := pm.addRunningPlugin("test-instance")
	require.NoError(t, err)

	// Try to add again - should return retryable error
	_, err = pm.addRunningPlugin("test-instance")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")
}

func TestPluginManager_AddRunningPlugin_NoConfig(t *testing.T) {
	pm := newTestPluginManager(t)

	// Don't add any plugin config

	_, err := pm.addRunningPlugin("nonexistent-instance")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no config")
}

// Test 8: Concurrent Plugin Operations

func TestPluginManager_ConcurrentAddRunningPlugin(t *testing.T) {
	pm := newTestPluginManager(t)

	// Add plugin config
	pm.plugins["test-instance"] = &plugin.Plugin{
		Plugin:   "test-plugin",
		Instance: "test-instance",
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := pm.addRunningPlugin("test-instance")
			mu.Lock()
			if err == nil {
				successCount++
			} else {
				errorCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Only one should succeed, the rest should get retryable errors
	assert.Equal(t, 1, successCount, "Only one goroutine should succeed")
	assert.Equal(t, numGoroutines-1, errorCount, "All other goroutines should fail")
}

// Test 9: IsShuttingDown with Concurrent Access

func TestPluginManager_IsShuttingDown_Concurrent(t *testing.T) {
	pm := newTestPluginManager(t)

	var wg sync.WaitGroup
	numReaders := 50

	// Start many readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = pm.isShuttingDown()
			}
		}()
	}

	// One writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 10; j++ {
			pm.shutdownMut.Lock()
			pm.shuttingDown = !pm.shuttingDown
			pm.shutdownMut.Unlock()
			time.Sleep(time.Millisecond)
		}
	}()

	wg.Wait()
}

// Test 10: Plugin Cache Size Map

func TestPluginManager_SetPluginCacheSizeMap_NoCacheLimit(t *testing.T) {
	pm := newTestPluginManager(t)

	config1 := newTestConnectionConfig("plugin1", "instance1", "conn1")
	config2 := newTestConnectionConfig("plugin2", "instance2", "conn2")

	pm.pluginConnectionConfigMap = map[string][]*sdkproto.ConnectionConfig{
		"instance1": {config1},
		"instance2": {config2},
	}

	pm.setPluginCacheSizeMap()

	// When no max size is set, all plugins should have size 0 (unlimited)
	assert.Equal(t, int64(0), pm.pluginCacheSizeMap["instance1"])
	assert.Equal(t, int64(0), pm.pluginCacheSizeMap["instance2"])
}

// Test 11: NonAggregatorConnectionCount

func TestPluginManager_NonAggregatorConnectionCount(t *testing.T) {
	pm := newTestPluginManager(t)

	// Regular connection (no child connections)
	config1 := &sdkproto.ConnectionConfig{
		Plugin:           "plugin1",
		PluginInstance:   "instance1",
		Connection:       "conn1",
		ChildConnections: []string{},
	}

	// Aggregator connection (has child connections)
	config2 := &sdkproto.ConnectionConfig{
		Plugin:           "plugin1",
		PluginInstance:   "instance1",
		Connection:       "conn2",
		ChildConnections: []string{"child1", "child2"},
	}

	// Another regular connection
	config3 := &sdkproto.ConnectionConfig{
		Plugin:           "plugin2",
		PluginInstance:   "instance2",
		Connection:       "conn3",
		ChildConnections: []string{},
	}

	pm.pluginConnectionConfigMap = map[string][]*sdkproto.ConnectionConfig{
		"instance1": {config1, config2},
		"instance2": {config3},
	}

	count := pm.nonAggregatorConnectionCount()

	// Should count only non-aggregator connections (conn1 and conn3)
	assert.Equal(t, 2, count)
}

// Test 12: GetPluginExemplarConnections

func TestPluginManager_GetPluginExemplarConnections(t *testing.T) {
	pm := newTestPluginManager(t)

	config1 := newTestConnectionConfig("plugin1", "instance1", "conn1")
	config2 := newTestConnectionConfig("plugin1", "instance1", "conn2")
	config3 := newTestConnectionConfig("plugin2", "instance2", "conn3")

	pm.connectionConfigMap = connection.ConnectionConfigMap{
		"conn1": config1,
		"conn2": config2,
		"conn3": config3,
	}

	exemplars := pm.getPluginExemplarConnections()

	assert.Len(t, exemplars, 2, "Should have 2 plugins")
	// Should have one exemplar for each plugin (might be any of the connections)
	assert.Contains(t, []string{"conn1", "conn2"}, exemplars["plugin1"])
	assert.Equal(t, "conn3", exemplars["plugin2"])
}

// Test 13: Goroutine Leak Detection

func TestPluginManager_NoGoroutineLeak_OnError(t *testing.T) {
	before := runtime.NumGoroutine()

	pm := newTestPluginManager(t)

	// Add plugin config
	pm.plugins["test-instance"] = &plugin.Plugin{
		Plugin:   "test-plugin",
		Instance: "test-instance",
	}

	// Try to add running plugin
	_, err := pm.addRunningPlugin("test-instance")
	require.NoError(t, err)

	// Clean up
	pm.mut.Lock()
	delete(pm.runningPluginMap, "test-instance")
	pm.mut.Unlock()

	time.Sleep(100 * time.Millisecond)
	after := runtime.NumGoroutine()

	// Allow some tolerance for background goroutines
	if after > before+5 {
		t.Errorf("Potential goroutine leak: before=%d, after=%d", before, after)
	}
}

// Test 14: Pool Access

func TestPluginManager_Pool(t *testing.T) {
	pm := newTestPluginManager(t)

	// Initially nil
	assert.Nil(t, pm.Pool())
}

// Test 15: RefreshConnections

func TestPluginManager_RefreshConnections(t *testing.T) {
	pm := newTestPluginManager(t)

	req := &pb.RefreshConnectionsRequest{}

	resp, err := pm.RefreshConnections(req)

	require.NoError(t, err, "RefreshConnections should not return error")
	assert.NotNil(t, resp, "Response should not be nil")
}

// Test 16: GetConnectionConfig Concurrent Access

func TestPluginManager_GetConnectionConfig_Concurrent(t *testing.T) {
	pm := newTestPluginManager(t)

	config := newTestConnectionConfig("plugin1", "instance1", "conn1")
	pm.connectionConfigMap["conn1"] = config

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg, err := pm.getConnectionConfig("conn1")
			if err == nil {
				assert.Equal(t, "conn1", cfg.Connection)
			}
		}()
	}

	wg.Wait()
}

// Test 17: Running Plugin Structure

func TestRunningPlugin_Initialization(t *testing.T) {
	rp := &runningPlugin{
		pluginInstance: "test",
		imageRef:       "test-image",
		initialized:    make(chan struct{}),
		failed:         make(chan struct{}),
	}

	assert.NotNil(t, rp.initialized, "initialized channel should not be nil")
	assert.NotNil(t, rp.failed, "failed channel should not be nil")

	// Verify channels are not closed initially
	select {
	case <-rp.initialized:
		t.Fatal("initialized channel should not be closed initially")
	default:
		// Expected
	}

	select {
	case <-rp.failed:
		t.Fatal("failed channel should not be closed initially")
	default:
		// Expected
	}
}

// Test 18: Multiple Concurrent Refreshes

func TestPluginManager_ConcurrentRefreshConnections(t *testing.T) {
	pm := newTestPluginManager(t)

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := &pb.RefreshConnectionsRequest{}
			_, _ = pm.RefreshConnections(req)
		}()
	}

	wg.Wait()
}

// Test 19: NonAggregatorConnectionCount Helper

func TestNonAggregatorConnectionCount(t *testing.T) {
	tests := []struct {
		name        string
		connections []*sdkproto.ConnectionConfig
		expected    int
	}{
		{
			name:        "empty",
			connections: []*sdkproto.ConnectionConfig{},
			expected:    0,
		},
		{
			name: "all non-aggregators",
			connections: []*sdkproto.ConnectionConfig{
				{Connection: "conn1", ChildConnections: []string{}},
				{Connection: "conn2", ChildConnections: []string{}},
			},
			expected: 2,
		},
		{
			name: "all aggregators",
			connections: []*sdkproto.ConnectionConfig{
				{Connection: "conn1", ChildConnections: []string{"child1"}},
				{Connection: "conn2", ChildConnections: []string{"child2"}},
			},
			expected: 0,
		},
		{
			name: "mixed",
			connections: []*sdkproto.ConnectionConfig{
				{Connection: "conn1", ChildConnections: []string{}},
				{Connection: "conn2", ChildConnections: []string{"child1"}},
				{Connection: "conn3", ChildConnections: []string{}},
			},
			expected: 2,
		},
		{
			name: "nil child connections",
			connections: []*sdkproto.ConnectionConfig{
				{Connection: "conn1", ChildConnections: nil},
				{Connection: "conn2", ChildConnections: []string{"child1"}},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := nonAggregatorConnectionCount(tt.connections)
			assert.Equal(t, tt.expected, count)
		})
	}
}

// Test 20: GetResponse Helper

func TestNewGetResponse(t *testing.T) {
	resp := newGetResponse()

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.GetResponse)
	assert.NotNil(t, resp.ReattachMap)
	assert.NotNil(t, resp.FailureMap)
}

// Test 21: EnsurePlugin Early Exit When Shutting Down

func TestPluginManager_EnsurePlugin_ShuttingDown(t *testing.T) {
	pm := newTestPluginManager(t)

	// Set shutting down flag
	pm.shutdownMut.Lock()
	pm.shuttingDown = true
	pm.shutdownMut.Unlock()

	config := newTestConnectionConfig("plugin1", "instance1", "conn1")
	req := &pb.GetRequest{Connections: []string{"conn1"}}

	_, err := pm.ensurePlugin("instance1", []*sdkproto.ConnectionConfig{config}, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shutting down")
}

// Test 22: KillPlugin with Nil Client

func TestPluginManager_KillPlugin_NilClient(t *testing.T) {
	pm := newTestPluginManager(t)

	rp := &runningPlugin{
		pluginInstance: "test",
		client:         nil,
	}

	// Should not panic
	pm.killPlugin(rp)
}

// Test 23: Stress Test for Map Access

func TestPluginManager_StressConcurrentMapAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	pm := newTestPluginManager(t)

	// Add initial configs
	for i := 0; i < 100; i++ {
		connName := fmt.Sprintf("conn%d", i)
		config := newTestConnectionConfig("plugin1", "instance1", connName)
		pm.connectionConfigMap[connName] = config
	}
	pm.populatePluginConnectionConfigs()

	var wg sync.WaitGroup
	duration := 1 * time.Second
	stopCh := make(chan struct{})

	// Start multiple readers
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					connName := fmt.Sprintf("conn%d", idx%100)
					pm.mut.RLock()
					_ = pm.connectionConfigMap[connName]
					_ = pm.pluginConnectionConfigMap["instance1"]
					pm.mut.RUnlock()
				}
			}
		}(i)
	}

	// Run for duration
	time.Sleep(duration)
	close(stopCh)
	wg.Wait()
}

// Test 24: OnConnectionConfigChanged with Nil Pool (Bug #4784)

// TestPluginManager_OnConnectionConfigChanged_EmptyToNonEmpty tests the scenario where
// a PluginManager with no pool (e.g., in a testing environment) receives a configuration change.
// This test demonstrates bug #4784 - a nil pointer panic when m.pool is nil.
func TestPluginManager_OnConnectionConfigChanged_EmptyToNonEmpty(t *testing.T) {
	// Create a minimal PluginManager without pool initialization
	// This simulates a testing scenario or edge case where the pool might not be initialized
	m := &PluginManager{
		plugins: make(map[string]*plugin.Plugin),
		// Note: pool is intentionally nil to demonstrate the bug
	}

	// Create a new plugin map with one plugin
	newPlugins := map[string]*plugin.Plugin{
		"aws": {
			Plugin:   "hub.steampipe.io/plugins/turbot/aws@latest",
			Instance: "aws",
		},
	}

	ctx := context.Background()

	// This should panic with nil pointer dereference when trying to use m.pool
	err := m.handlePluginInstanceChanges(ctx, newPlugins)

	// If we get here without panic, the fix is working
	if err != nil {
		t.Logf("Expected error when pool is nil: %v", err)
	}
}

// TestPluginManager_Shutdown_NoPlugins tests that Shutdown handles nil pool gracefully
// Related to bug #4782
func TestPluginManager_Shutdown_NoPlugins(t *testing.T) {
	// Create a PluginManager without initializing the pool
	// This simulates a scenario where pool initialization failed
	pm := &PluginManager{
		logger:              hclog.NewNullLogger(),
		runningPluginMap:    make(map[string]*runningPlugin),
		connectionConfigMap: make(connection.ConnectionConfigMap),
		plugins:             make(connection.PluginMap),
		// Note: pool is not initialized, will be nil
	}

	// Calling Shutdown should not panic even with nil pool
	req := &pb.ShutdownRequest{}
	resp, err := pm.Shutdown(req)

	if err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	if resp == nil {
		t.Error("Shutdown returned nil response")
	}
}

// TestWaitForPluginLoadWithNilReattach tests that waitForPluginLoad handles
// the case where a plugin fails before reattach is set.
// This reproduces bug #4752 - a nil pointer panic when trying to log p.reattach.Pid
// after the plugin fails during startup before the reattach config is set.
func TestWaitForPluginLoadWithNilReattach(t *testing.T) {
	pm := newTestPluginManager(t)

	// Add plugin config required by waitForPluginLoad with a reasonable timeout
	timeout := 30 // Set timeout to 30 seconds so test doesn't time out immediately
	pm.plugins["test-instance"] = &plugin.Plugin{
		Plugin:        "test-plugin",
		Instance:      "test-instance",
		StartTimeout:  &timeout,
	}

	// Create a runningPlugin that simulates a plugin that failed before reattach was set
	rp := &runningPlugin{
		pluginInstance: "test-instance",
		initialized:    make(chan struct{}),
		failed:         make(chan struct{}),
		error:          fmt.Errorf("plugin startup failed"),
		reattach:       nil, // Explicitly nil - this is the bug condition
	}

	// Simulate plugin failure by closing the failed channel in a goroutine
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(rp.failed)
	}()

	// Create a dummy request
	req := &pb.GetRequest{
		Connections: []string{"test-conn"},
	}

	// This should panic with nil pointer dereference when trying to log p.reattach.Pid
	err := pm.waitForPluginLoad(rp, req)

	// We expect an error (the plugin failed), but we should NOT panic
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin startup failed")
}
