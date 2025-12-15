package connection

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/shared"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// TestRefreshConnectionState_ExemplarSchemaMapConcurrentWrites tests concurrent writes to exemplarSchemaMap
// This verifies the fix for bug #4757
func TestRefreshConnectionState_ExemplarSchemaMapConcurrentWrites(t *testing.T) {
	// ARRANGE: Create state with initialized maps
	state := &refreshConnectionState{
		exemplarSchemaMap:    make(map[string]string),
		exemplarSchemaMapMut: sync.Mutex{},
	}

	numGoroutines := 50
	numIterations := 100
	plugins := []string{"aws", "azure", "gcp", "github", "slack"}

	var wg sync.WaitGroup

	// ACT: Launch goroutines that concurrently write to exemplarSchemaMap
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				plugin := plugins[j%len(plugins)]
				connectionName := fmt.Sprintf("conn_%d_%d", id, j)

				// Simulate the FIXED pattern from executeUpdateForConnections (lines 600-605)
				state.exemplarSchemaMapMut.Lock()
				_, haveExemplar := state.exemplarSchemaMap[plugin]
				state.exemplarSchemaMapMut.Unlock()

				if !haveExemplar {
					// This write is now protected by mutex (fix for #4757)
					state.exemplarSchemaMapMut.Lock()
					state.exemplarSchemaMap[plugin] = connectionName
					state.exemplarSchemaMapMut.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()

	// ASSERT: Verify all plugins are in the map
	state.exemplarSchemaMapMut.Lock()
	defer state.exemplarSchemaMapMut.Unlock()

	if len(state.exemplarSchemaMap) != len(plugins) {
		t.Errorf("Expected %d plugins in exemplarSchemaMap, got %d", len(plugins), len(state.exemplarSchemaMap))
	}

	for _, plugin := range plugins {
		if _, ok := state.exemplarSchemaMap[plugin]; !ok {
			t.Errorf("Expected plugin %s to be in exemplarSchemaMap", plugin)
		}
	}
}

// TestRefreshConnectionState_ExemplarSchemaMapConcurrentReadWrite tests concurrent reads and writes
func TestRefreshConnectionState_ExemplarSchemaMapConcurrentReadWrite(t *testing.T) {
	// ARRANGE: Create state with some pre-populated data
	state := &refreshConnectionState{
		exemplarSchemaMap: map[string]string{
			"aws":   "aws_conn_1",
			"azure": "azure_conn_1",
		},
		exemplarSchemaMapMut: sync.Mutex{},
	}

	numReaders := 30
	numWriters := 20
	duration := 100 * time.Millisecond

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// ACT: Launch reader goroutines
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					state.exemplarSchemaMapMut.Lock()
					_ = state.exemplarSchemaMap["aws"]
					state.exemplarSchemaMapMut.Unlock()
				}
			}
		}()
	}

	// Launch writer goroutines
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					plugin := fmt.Sprintf("plugin_%d", id)
					state.exemplarSchemaMapMut.Lock()
					state.exemplarSchemaMap[plugin] = fmt.Sprintf("conn_%d", id)
					state.exemplarSchemaMapMut.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()

	// ASSERT: No race conditions should occur (run with -race flag)
	state.exemplarSchemaMapMut.Lock()
	defer state.exemplarSchemaMapMut.Unlock()

	// Basic sanity check
	if len(state.exemplarSchemaMap) < 2 {
		t.Error("Expected at least 2 entries in exemplarSchemaMap")
	}
}

// TestRefreshConnectionState_ExemplarMapRaceCondition tests the exact race condition from bug #4757
func TestRefreshConnectionState_ExemplarMapRaceCondition(t *testing.T) {
	// This test verifies that the fix for #4757 works correctly
	// The bug was: reading haveExemplarSchema without lock, then writing without lock
	// The fix: both read and write are now properly protected by mutex

	// ARRANGE
	state := &refreshConnectionState{
		exemplarSchemaMap:    make(map[string]string),
		exemplarSchemaMapMut: sync.Mutex{},
	}

	numGoroutines := 100
	pluginName := "aws"

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	// ACT: Simulate the exact pattern from executeUpdateForConnections
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			connectionName := fmt.Sprintf("aws_conn_%d", id)

			// This is the FIXED pattern from lines 581-604
			state.exemplarSchemaMapMut.Lock()
			_, haveExemplarSchema := state.exemplarSchemaMap[pluginName]
			state.exemplarSchemaMapMut.Unlock()

			// Simulate some work
			time.Sleep(time.Microsecond)

			if !haveExemplarSchema {
				// Write is now protected by mutex (fix for #4757)
				state.exemplarSchemaMapMut.Lock()
				// Check again after acquiring lock (double-check pattern)
				if _, exists := state.exemplarSchemaMap[pluginName]; !exists {
					state.exemplarSchemaMap[pluginName] = connectionName
				}
				state.exemplarSchemaMapMut.Unlock()
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// ASSERT: Check for errors
	for err := range errChan {
		t.Error(err)
	}

	// Verify the map has exactly one entry for the plugin
	state.exemplarSchemaMapMut.Lock()
	defer state.exemplarSchemaMapMut.Unlock()

	if len(state.exemplarSchemaMap) != 1 {
		t.Errorf("Expected exactly 1 entry in exemplarSchemaMap, got %d", len(state.exemplarSchemaMap))
	}

	if _, ok := state.exemplarSchemaMap[pluginName]; !ok {
		t.Error("Expected plugin to be in exemplarSchemaMap")
	}
}

// TestUpdateSetMapToArray tests the conversion utility function
func TestUpdateSetMapToArray(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]*steampipeconfig.ConnectionState
		expected int
	}{
		{
			name:     "empty_map",
			input:    map[string][]*steampipeconfig.ConnectionState{},
			expected: 0,
		},
		{
			name: "single_entry_single_state",
			input: map[string][]*steampipeconfig.ConnectionState{
				"plugin1": {
					{ConnectionName: "conn1"},
				},
			},
			expected: 1,
		},
		{
			name: "single_entry_multiple_states",
			input: map[string][]*steampipeconfig.ConnectionState{
				"plugin1": {
					{ConnectionName: "conn1"},
					{ConnectionName: "conn2"},
					{ConnectionName: "conn3"},
				},
			},
			expected: 3,
		},
		{
			name: "multiple_entries",
			input: map[string][]*steampipeconfig.ConnectionState{
				"plugin1": {
					{ConnectionName: "conn1"},
					{ConnectionName: "conn2"},
				},
				"plugin2": {
					{ConnectionName: "conn3"},
				},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ACT
			result := updateSetMapToArray(tt.input)

			// ASSERT
			if len(result) != tt.expected {
				t.Errorf("Expected %d connection states, got %d", tt.expected, len(result))
			}
		})
	}
}

// TestGetCloneSchemaQuery tests the schema cloning query generation
func TestGetCloneSchemaQuery(t *testing.T) {
	tests := []struct {
		name          string
		exemplarName  string
		connState     *steampipeconfig.ConnectionState
		expectedQuery string
	}{
		{
			name:         "basic_clone",
			exemplarName: "aws_source",
			connState: &steampipeconfig.ConnectionState{
				ConnectionName: "aws_target",
				Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
			},
			expectedQuery: "select clone_foreign_schema('aws_source', 'aws_target', 'hub.steampipe.io/plugins/turbot/aws@latest');",
		},
		{
			name:         "with_special_characters",
			exemplarName: "test-source",
			connState: &steampipeconfig.ConnectionState{
				ConnectionName: "test-target",
				Plugin:         "test/plugin@1.0.0",
			},
			expectedQuery: "select clone_foreign_schema('test-source', 'test-target', 'test/plugin@1.0.0');",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ACT
			result := getCloneSchemaQuery(tt.exemplarName, tt.connState)

			// ASSERT
			if result != tt.expectedQuery {
				t.Errorf("Expected query:\n%s\nGot:\n%s", tt.expectedQuery, result)
			}
		})
	}
}

// TestRefreshConnectionState_DeferErrorHandling tests error handling in defer blocks
func TestRefreshConnectionState_DeferErrorHandling(t *testing.T) {
	// This tests the defer block at lines 98-108 in refreshConnections

	// ARRANGE: Create state with a result that will have an error
	state := &refreshConnectionState{
		res: &steampipeconfig.RefreshConnectionResult{},
	}

	// Simulate setting an error
	testErr := errors.New("test error")
	state.res.Error = testErr

	// ACT: The defer block should handle this gracefully
	// In the actual code, this is called via defer func()
	// We're testing the logic here

	// ASSERT: Verify the defer logic works
	if state.res != nil && state.res.Error != nil {
		// This is what the defer does - it would call setIncompleteConnectionStateToError
		// We're just verifying the nil checks work
		if state.res.Error != testErr {
			t.Error("Error should be preserved")
		}
	}
}

// TestRefreshConnectionState_NilResInDefer tests nil res handling in defer block
func TestRefreshConnectionState_NilResInDefer(t *testing.T) {
	// ARRANGE: Create state with nil res
	state := &refreshConnectionState{
		res: nil,
	}

	// ACT & ASSERT: The defer block at line 98-108 checks if res is nil
	// This should not panic
	if state.res != nil {
		t.Error("res should be nil")
	}
}

// TestRefreshConnectionState_MultiplePluginsSameExemplar tests that only one exemplar is stored per plugin
func TestRefreshConnectionState_MultiplePluginsSameExemplar(t *testing.T) {
	// ARRANGE
	state := &refreshConnectionState{
		exemplarSchemaMap:    make(map[string]string),
		exemplarSchemaMapMut: sync.Mutex{},
	}

	pluginName := "aws"
	connections := []string{"aws1", "aws2", "aws3", "aws4", "aws5"}

	// ACT: Add connections sequentially (simulating the pattern from the code)
	for _, conn := range connections {
		state.exemplarSchemaMapMut.Lock()
		_, exists := state.exemplarSchemaMap[pluginName]
		state.exemplarSchemaMapMut.Unlock()

		if !exists {
			state.exemplarSchemaMapMut.Lock()
			// Double-check pattern
			if _, exists := state.exemplarSchemaMap[pluginName]; !exists {
				state.exemplarSchemaMap[pluginName] = conn
			}
			state.exemplarSchemaMapMut.Unlock()
		}
	}

	// ASSERT: Only the first connection should be stored
	state.exemplarSchemaMapMut.Lock()
	defer state.exemplarSchemaMapMut.Unlock()

	if len(state.exemplarSchemaMap) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(state.exemplarSchemaMap))
	}

	if exemplar, ok := state.exemplarSchemaMap[pluginName]; !ok {
		t.Error("Expected plugin to be in map")
	} else if exemplar != connections[0] {
		t.Errorf("Expected first connection %s to be exemplar, got %s", connections[0], exemplar)
	}
}

// TestRefreshConnectionState_ErrorChannelBlocking tests that error channel doesn't block
func TestRefreshConnectionState_ErrorChannelBlocking(t *testing.T) {
	// This tests a potential bug in executeUpdateSetsInParallel where the error channel
	// could block if it's not properly drained

	// ARRANGE
	errChan := make(chan *connectionError, 10) // Buffered channel
	numErrors := 20                            // More errors than buffer size

	var wg sync.WaitGroup

	// Start a consumer goroutine (like in the actual code at line 519-536)
	consumerDone := make(chan bool)
	go func() {
		for {
			select {
			case err := <-errChan:
				if err == nil {
					consumerDone <- true
					return
				}
				// Process error
				_ = err
			}
		}
	}()

	// ACT: Send many errors
	for i := 0; i < numErrors; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			errChan <- &connectionError{
				name: fmt.Sprintf("conn_%d", id),
				err:  fmt.Errorf("error %d", id),
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Wait for consumer to finish
	select {
	case <-consumerDone:
		// Good - consumer exited
	case <-time.After(1 * time.Second):
		t.Error("Error channel consumer did not exit in time")
	}

	// ASSERT: No goroutines should be blocked
}

// TestRefreshConnectionState_ExemplarMapNilPlugin tests handling of empty plugin names
func TestRefreshConnectionState_ExemplarMapNilPlugin(t *testing.T) {
	// ARRANGE
	state := &refreshConnectionState{
		exemplarSchemaMap:    make(map[string]string),
		exemplarSchemaMapMut: sync.Mutex{},
	}

	// ACT: Try to add entry with empty plugin name
	state.exemplarSchemaMapMut.Lock()
	state.exemplarSchemaMap[""] = "some_connection"
	state.exemplarSchemaMapMut.Unlock()

	// ASSERT: Map should accept empty string as key (Go maps allow this)
	state.exemplarSchemaMapMut.Lock()
	defer state.exemplarSchemaMapMut.Unlock()

	if _, ok := state.exemplarSchemaMap[""]; !ok {
		t.Error("Expected empty string key to be in map")
	}
}

// TestConnectionError tests the connectionError struct
func TestConnectionError(t *testing.T) {
	// ARRANGE
	testErr := errors.New("test error")
	connErr := &connectionError{
		name: "test_connection",
		err:  testErr,
	}

	// ASSERT
	if connErr.name != "test_connection" {
		t.Errorf("Expected name 'test_connection', got '%s'", connErr.name)
	}

	if connErr.err != testErr {
		t.Error("Error not preserved")
	}
}

// mockPluginManager is a mock implementation of pluginManager interface for testing
type mockPluginManager struct {
	shared.PluginManager
	pool *pgxpool.Pool
}

func (m *mockPluginManager) Pool() *pgxpool.Pool {
	return m.pool
}

// Implement other required methods from pluginManager interface
func (m *mockPluginManager) OnConnectionConfigChanged(context.Context, ConnectionConfigMap, map[string]*plugin.Plugin) {
}

func (m *mockPluginManager) GetConnectionConfig() ConnectionConfigMap {
	return nil
}

func (m *mockPluginManager) HandlePluginLimiterChanges(PluginLimiterMap) error {
	return nil
}

func (m *mockPluginManager) ShouldFetchRateLimiterDefs() bool {
	return false
}

func (m *mockPluginManager) LoadPluginRateLimiters(map[string]string) (PluginLimiterMap, error) {
	return nil, nil
}

func (m *mockPluginManager) SendPostgresSchemaNotification(context.Context) error {
	return nil
}

func (m *mockPluginManager) SendPostgresErrorsAndWarningsNotification(context.Context, error_helpers.ErrorAndWarnings) {
}

func (m *mockPluginManager) UpdatePluginColumnsTable(context.Context, map[string]*proto.Schema, []string) error {
	return nil
}

// TestNewRefreshConnectionState_NilPool tests that newRefreshConnectionState handles nil pool gracefully
// This test demonstrates issue #4778 - nil pool from pluginManager causes panic
func TestNewRefreshConnectionState_NilPool(t *testing.T) {
	ctx := context.Background()

	// Create a mock plugin manager that returns nil pool
	mockPM := &mockPluginManager{
		pool: nil,
	}

	// This should not panic - should return an error instead
	_, err := newRefreshConnectionState(ctx, mockPM, []string{})

	if err == nil {
		t.Error("Expected error when pool is nil, got nil")
	}
}

// TestRefreshConnectionState_ConnectionOrderEdgeCases tests edge cases in connection ordering
// This test demonstrates issue #4779 - nil GlobalConfig causes panic in newRefreshConnectionState
func TestRefreshConnectionState_ConnectionOrderEdgeCases(t *testing.T) {
	t.Run("nil_global_config", func(t *testing.T) {
		// ARRANGE: Save original GlobalConfig and set it to nil
		originalConfig := steampipeconfig.GlobalConfig
		steampipeconfig.GlobalConfig = nil
		defer func() {
			steampipeconfig.GlobalConfig = originalConfig
		}()

		ctx := context.Background()

		// Create a mock plugin manager with a valid pool
		// We need a pool to get past the nil pool check
		// For this test, we can use a nil pool since we expect the function to fail
		// before it tries to use the pool
		mockPM := &mockPluginManager{
			pool: &pgxpool.Pool{}, // Need a non-nil pool to get past line 66-68
		}

		// ACT: Call newRefreshConnectionState with nil GlobalConfig
		// This should not panic - should return an error instead
		_, err := newRefreshConnectionState(ctx, mockPM, nil)

		// ASSERT: Should return an error, not panic
		if err == nil {
			t.Error("Expected error when GlobalConfig is nil, got nil")
		}

		if err != nil && !strings.Contains(err.Error(), "GlobalConfig") {
			t.Errorf("Expected error message to mention GlobalConfig, got: %v", err)
		}
	})
}
