package pluginmanager_service

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

// Test helpers for message server tests

func newTestMessageServer(t *testing.T) *PluginMessageServer {
	t.Helper()
	pm := newTestPluginManager(t)
	return &PluginMessageServer{
		pluginManager: pm,
	}
}

// Test 1: NewPluginMessageServer

func TestNewPluginMessageServer(t *testing.T) {
	pm := newTestPluginManager(t)

	ms, err := NewPluginMessageServer(pm)

	require.NoError(t, err)
	assert.NotNil(t, ms)
	assert.Equal(t, pm, ms.pluginManager)
}

// Test 2: PluginMessageServer Initialization

func TestPluginManager_MessageServerInitialization(t *testing.T) {
	pm := newTestPluginManager(t)

	assert.NotNil(t, pm.messageServer, "messageServer should be initialized")
	assert.Equal(t, pm, pm.messageServer.pluginManager, "messageServer should reference parent PluginManager")
}

// Test 3: Concurrent Access

func TestPluginMessageServer_ConcurrentAccess(t *testing.T) {
	ms := newTestMessageServer(t)

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ms.pluginManager
		}()
	}

	wg.Wait()
}

// Test 4: LogReceiveError with Valid Errors

func TestPluginMessageServer_LogReceiveError(t *testing.T) {
	ms := newTestMessageServer(t)

	// Should not panic for various error types
	ms.logReceiveError(context.Canceled, "test-connection")
	ms.logReceiveError(context.DeadlineExceeded, "test-connection")
}

// TestPluginMessageServer_LogReceiveError_NilError tests that logReceiveError
// handles nil error gracefully without panicking
func TestPluginMessageServer_LogReceiveError_NilError(t *testing.T) {
	// Create a message server
	pm := &PluginManager{}
	server := &PluginMessageServer{
		pluginManager: pm,
	}

	// This should not panic - calling logReceiveError with nil error
	server.logReceiveError(nil, "test-connection")
}

// Test 5: Multiple Message Servers

func TestPluginManager_MultipleMessageServers(t *testing.T) {
	pm := newTestPluginManager(t)

	ms1, err1 := NewPluginMessageServer(pm)
	ms2, err2 := NewPluginMessageServer(pm)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotNil(t, ms1)
	assert.NotNil(t, ms2)

	// Both should reference the same plugin manager
	assert.Equal(t, pm, ms1.pluginManager)
	assert.Equal(t, pm, ms2.pluginManager)
}

// Test 6: Message Server with Nil Plugin Manager

func TestPluginMessageServer_NilPluginManager(t *testing.T) {
	ms := &PluginMessageServer{
		pluginManager: nil,
	}

	assert.Nil(t, ms.pluginManager)
}

// Test 7: Goroutine Cleanup

func TestPluginMessageServer_GoroutineCleanup(t *testing.T) {
	before := runtime.NumGoroutine()

	ms := newTestMessageServer(t)
	_ = ms

	time.Sleep(100 * time.Millisecond)
	after := runtime.NumGoroutine()

	// Creating a message server shouldn't leak goroutines
	if after > before+5 {
		t.Errorf("Potential goroutine leak: before=%d, after=%d", before, after)
	}
}

// Test 8: Message Type Structure

func TestPluginMessage_SchemaUpdatedType(t *testing.T) {
	message := &sdkproto.PluginMessage{
		MessageType: sdkproto.PluginMessageType_SCHEMA_UPDATED,
		Connection:  "test-connection",
	}

	assert.Equal(t, sdkproto.PluginMessageType_SCHEMA_UPDATED, message.MessageType)
	assert.Equal(t, "test-connection", message.Connection)
}

// Test 9: LogReceiveError with Different Error Types

func TestPluginMessageServer_LogReceiveError_ErrorTypes(t *testing.T) {
	ms := newTestMessageServer(t)

	// Test various error types don't cause panics
	errors := []error{
		context.Canceled,
		context.DeadlineExceeded,
		assert.AnError,
	}

	for _, err := range errors {
		ms.logReceiveError(err, "test-connection")
	}
}

// Test 10: Message Server Initialization Consistency

func TestPluginManager_MessageServer_Consistency(t *testing.T) {
	pm := newTestPluginManager(t)

	// Verify messageServer is initialized and consistent
	assert.NotNil(t, pm.messageServer)
	assert.Equal(t, pm, pm.messageServer.pluginManager)

	// Accessing it multiple times should return the same instance
	ms1 := pm.messageServer
	ms2 := pm.messageServer
	assert.Equal(t, ms1, ms2)
}

// Test 11: Message Server Survives Plugin Manager Operations

func TestPluginMessageServer_SurvivesPluginManagerOperations(t *testing.T) {
	pm := newTestPluginManager(t)
	ms := pm.messageServer

	// Perform various plugin manager operations
	pm.populatePluginConnectionConfigs()
	pm.setPluginCacheSizeMap()
	pm.nonAggregatorConnectionCount()

	// Message server should still be accessible
	assert.Equal(t, pm, ms.pluginManager)
	assert.NotNil(t, pm.messageServer)
}

// Test 12: Concurrent NewPluginMessageServer Calls

func TestNewPluginMessageServer_Concurrent(t *testing.T) {
	pm := newTestPluginManager(t)

	var wg sync.WaitGroup
	numGoroutines := 50
	servers := make([]*PluginMessageServer, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			servers[idx], errors[idx] = NewPluginMessageServer(pm)
		}(i)
	}

	wg.Wait()

	// All should succeed
	for i := 0; i < numGoroutines; i++ {
		assert.NoError(t, errors[i])
		assert.NotNil(t, servers[i])
		assert.Equal(t, pm, servers[i].pluginManager)
	}
}

// Test 13: Message Server Pointer Stability

func TestPluginMessageServer_PointerStability(t *testing.T) {
	pm := newTestPluginManager(t)

	ms1 := pm.messageServer
	ms2 := pm.messageServer

	// Should be the same pointer
	assert.True(t, ms1 == ms2, "messageServer pointer should be stable")
}

// Test 14: LogReceiveError Concurrent Calls

func TestPluginMessageServer_LogReceiveError_Concurrent(t *testing.T) {
	ms := newTestMessageServer(t)

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := assert.AnError
			if idx%2 == 0 {
				err = context.Canceled
			}
			ms.logReceiveError(err, "test-connection")
		}(i)
	}

	wg.Wait()
}

// Test 15: Message Server Field Access

func TestPluginMessageServer_FieldAccess(t *testing.T) {
	ms := newTestMessageServer(t)

	// Verify fields are accessible and not nil
	assert.NotNil(t, ms.pluginManager)
	assert.NotNil(t, ms.pluginManager.logger)
	assert.NotNil(t, ms.pluginManager.runningPluginMap)
}

// Test 16: Message Server Doesn't Block Plugin Manager

func TestPluginMessageServer_DoesNotBlockPluginManager(t *testing.T) {
	pm := newTestPluginManager(t)

	// Message server should not prevent these operations
	config := newTestConnectionConfig("plugin1", "instance1", "conn1")
	pm.connectionConfigMap["conn1"] = config
	pm.populatePluginConnectionConfigs()

	// Verify operations worked
	assert.Len(t, pm.pluginConnectionConfigMap, 1)

	// Message server should still be valid
	assert.NotNil(t, pm.messageServer)
	assert.Equal(t, pm, pm.messageServer.pluginManager)
}

// Test 17: Stress Test for Concurrent Access

func TestPluginMessageServer_StressConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	pm := newTestPluginManager(t)
	ms := pm.messageServer

	var wg sync.WaitGroup
	duration := 1 * time.Second
	stopCh := make(chan struct{})

	// Multiple readers accessing pluginManager
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					_ = ms.pluginManager
					if ms.pluginManager != nil {
						_ = ms.pluginManager.connectionConfigMap
					}
				}
			}
		}()
	}

	time.Sleep(duration)
	close(stopCh)
	wg.Wait()
}

// Test 18: UpdateConnectionSchema with Nil Pool
// Tests that updateConnectionSchema handles nil pool gracefully without panicking
// Issue #4783: The method calls RefreshConnections which accesses m.pool before the nil check
func TestPluginManager_UpdateConnectionSchema_NilPool(t *testing.T) {
	// Create a PluginManager with a nil pool
	pm := &PluginManager{
		runningPluginMap: make(map[string]*runningPlugin),
		pool:             nil, // explicitly nil pool
	}

	ctx := context.Background()

	// This should not panic - calling updateConnectionSchema with nil pool
	// Previously this would panic because RefreshConnections accesses pool before nil check
	pm.updateConnectionSchema(ctx, "test-connection")

	// If we get here without panicking, the test passes
}

// Test 19: UpdateConnectionSchema with Nil Pool Concurrent
// Tests that concurrent calls to updateConnectionSchema with nil pool don't cause race conditions or panics
func TestPluginManager_UpdateConnectionSchema_NilPool_Concurrent(t *testing.T) {
	pm := &PluginManager{
		runningPluginMap: make(map[string]*runningPlugin),
		pool:             nil,
	}

	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Should not panic
			pm.updateConnectionSchema(ctx, "test-connection")
		}(i)
	}

	wg.Wait()
}
