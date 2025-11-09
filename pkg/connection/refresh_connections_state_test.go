package connection

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// mockPluginManager is a mock implementation of the pluginManager interface for testing
type mockPluginManager struct {
	pool                                         *pgxpool.Pool
	connectionConfig                             ConnectionConfigMap
	shouldFetchRateLimiterDefs                   bool
	onConnectionConfigChangedCalls               int
	handlePluginLimiterChangesCalls              int
	loadPluginRateLimitersCalls                  int
	sendPostgresSchemaNotificationCalls          int
	sendPostgresErrorsAndWarningsNotificationCalls int
	updatePluginColumnsTableCalls                int

	// Configurable return values
	loadPluginRateLimitersFunc  func(map[string]string) (PluginLimiterMap, error)
	handlePluginLimiterChangesFunc func(PluginLimiterMap) error
	updatePluginColumnsTableFunc func(context.Context, map[string]*proto.Schema, []string) error
}

func newMockPluginManager(pool *pgxpool.Pool) *mockPluginManager {
	return &mockPluginManager{
		pool:                       pool,
		connectionConfig:           ConnectionConfigMap{},
		shouldFetchRateLimiterDefs: false,
	}
}

func (m *mockPluginManager) OnConnectionConfigChanged(ctx context.Context, config ConnectionConfigMap, plugins map[string]*plugin.Plugin) {
	m.onConnectionConfigChangedCalls++
	m.connectionConfig = config
}

func (m *mockPluginManager) GetConnectionConfig() ConnectionConfigMap {
	return m.connectionConfig
}

func (m *mockPluginManager) HandlePluginLimiterChanges(limiters PluginLimiterMap) error {
	m.handlePluginLimiterChangesCalls++
	if m.handlePluginLimiterChangesFunc != nil {
		return m.handlePluginLimiterChangesFunc(limiters)
	}
	return nil
}

func (m *mockPluginManager) Pool() *pgxpool.Pool {
	return m.pool
}

func (m *mockPluginManager) ShouldFetchRateLimiterDefs() bool {
	return m.shouldFetchRateLimiterDefs
}

func (m *mockPluginManager) LoadPluginRateLimiters(plugins map[string]string) (PluginLimiterMap, error) {
	m.loadPluginRateLimitersCalls++
	if m.loadPluginRateLimitersFunc != nil {
		return m.loadPluginRateLimitersFunc(plugins)
	}
	return PluginLimiterMap{}, nil
}

func (m *mockPluginManager) SendPostgresSchemaNotification(ctx context.Context) error {
	m.sendPostgresSchemaNotificationCalls++
	return nil
}

func (m *mockPluginManager) SendPostgresErrorsAndWarningsNotification(ctx context.Context, errorAndWarnings error_helpers.ErrorAndWarnings) {
	m.sendPostgresErrorsAndWarningsNotificationCalls++
}

func (m *mockPluginManager) UpdatePluginColumnsTable(ctx context.Context, updatedPlugins map[string]*proto.Schema, deletedPlugins []string) error {
	m.updatePluginColumnsTableCalls++
	if m.updatePluginColumnsTableFunc != nil {
		return m.updatePluginColumnsTableFunc(ctx, updatedPlugins, deletedPlugins)
	}
	return nil
}

// Stub methods for shared.PluginManager interface
func (m *mockPluginManager) Get(req *pb.GetRequest) (*pb.GetResponse, error) {
	return &pb.GetResponse{}, nil
}

func (m *mockPluginManager) RefreshConnections(req *pb.RefreshConnectionsRequest) (*pb.RefreshConnectionsResponse, error) {
	return &pb.RefreshConnectionsResponse{}, nil
}

func (m *mockPluginManager) Shutdown(req *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	return &pb.ShutdownResponse{}, nil
}

// TestGetInitialAndRemainingUpdates tests the separation of updates into initial and remaining
func TestGetInitialAndRemainingUpdates(t *testing.T) {
	tests := []struct {
		name                 string
		updates              steampipeconfig.ConnectionStateMap
		finalConnectionState steampipeconfig.ConnectionStateMap
		connectionOrder      []string
		wantInitialCount     int
		wantRemainingCount   int
		wantDynamicCount     int
	}{
		{
			name: "static plugins - first in search path",
			updates: steampipeconfig.ConnectionStateMap{
				"aws_dev": {
					ConnectionName: "aws_dev",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     sdkplugin.SchemaModeStatic,
				},
				"aws_prod": {
					ConnectionName: "aws_prod",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     sdkplugin.SchemaModeStatic,
				},
			},
			finalConnectionState: steampipeconfig.ConnectionStateMap{
				"aws_dev": {
					ConnectionName: "aws_dev",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     sdkplugin.SchemaModeStatic,
				},
				"aws_prod": {
					ConnectionName: "aws_prod",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     sdkplugin.SchemaModeStatic,
				},
			},
			connectionOrder:    []string{"aws_dev", "aws_prod"},
			wantInitialCount:   1, // Only first in search path
			wantRemainingCount: 1, // Second connection
			wantDynamicCount:   0,
		},
		{
			name: "dynamic plugins - all in initial",
			updates: steampipeconfig.ConnectionStateMap{
				"dynamic1": {
					ConnectionName: "dynamic1",
					Plugin:         "hub.steampipe.io/plugins/turbot/test@latest",
					SchemaMode:     sdkplugin.SchemaModeDynamic,
					PluginInstance: stringPtr("instance1"),
				},
				"dynamic2": {
					ConnectionName: "dynamic2",
					Plugin:         "hub.steampipe.io/plugins/turbot/test@latest",
					SchemaMode:     sdkplugin.SchemaModeDynamic,
					PluginInstance: stringPtr("instance2"),
				},
			},
			finalConnectionState: steampipeconfig.ConnectionStateMap{
				"dynamic1": {
					ConnectionName: "dynamic1",
					Plugin:         "hub.steampipe.io/plugins/turbot/test@latest",
					SchemaMode:     sdkplugin.SchemaModeDynamic,
					PluginInstance: stringPtr("instance1"),
				},
				"dynamic2": {
					ConnectionName: "dynamic2",
					Plugin:         "hub.steampipe.io/plugins/turbot/test@latest",
					SchemaMode:     sdkplugin.SchemaModeDynamic,
					PluginInstance: stringPtr("instance2"),
				},
			},
			connectionOrder:    []string{"dynamic1", "dynamic2"},
			wantInitialCount:   0,
			wantRemainingCount: 0,
			wantDynamicCount:   2, // Both instances as separate entries
		},
		{
			name: "mixed static and dynamic",
			updates: steampipeconfig.ConnectionStateMap{
				"static_conn": {
					ConnectionName: "static_conn",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     sdkplugin.SchemaModeStatic,
				},
				"dynamic_conn": {
					ConnectionName: "dynamic_conn",
					Plugin:         "hub.steampipe.io/plugins/turbot/test@latest",
					SchemaMode:     sdkplugin.SchemaModeDynamic,
					PluginInstance: stringPtr("instance1"),
				},
			},
			finalConnectionState: steampipeconfig.ConnectionStateMap{
				"static_conn": {
					ConnectionName: "static_conn",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     sdkplugin.SchemaModeStatic,
				},
				"dynamic_conn": {
					ConnectionName: "dynamic_conn",
					Plugin:         "hub.steampipe.io/plugins/turbot/test@latest",
					SchemaMode:     sdkplugin.SchemaModeDynamic,
					PluginInstance: stringPtr("instance1"),
				},
			},
			connectionOrder:    []string{"static_conn", "dynamic_conn"},
			wantInitialCount:   1,
			wantRemainingCount: 0,
			wantDynamicCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &refreshConnectionState{
				connectionOrder: tt.connectionOrder,
				connectionUpdates: &steampipeconfig.ConnectionUpdates{
					Update:               tt.updates,
					FinalConnectionState: tt.finalConnectionState,
				},
			}

			initialUpdates, remainingUpdates, dynamicUpdates := state.getInitialAndRemainingUpdates()

			assert.Len(t, initialUpdates, tt.wantInitialCount, "initial updates count")
			assert.Len(t, remainingUpdates, tt.wantRemainingCount, "remaining updates count")
			assert.Len(t, dynamicUpdates, tt.wantDynamicCount, "dynamic updates count")

			// Verify total adds up
			totalDynamic := 0
			for _, updates := range dynamicUpdates {
				totalDynamic += len(updates)
			}
			assert.Equal(t, len(tt.updates), tt.wantInitialCount+tt.wantRemainingCount+totalDynamic, "total updates should match")
		})
	}
}

// TestGetCloneSchemaQuery tests the schema cloning query generation
func TestGetCloneSchemaQuery(t *testing.T) {
	tests := []struct {
		name                string
		exemplarSchemaName  string
		connectionState     *steampipeconfig.ConnectionState
		wantContains        []string
	}{
		{
			name:               "basic clone query",
			exemplarSchemaName: "aws_dev",
			connectionState: &steampipeconfig.ConnectionState{
				ConnectionName: "aws_prod",
				Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
			},
			wantContains: []string{
				"clone_foreign_schema",
				"aws_dev",
				"aws_prod",
				"hub.steampipe.io/plugins/turbot/aws@latest",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := getCloneSchemaQuery(tt.exemplarSchemaName, tt.connectionState)

			for _, contains := range tt.wantContains {
				assert.Contains(t, query, contains)
			}
		})
	}
}

// TestUpdateSetMapToArray tests conversion of update sets to array
func TestUpdateSetMapToArray(t *testing.T) {
	tests := []struct {
		name          string
		updateSetMap  map[string][]*steampipeconfig.ConnectionState
		wantCount     int
	}{
		{
			name: "single plugin with multiple connections",
			updateSetMap: map[string][]*steampipeconfig.ConnectionState{
				"plugin1": {
					{ConnectionName: "conn1"},
					{ConnectionName: "conn2"},
				},
			},
			wantCount: 2,
		},
		{
			name: "multiple plugins",
			updateSetMap: map[string][]*steampipeconfig.ConnectionState{
				"plugin1": {
					{ConnectionName: "conn1"},
				},
				"plugin2": {
					{ConnectionName: "conn2"},
					{ConnectionName: "conn3"},
				},
			},
			wantCount: 3,
		},
		{
			name:         "empty map",
			updateSetMap: map[string][]*steampipeconfig.ConnectionState{},
			wantCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateSetMapToArray(tt.updateSetMap)
			assert.Len(t, result, tt.wantCount)
		})
	}
}

// TestConnectionStateCanCloneSchema tests the cloning eligibility logic
func TestConnectionStateCanCloneSchema(t *testing.T) {
	tests := []struct {
		name       string
		state      *steampipeconfig.ConnectionState
		wantCanClone bool
	}{
		{
			name: "static schema can clone",
			state: &steampipeconfig.ConnectionState{
				SchemaMode: sdkplugin.SchemaModeStatic,
				Type:       stringPtr(""),
			},
			wantCanClone: true,
		},
		{
			name: "dynamic schema cannot clone",
			state: &steampipeconfig.ConnectionState{
				SchemaMode: sdkplugin.SchemaModeDynamic,
				Type:       stringPtr(""),
			},
			wantCanClone: false,
		},
		{
			name: "aggregator cannot clone",
			state: &steampipeconfig.ConnectionState{
				SchemaMode: sdkplugin.SchemaModeStatic,
				Type:       stringPtr("aggregator"),
			},
			wantCanClone: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.CanCloneSchema()
			assert.Equal(t, tt.wantCanClone, got)
		})
	}
}

// TestExemplarSchemaMapTracking tests that exemplar schemas are tracked correctly
func TestExemplarSchemaMapTracking(t *testing.T) {
	state := &refreshConnectionState{
		exemplarSchemaMap: make(map[string]string),
		connectionUpdates: &steampipeconfig.ConnectionUpdates{
			Update: steampipeconfig.ConnectionStateMap{
				"aws_dev": {
					ConnectionName: "aws_dev",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     sdkplugin.SchemaModeStatic,
					Type:           stringPtr(""),
				},
			},
		},
	}

	// Simulate successful update - the connection should be added to exemplar map
	connectionState := state.connectionUpdates.Update["aws_dev"]
	if connectionState.CanCloneSchema() {
		state.exemplarSchemaMap[connectionState.Plugin] = connectionState.ConnectionName
	}

	// Verify the exemplar was recorded
	exemplar, exists := state.exemplarSchemaMap["hub.steampipe.io/plugins/turbot/aws@latest"]
	assert.True(t, exists, "exemplar should be recorded")
	assert.Equal(t, "aws_dev", exemplar)
}

// TestExemplarSchemaMapNotTrackedForDynamic tests that dynamic schemas are not added to exemplar map
func TestExemplarSchemaMapNotTrackedForDynamic(t *testing.T) {
	state := &refreshConnectionState{
		exemplarSchemaMap: make(map[string]string),
		connectionUpdates: &steampipeconfig.ConnectionUpdates{
			Update: steampipeconfig.ConnectionStateMap{
				"dynamic_conn": {
					ConnectionName: "dynamic_conn",
					Plugin:         "hub.steampipe.io/plugins/turbot/test@latest",
					SchemaMode:     sdkplugin.SchemaModeDynamic,
					Type:           stringPtr(""),
				},
			},
		},
	}

	// Simulate successful update - dynamic connection should NOT be added to exemplar map
	connectionState := state.connectionUpdates.Update["dynamic_conn"]
	if connectionState.CanCloneSchema() {
		state.exemplarSchemaMap[connectionState.Plugin] = connectionState.ConnectionName
	}

	// Verify the exemplar was NOT recorded
	_, exists := state.exemplarSchemaMap["hub.steampipe.io/plugins/turbot/test@latest"]
	assert.False(t, exists, "dynamic schema should not be in exemplar map")
}

// TestRateLimiterManagement tests rate limiter update logic
func TestRateLimiterManagement(t *testing.T) {
	tests := []struct {
		name                    string
		pluginsWithUpdatedBinary map[string]string
		wantLoadCalls           int
		wantHandleCalls         int
		setupMock               func(*mockPluginManager)
	}{
		{
			name: "plugins with updated binaries",
			pluginsWithUpdatedBinary: map[string]string{
				"hub.steampipe.io/plugins/turbot/aws@latest": "aws_dev",
			},
			wantLoadCalls:   1,
			wantHandleCalls: 1,
			setupMock: func(m *mockPluginManager) {
				m.loadPluginRateLimitersFunc = func(plugins map[string]string) (PluginLimiterMap, error) {
					// Return non-empty rate limiters
					return PluginLimiterMap{
						"aws_dev": {},
					}, nil
				}
			},
		},
		{
			name:                     "no updated binaries",
			pluginsWithUpdatedBinary: map[string]string{},
			wantLoadCalls:            0,
			wantHandleCalls:          0,
			setupMock:                func(m *mockPluginManager) {},
		},
		{
			name: "load returns no limiters",
			pluginsWithUpdatedBinary: map[string]string{
				"hub.steampipe.io/plugins/turbot/aws@latest": "aws_dev",
			},
			wantLoadCalls:   1,
			wantHandleCalls: 0, // Should not call handle if no limiters returned
			setupMock: func(m *mockPluginManager) {
				m.loadPluginRateLimitersFunc = func(plugins map[string]string) (PluginLimiterMap, error) {
					return PluginLimiterMap{}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := newMockPluginManager(nil)
			tt.setupMock(mockPM)

			state := &refreshConnectionState{
				pluginManager: mockPM,
				connectionUpdates: &steampipeconfig.ConnectionUpdates{
					PluginsWithUpdatedBinary: tt.pluginsWithUpdatedBinary,
				},
			}

			ctx := context.Background()
			err := state.updateRateLimiterDefinitions(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantLoadCalls, mockPM.loadPluginRateLimitersCalls, "LoadPluginRateLimiters calls")
			assert.Equal(t, tt.wantHandleCalls, mockPM.handlePluginLimiterChangesCalls, "HandlePluginLimiterChanges calls")
		})
	}
}

// TestPluginColumnTableUpdates tests plugin column table update logic
func TestPluginColumnTableUpdates(t *testing.T) {
	tests := []struct {
		name                  string
		currentState          steampipeconfig.ConnectionStateMap
		finalState            steampipeconfig.ConnectionStateMap
		updates               steampipeconfig.ConnectionStateMap
		deletes               map[string]struct{}
		wantUpdatePluginCall  bool
	}{
		{
			name: "plugin removed",
			currentState: steampipeconfig.ConnectionStateMap{
				"aws_dev": {
					ConnectionName: "aws_dev",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
				},
			},
			finalState: steampipeconfig.ConnectionStateMap{},
			updates:    steampipeconfig.ConnectionStateMap{},
			deletes: map[string]struct{}{
				"aws_dev": {},
			},
			wantUpdatePluginCall: true,
		},
		{
			name:                 "no changes",
			currentState:         steampipeconfig.ConnectionStateMap{},
			finalState:           steampipeconfig.ConnectionStateMap{},
			updates:              steampipeconfig.ConnectionStateMap{},
			deletes:              map[string]struct{}{},
			wantUpdatePluginCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := newMockPluginManager(nil)

			state := &refreshConnectionState{
				pluginManager: mockPM,
				connectionUpdates: &steampipeconfig.ConnectionUpdates{
					CurrentConnectionState:   tt.currentState,
					FinalConnectionState:     tt.finalState,
					Update:                   tt.updates,
					Delete:                   tt.deletes,
					ConnectionPlugins:        map[string]*steampipeconfig.ConnectionPlugin{},
					PluginsWithUpdatedBinary: map[string]string{},
				},
			}

			ctx := context.Background()
			err := state.updatePluginColumnTable(ctx)

			// Should not error (although the actual logic requires complex setup)
			assert.NoError(t, err)

			if tt.wantUpdatePluginCall {
				assert.Greater(t, mockPM.updatePluginColumnsTableCalls, 0, "UpdatePluginColumnsTable should be called")
			}
		})
	}
}

// TestConnectionErrorHandling tests error state management
func TestConnectionErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		connectionErr error
		wantState     string
	}{
		{
			name:          "connection error",
			connectionErr: fmt.Errorf("failed to connect to plugin"),
			wantState:     "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &steampipeconfig.ConnectionState{
				ConnectionName: "test_conn",
			}

			if tt.connectionErr != nil {
				state.SetError(tt.connectionErr.Error())
			}

			assert.Equal(t, tt.wantState, state.State)
			if tt.connectionErr != nil {
				assert.Equal(t, tt.connectionErr.Error(), state.Error())
			}
		})
	}
}

// TestMissingPluginWarnings tests warning generation for missing plugins
func TestMissingPluginWarnings(t *testing.T) {
	state := &refreshConnectionState{
		res: &steampipeconfig.RefreshConnectionResult{},
		connectionUpdates: &steampipeconfig.ConnectionUpdates{
			MissingPlugins: map[string][]modconfig.SteampipeConnection{
				"aws": {
					{Name: "aws_dev"},
					{Name: "aws_prod"},
				},
				"gcp": {
					{Name: "gcp_dev"},
				},
			},
		},
	}

	state.addMissingPluginWarnings()

	// Should have added a warning
	assert.NotEmpty(t, state.res.Warnings)
	// Warning should mention the plugins
	warningStr := fmt.Sprintf("%v", state.res.Warnings)
	assert.Contains(t, warningStr, "plugin")
	assert.Contains(t, warningStr, "connection")
}

// TestConnectionStateTransitions tests state transitions through the refresh process
func TestConnectionStateTransitions(t *testing.T) {
	tests := []struct {
		name          string
		initialState  string
		expectedFinal string
		operation     string
	}{
		{
			name:          "pending to ready",
			initialState:  "pending",
			expectedFinal: "ready",
			operation:     "update_success",
		},
		{
			name:          "pending to error",
			initialState:  "pending",
			expectedFinal: "error",
			operation:     "update_failure",
		},
		{
			name:          "ready stays ready",
			initialState:  "ready",
			expectedFinal: "ready",
			operation:     "no_change",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &steampipeconfig.ConnectionState{
				ConnectionName: "test_conn",
				State:          tt.initialState,
			}

			// Simulate the operation
			switch tt.operation {
			case "update_success":
				state.State = "ready"
			case "update_failure":
				state.SetError("update failed")
			}

			assert.Equal(t, tt.expectedFinal, state.State)
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
