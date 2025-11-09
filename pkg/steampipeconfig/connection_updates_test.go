package steampipeconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// TestConnectionUpdatesHasUpdates tests the HasUpdates method
func TestConnectionUpdatesHasUpdates(t *testing.T) {
	tests := []struct {
		name           string
		updates        *ConnectionUpdates
		wantHasUpdates bool
	}{
		{
			name: "has updates",
			updates: &ConnectionUpdates{
				Update: ConnectionStateMap{
					"conn1": {},
				},
				Delete:          map[string]struct{}{},
				MissingComments: ConnectionStateMap{},
			},
			wantHasUpdates: true,
		},
		{
			name: "has deletes",
			updates: &ConnectionUpdates{
				Update: ConnectionStateMap{},
				Delete: map[string]struct{}{
					"conn1": {},
				},
				MissingComments: ConnectionStateMap{},
			},
			wantHasUpdates: true,
		},
		{
			name: "has missing comments",
			updates: &ConnectionUpdates{
				Update: ConnectionStateMap{},
				Delete: map[string]struct{}{},
				MissingComments: ConnectionStateMap{
					"conn1": {},
				},
			},
			wantHasUpdates: true,
		},
		{
			name: "no updates",
			updates: &ConnectionUpdates{
				Update:          ConnectionStateMap{},
				Delete:          map[string]struct{}{},
				MissingComments: ConnectionStateMap{},
			},
			wantHasUpdates: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.updates.HasUpdates()
			assert.Equal(t, tt.wantHasUpdates, got)
		})
	}
}

// TestConnectionUpdatesSetError tests setting connection errors
func TestConnectionUpdatesSetError(t *testing.T) {
	updates := &ConnectionUpdates{
		FinalConnectionState: ConnectionStateMap{
			"test_conn": {
				ConnectionName: "test_conn",
				State:          constants.ConnectionStateReady,
			},
		},
		Update: ConnectionStateMap{
			"test_conn": {
				ConnectionName: "test_conn",
			},
		},
	}

	errorMsg := "connection failed"
	updates.setError("test_conn", errorMsg)

	// Check that connection is now in error state
	assert.Equal(t, constants.ConnectionStateError, updates.FinalConnectionState["test_conn"].State)
	assert.Equal(t, errorMsg, updates.FinalConnectionState["test_conn"].Error())

	// Check that connection is removed from updates
	_, inUpdates := updates.Update["test_conn"]
	assert.False(t, inUpdates)
}

// TestConnectionUpdatesIdentifyMissingComments tests identifying connections with missing comments
func TestConnectionUpdatesIdentifyMissingComments(t *testing.T) {
	tests := []struct {
		name               string
		currentState       ConnectionStateMap
		finalState         ConnectionStateMap
		updates            ConnectionStateMap
		deletes            map[string]struct{}
		wantMissingCount   int
	}{
		{
			name: "connection with missing comments",
			currentState: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
					CommentsSet:    false,
				},
			},
			finalState: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
				},
			},
			updates:          ConnectionStateMap{},
			deletes:          map[string]struct{}{},
			wantMissingCount: 1,
		},
		{
			name: "connection with comments set",
			currentState: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
					CommentsSet:    true,
				},
			},
			finalState: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
				},
			},
			updates:          ConnectionStateMap{},
			deletes:          map[string]struct{}{},
			wantMissingCount: 0,
		},
		{
			name: "skip connection being updated",
			currentState: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
					CommentsSet:    false,
				},
			},
			finalState: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
				},
			},
			updates: ConnectionStateMap{
				"conn1": {},
			},
			deletes:          map[string]struct{}{},
			wantMissingCount: 0, // Should skip because it's being updated
		},
		{
			name: "connection being deleted is marked as missing",
			currentState: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
					CommentsSet:    false,
				},
			},
			finalState: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
				},
			},
			updates: ConnectionStateMap{},
			deletes: map[string]struct{}{
				"conn1": {},
			},
			wantMissingCount: 1, // Will be marked as missing because deleting is true (logic: !updating || deleting)
		},
		{
			name: "skip connection in error state",
			currentState: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
					CommentsSet:    false,
				},
			},
			finalState: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					State:          constants.ConnectionStateError,
				},
			},
			updates:          ConnectionStateMap{},
			deletes:          map[string]struct{}{},
			wantMissingCount: 0, // Should skip because it's in error state
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updates := &ConnectionUpdates{
				CurrentConnectionState: tt.currentState,
				FinalConnectionState:   tt.finalState,
				Update:                 tt.updates,
				Delete:                 tt.deletes,
				MissingComments:        ConnectionStateMap{},
			}

			updates.IdentifyMissingComments()

			assert.Len(t, updates.MissingComments, tt.wantMissingCount)
		})
	}
}

// TestConnectionUpdatesDynamicUpdates tests getting dynamic plugin updates
func TestConnectionUpdatesDynamicUpdates(t *testing.T) {
	updates := &ConnectionUpdates{
		Update: ConnectionStateMap{
			"static_conn": {
				ConnectionName: "static_conn",
				SchemaMode:     plugin.SchemaModeStatic,
			},
			"dynamic_conn1": {
				ConnectionName: "dynamic_conn1",
				SchemaMode:     plugin.SchemaModeDynamic,
			},
			"dynamic_conn2": {
				ConnectionName: "dynamic_conn2",
				SchemaMode:     plugin.SchemaModeDynamic,
			},
		},
	}

	dynamicUpdates := updates.DynamicUpdates()

	assert.Len(t, dynamicUpdates, 2)
	assert.Contains(t, dynamicUpdates, "dynamic_conn1")
	assert.Contains(t, dynamicUpdates, "dynamic_conn2")
	assert.NotContains(t, dynamicUpdates, "static_conn")
}

// TestConnectionUpdatesGetConnectionsToDelete tests getting all connections to delete
func TestConnectionUpdatesGetConnectionsToDelete(t *testing.T) {
	updates := &ConnectionUpdates{
		Delete: map[string]struct{}{
			"delete1": {},
			"delete2": {},
		},
		Error: map[string]struct{}{
			"error1": {},
		},
	}

	toDelete := updates.GetConnectionsToDelete()

	assert.Len(t, toDelete, 3)
	assert.Contains(t, toDelete, "delete1")
	assert.Contains(t, toDelete, "delete2")
	assert.Contains(t, toDelete, "error1")
}

// TestConnectionRequiresUpdate tests the logic for determining if a connection needs updating
func TestConnectionRequiresUpdate(t *testing.T) {
	baseTime := time.Now()
	laterTime := baseTime.Add(1 * time.Hour)

	tests := []struct {
		name                    string
		forceUpdateNames        []string
		connectionName          string
		currentState            ConnectionStateMap
		requiredState           *ConnectionState
		wantRequiresUpdate      bool
		wantPluginBinaryChanged bool
	}{
		{
			name:             "new connection",
			connectionName:   "new_conn",
			currentState:     ConnectionStateMap{},
			requiredState: &ConnectionState{
				ConnectionName: "new_conn",
				State:          constants.ConnectionStateReady,
			},
			wantRequiresUpdate:      true,
			wantPluginBinaryChanged: false,
		},
		{
			name:           "plugin binary changed",
			connectionName: "test_conn",
			currentState: ConnectionStateMap{
				"test_conn": {
					ConnectionName: "test_conn",
					PluginModTime:  baseTime,
				},
			},
			requiredState: &ConnectionState{
				ConnectionName: "test_conn",
				PluginModTime:  laterTime,
			},
			wantRequiresUpdate:      true,
			wantPluginBinaryChanged: true,
		},
		{
			name:           "connection previously disabled, now enabled",
			connectionName: "test_conn",
			currentState: ConnectionStateMap{
				"test_conn": {
					ConnectionName: "test_conn",
					State:          constants.ConnectionStateDisabled,
					PluginModTime:  baseTime,
				},
			},
			requiredState: &ConnectionState{
				ConnectionName: "test_conn",
				State:          constants.ConnectionStateReady,
				PluginModTime:  baseTime,
			},
			wantRequiresUpdate:      true,
			wantPluginBinaryChanged: false,
		},
		{
			name:             "forced update",
			forceUpdateNames: []string{"test_conn"},
			connectionName:   "test_conn",
			currentState: ConnectionStateMap{
				"test_conn": {
					ConnectionName: "test_conn",
					State:          constants.ConnectionStateReady,
					PluginModTime:  baseTime,
				},
			},
			requiredState: &ConnectionState{
				ConnectionName: "test_conn",
				State:          constants.ConnectionStateReady,
				PluginModTime:  baseTime,
			},
			wantRequiresUpdate:      true,
			wantPluginBinaryChanged: false,
		},
		{
			name:           "connection previously incomplete",
			connectionName: "test_conn",
			currentState: ConnectionStateMap{
				"test_conn": {
					ConnectionName: "test_conn",
					State:          constants.ConnectionStatePendingIncomplete,
					PluginModTime:  baseTime,
				},
			},
			requiredState: &ConnectionState{
				ConnectionName: "test_conn",
				State:          constants.ConnectionStateReady,
				PluginModTime:  baseTime,
			},
			wantRequiresUpdate:      true,
			wantPluginBinaryChanged: false,
		},
		{
			name:           "no changes needed",
			connectionName: "test_conn",
			currentState: ConnectionStateMap{
				"test_conn": {
					ConnectionName: "test_conn",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					State:          constants.ConnectionStateReady,
					ImportSchema:   "enabled",
					PluginModTime:  baseTime,
				},
			},
			requiredState: &ConnectionState{
				ConnectionName: "test_conn",
				Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
				State:          constants.ConnectionStateReady,
				ImportSchema:   "enabled",
				PluginModTime:  baseTime,
			},
			wantRequiresUpdate:      false,
			wantPluginBinaryChanged: false,
		},
		{
			name:           "required connection in error state",
			connectionName: "test_conn",
			currentState: ConnectionStateMap{
				"test_conn": {
					ConnectionName: "test_conn",
					State:          constants.ConnectionStateReady,
				},
			},
			requiredState: &ConnectionState{
				ConnectionName: "test_conn",
				State:          constants.ConnectionStateError,
			},
			wantRequiresUpdate:      false,
			wantPluginBinaryChanged: false,
		},
		{
			name:           "required connection disabled",
			connectionName: "test_conn",
			currentState: ConnectionStateMap{
				"test_conn": {
					ConnectionName: "test_conn",
					State:          constants.ConnectionStateReady,
				},
			},
			requiredState: &ConnectionState{
				ConnectionName: "test_conn",
				State:          constants.ConnectionStateDisabled,
			},
			wantRequiresUpdate:      false,
			wantPluginBinaryChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := connectionRequiresUpdate(tt.forceUpdateNames, tt.connectionName, tt.currentState, tt.requiredState)

			assert.Equal(t, tt.wantRequiresUpdate, result.requiresUpdate, "requiresUpdate mismatch")
			assert.Equal(t, tt.wantPluginBinaryChanged, result.pluginBinaryChanged, "pluginBinaryChanged mismatch")
		})
	}
}

// TestConnectionUpdatesPopulateAggregators tests populating aggregator updates
func TestConnectionUpdatesPopulateAggregators(t *testing.T) {
	tests := []struct {
		name                string
		updates             *ConnectionUpdates
		wantAggregatorCount int
	}{
		{
			name: "update aggregator when child updated",
			updates: &ConnectionUpdates{
				Update: ConnectionStateMap{
					"aws_dev": {
						ConnectionName: "aws_dev",
						Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					},
				},
				Delete: map[string]struct{}{},
				FinalConnectionState: ConnectionStateMap{
					"aws_dev": {
						ConnectionName: "aws_dev",
						Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
						Type:           stringPtr(""),
					},
					"all_aws": {
						ConnectionName: "all_aws",
						Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
						Type:           stringPtr(modconfig.ConnectionTypeAggregator),
					},
				},
			},
			wantAggregatorCount: 1, // all_aws should be added to updates
		},
		{
			name: "update aggregator when child deleted",
			updates: &ConnectionUpdates{
				Update: ConnectionStateMap{},
				Delete: map[string]struct{}{
					"aws_dev": {},
				},
				CurrentConnectionState: ConnectionStateMap{
					"aws_dev": {
						ConnectionName: "aws_dev",
						Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					},
				},
				FinalConnectionState: ConnectionStateMap{
					"all_aws": {
						ConnectionName: "all_aws",
						Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
						Type:           stringPtr(modconfig.ConnectionTypeAggregator),
					},
				},
			},
			wantAggregatorCount: 1, // all_aws should be added to updates
		},
		{
			name: "no aggregators to update",
			updates: &ConnectionUpdates{
				Update: ConnectionStateMap{
					"aws_dev": {
						ConnectionName: "aws_dev",
						Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					},
				},
				Delete: map[string]struct{}{},
				FinalConnectionState: ConnectionStateMap{
					"aws_dev": {
						ConnectionName: "aws_dev",
						Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
						Type:           stringPtr(""),
					},
				},
			},
			wantAggregatorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialUpdateCount := len(tt.updates.Update)
			tt.updates.populateAggregators()

			aggregatorCount := 0
			for _, state := range tt.updates.Update {
				if state.GetType() == modconfig.ConnectionTypeAggregator {
					aggregatorCount++
				}
			}

			// Check that we added the expected number of aggregators
			expectedTotal := initialUpdateCount + tt.wantAggregatorCount
			assert.Equal(t, expectedTotal, len(tt.updates.Update))
			assert.Equal(t, tt.wantAggregatorCount, aggregatorCount)
		})
	}
}

// TestConnectionUpdatesUpdateRequiredStateWithSchemaProperties tests schema property updates
func TestConnectionUpdatesUpdateRequiredStateWithSchemaProperties(t *testing.T) {
	tests := []struct {
		name              string
		updates           *ConnectionUpdates
		dynamicSchemaHash map[string]string
		wantSchemaMode    string
		wantSchemaHash    string
	}{
		{
			name: "update from current state",
			updates: &ConnectionUpdates{
				CurrentConnectionState: ConnectionStateMap{
					"conn1": {
						SchemaMode: plugin.SchemaModeStatic,
						SchemaHash: "abc123",
					},
				},
				FinalConnectionState: ConnectionStateMap{
					"conn1": {},
				},
				ConnectionPlugins: map[string]*ConnectionPlugin{},
			},
			dynamicSchemaHash: map[string]string{},
			wantSchemaMode:    plugin.SchemaModeStatic,
			wantSchemaHash:    "abc123",
		},
		{
			name: "update from dynamic schema hash map",
			updates: &ConnectionUpdates{
				CurrentConnectionState: ConnectionStateMap{},
				FinalConnectionState: ConnectionStateMap{
					"conn1": {},
				},
				ConnectionPlugins: map[string]*ConnectionPlugin{},
			},
			dynamicSchemaHash: map[string]string{
				"conn1": "def456",
			},
			wantSchemaMode: "",
			wantSchemaHash: "def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.updates.updateRequiredStateWithSchemaProperties(tt.dynamicSchemaHash)

			for connName, state := range tt.updates.FinalConnectionState {
				if tt.wantSchemaMode != "" {
					assert.Equal(t, tt.wantSchemaMode, state.SchemaMode, "schema mode for %s", connName)
				}
				if tt.wantSchemaHash != "" {
					assert.Equal(t, tt.wantSchemaHash, state.SchemaHash, "schema hash for %s", connName)
				}
			}
		})
	}
}

// TestConnectionUpdatesString tests the string representation
func TestConnectionUpdatesString(t *testing.T) {
	updates := &ConnectionUpdates{
		Update: ConnectionStateMap{
			"conn1": {},
			"conn2": {},
		},
		Delete: map[string]struct{}{
			"conn3": {},
		},
		FinalConnectionState: ConnectionStateMap{
			"conn1": {},
			"conn2": {},
		},
	}

	str := updates.String()

	// String should contain update and delete information
	assert.Contains(t, str, "Update:")
	assert.Contains(t, str, "Delete:")
	assert.Contains(t, str, "Connection state:")
}
