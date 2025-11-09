package steampipeconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// TestConnectionStateMapGetSummary tests the GetSummary method
func TestConnectionStateMapGetSummary(t *testing.T) {
	stateMap := ConnectionStateMap{
		"conn1": {State: constants.ConnectionStateReady},
		"conn2": {State: constants.ConnectionStateReady},
		"conn3": {State: constants.ConnectionStateError},
		"conn4": {State: constants.ConnectionStatePending},
		"conn5": {State: constants.ConnectionStateDisabled},
	}

	summary := stateMap.GetSummary()

	assert.Equal(t, 2, summary[constants.ConnectionStateReady])
	assert.Equal(t, 1, summary[constants.ConnectionStateError])
	assert.Equal(t, 1, summary[constants.ConnectionStatePending])
	assert.Equal(t, 1, summary[constants.ConnectionStateDisabled])
}

// TestConnectionStateMapPending tests the Pending method
func TestConnectionStateMapPending(t *testing.T) {
	tests := []struct {
		name        string
		stateMap    ConnectionStateMap
		wantPending bool
	}{
		{
			name: "has pending connections",
			stateMap: ConnectionStateMap{
				"conn1": {State: constants.ConnectionStateReady},
				"conn2": {State: constants.ConnectionStatePending},
			},
			wantPending: true,
		},
		{
			name: "has pending incomplete connections",
			stateMap: ConnectionStateMap{
				"conn1": {State: constants.ConnectionStateReady},
				"conn2": {State: constants.ConnectionStatePendingIncomplete},
			},
			wantPending: true,
		},
		{
			name: "no pending connections",
			stateMap: ConnectionStateMap{
				"conn1": {State: constants.ConnectionStateReady},
				"conn2": {State: constants.ConnectionStateError},
			},
			wantPending: false,
		},
		{
			name:        "empty map",
			stateMap:    ConnectionStateMap{},
			wantPending: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stateMap.Pending()
			assert.Equal(t, tt.wantPending, got)
		})
	}
}

// TestConnectionStateMapLoaded tests the Loaded method
func TestConnectionStateMapLoaded(t *testing.T) {
	tests := []struct {
		name        string
		stateMap    ConnectionStateMap
		checkConns  []string
		wantLoaded  bool
	}{
		{
			name: "all connections loaded",
			stateMap: ConnectionStateMap{
				"conn1": {State: constants.ConnectionStateReady},
				"conn2": {State: constants.ConnectionStateError},
				"conn3": {State: constants.ConnectionStateDisabled},
			},
			checkConns: []string{},
			wantLoaded: true,
		},
		{
			name: "some connections pending",
			stateMap: ConnectionStateMap{
				"conn1": {State: constants.ConnectionStateReady},
				"conn2": {State: constants.ConnectionStatePending},
			},
			checkConns: []string{},
			wantLoaded: false,
		},
		{
			name: "check specific loaded connections",
			stateMap: ConnectionStateMap{
				"conn1": {State: constants.ConnectionStateReady},
				"conn2": {State: constants.ConnectionStatePending},
			},
			checkConns: []string{"conn1"},
			wantLoaded: true,
		},
		{
			name: "check specific pending connections",
			stateMap: ConnectionStateMap{
				"conn1": {State: constants.ConnectionStateReady},
				"conn2": {State: constants.ConnectionStatePending},
			},
			checkConns: []string{"conn2"},
			wantLoaded: false,
		},
		{
			name: "check non-existent connection",
			stateMap: ConnectionStateMap{
				"conn1": {State: constants.ConnectionStateReady},
			},
			checkConns: []string{"conn_nonexistent"},
			wantLoaded: true, // Non-existent connections are ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stateMap.Loaded(tt.checkConns...)
			assert.Equal(t, tt.wantLoaded, got)
		})
	}
}

// TestConnectionStateMapConnectionsInState tests the ConnectionsInState method
func TestConnectionStateMapConnectionsInState(t *testing.T) {
	stateMap := ConnectionStateMap{
		"conn1": {State: constants.ConnectionStateReady},
		"conn2": {State: constants.ConnectionStateError},
		"conn3": {State: constants.ConnectionStatePending},
	}

	tests := []struct {
		name       string
		states     []string
		wantInState bool
	}{
		{
			name:       "has ready connections",
			states:     []string{constants.ConnectionStateReady},
			wantInState: true,
		},
		{
			name:       "has error or pending connections",
			states:     []string{constants.ConnectionStateError, constants.ConnectionStatePending},
			wantInState: true,
		},
		{
			name:       "no updating connections",
			states:     []string{constants.ConnectionStateUpdating},
			wantInState: false,
		},
		{
			name:       "no deleting connections",
			states:     []string{constants.ConnectionStateDeleting},
			wantInState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stateMap.ConnectionsInState(tt.states...)
			assert.Equal(t, tt.wantInState, got)
		})
	}
}

// TestConnectionStateMapEquals tests the Equals method
func TestConnectionStateMapEquals(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name       string
		map1       ConnectionStateMap
		map2       ConnectionStateMap
		wantEquals bool
	}{
		{
			name: "identical maps",
			map1: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					State:          constants.ConnectionStateReady,
					PluginModTime:  baseTime,
				},
			},
			map2: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					State:          constants.ConnectionStateReady,
					PluginModTime:  baseTime,
				},
			},
			wantEquals: true,
		},
		{
			name: "different connection plugins",
			map1: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					State:          constants.ConnectionStateReady,
					ImportSchema:   "enabled",
					PluginModTime:  baseTime,
				},
			},
			map2: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					Plugin:         "hub.steampipe.io/plugins/turbot/gcp@latest",
					State:          constants.ConnectionStateReady,
					ImportSchema:   "enabled",
					PluginModTime:  baseTime,
				},
			},
			wantEquals: false,
		},
		{
			name: "different number of connections",
			map1: ConnectionStateMap{
				"conn1": {ConnectionName: "conn1"},
				"conn2": {ConnectionName: "conn2"},
			},
			map2: ConnectionStateMap{
				"conn1": {ConnectionName: "conn1"},
			},
			wantEquals: false,
		},
		{
			name:       "both empty",
			map1:       ConnectionStateMap{},
			map2:       ConnectionStateMap{},
			wantEquals: true,
		},
		{
			name: "nil vs non-empty",
			map1: nil,
			map2: ConnectionStateMap{
				"conn1": {ConnectionName: "conn1"},
			},
			wantEquals: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.map1.Equals(tt.map2)
			assert.Equal(t, tt.wantEquals, got)
		})
	}
}

// TestConnectionStateMapConnectionModTime tests the ConnectionModTime method
func TestConnectionStateMapConnectionModTime(t *testing.T) {
	baseTime := time.Now()
	laterTime := baseTime.Add(1 * time.Hour)
	latestTime := baseTime.Add(2 * time.Hour)

	stateMap := ConnectionStateMap{
		"conn1": {ConnectionModTime: baseTime},
		"conn2": {ConnectionModTime: laterTime},
		"conn3": {ConnectionModTime: latestTime},
	}

	got := stateMap.ConnectionModTime()
	assert.Equal(t, latestTime, got)
}

// TestConnectionStateMapGetFirstSearchPathConnectionForPlugins tests getting first search path connections
func TestConnectionStateMapGetFirstSearchPathConnectionForPlugins(t *testing.T) {
	staticType := ""

	tests := []struct {
		name       string
		stateMap   ConnectionStateMap
		searchPath []string
		wantConns  int // Number of connections expected
	}{
		{
			name: "static plugins - first in search path only",
			stateMap: ConnectionStateMap{
				"aws_dev": {
					ConnectionName: "aws_dev",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     plugin.SchemaModeStatic,
					Type:           &staticType,
				},
				"aws_prod": {
					ConnectionName: "aws_prod",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     plugin.SchemaModeStatic,
					Type:           &staticType,
				},
				"gcp_dev": {
					ConnectionName: "gcp_dev",
					Plugin:         "hub.steampipe.io/plugins/turbot/gcp@latest",
					SchemaMode:     plugin.SchemaModeStatic,
					Type:           &staticType,
				},
			},
			searchPath: []string{"aws_dev", "aws_prod", "gcp_dev"},
			wantConns:  2, // aws_dev (first AWS) and gcp_dev (first GCP)
		},
		{
			name: "dynamic plugins - all in search path",
			stateMap: ConnectionStateMap{
				"dynamic1": {
					ConnectionName: "dynamic1",
					Plugin:         "hub.steampipe.io/plugins/turbot/test@latest",
					SchemaMode:     plugin.SchemaModeDynamic,
					Type:           &staticType,
				},
				"dynamic2": {
					ConnectionName: "dynamic2",
					Plugin:         "hub.steampipe.io/plugins/turbot/test@latest",
					SchemaMode:     plugin.SchemaModeDynamic,
					Type:           &staticType,
				},
			},
			searchPath: []string{"dynamic1", "dynamic2"},
			wantConns:  2, // Both dynamic connections included
		},
		{
			name: "skip disabled connections",
			stateMap: ConnectionStateMap{
				"conn1": {
					ConnectionName: "conn1",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     plugin.SchemaModeStatic,
					State:          constants.ConnectionStateDisabled,
					Type:           &staticType,
				},
				"conn2": {
					ConnectionName: "conn2",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					SchemaMode:     plugin.SchemaModeStatic,
					State:          constants.ConnectionStateReady,
					Type:           &staticType,
				},
			},
			searchPath: []string{"conn1", "conn2"},
			wantConns:  1, // Only conn2 (conn1 is disabled)
		},
		{
			name:       "empty search path",
			stateMap:   ConnectionStateMap{},
			searchPath: []string{},
			wantConns:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stateMap.GetFirstSearchPathConnectionForPlugins(tt.searchPath)
			assert.Len(t, got, tt.wantConns)
		})
	}
}

// TestConnectionStateMapGetPluginToConnectionMap tests plugin to connection mapping
func TestConnectionStateMapGetPluginToConnectionMap(t *testing.T) {
	staticType := ""

	stateMap := ConnectionStateMap{
		"aws_dev": {
			ConnectionName: "aws_dev",
			Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
			Type:           &staticType,
		},
		"aws_prod": {
			ConnectionName: "aws_prod",
			Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
			Type:           &staticType,
		},
		"gcp_dev": {
			ConnectionName: "gcp_dev",
			Plugin:         "hub.steampipe.io/plugins/turbot/gcp@latest",
			Type:           &staticType,
		},
	}

	pluginMap := stateMap.GetPluginToConnectionMap()

	assert.Len(t, pluginMap, 2) // Two plugins
	assert.Len(t, pluginMap["hub.steampipe.io/plugins/turbot/aws@latest"], 2) // Two AWS connections
	assert.Len(t, pluginMap["hub.steampipe.io/plugins/turbot/gcp@latest"], 1) // One GCP connection
	assert.Contains(t, pluginMap["hub.steampipe.io/plugins/turbot/aws@latest"], "aws_dev")
	assert.Contains(t, pluginMap["hub.steampipe.io/plugins/turbot/aws@latest"], "aws_prod")
	assert.Contains(t, pluginMap["hub.steampipe.io/plugins/turbot/gcp@latest"], "gcp_dev")
}

// TestConnectionStateMapSetConnectionsToPendingOrIncomplete tests state transitions
func TestConnectionStateMapSetConnectionsToPendingOrIncomplete(t *testing.T) {
	beforeTime := time.Now()

	stateMap := ConnectionStateMap{
		"ready_conn": {
			ConnectionName:    "ready_conn",
			State:             constants.ConnectionStateReady,
			ConnectionModTime: beforeTime,
		},
		"error_conn": {
			ConnectionName:    "error_conn",
			State:             constants.ConnectionStateError,
			ConnectionModTime: beforeTime,
		},
		"disabled_conn": {
			ConnectionName:    "disabled_conn",
			State:             constants.ConnectionStateDisabled,
			ConnectionModTime: beforeTime,
		},
	}

	stateMap.SetConnectionsToPendingOrIncomplete()

	// Ready connections should become pending
	assert.Equal(t, constants.ConnectionStatePending, stateMap["ready_conn"].State)
	assert.True(t, stateMap["ready_conn"].ConnectionModTime.After(beforeTime))

	// Error connections should become pending incomplete
	assert.Equal(t, constants.ConnectionStatePendingIncomplete, stateMap["error_conn"].State)
	assert.True(t, stateMap["error_conn"].ConnectionModTime.After(beforeTime))

	// Disabled connections should stay disabled
	assert.Equal(t, constants.ConnectionStateDisabled, stateMap["disabled_conn"].State)
}

// TestGetRequiredConnectionStateMap tests building required connection state
func TestGetRequiredConnectionStateMap(t *testing.T) {
	tests := []struct {
		name                   string
		connectionMap          map[string]*modconfig.SteampipeConnection
		currentConnectionState ConnectionStateMap
		wantCount              int
		wantMissingPlugins     int
		wantError              bool
	}{
		{
			name: "connection with error",
			connectionMap: map[string]*modconfig.SteampipeConnection{
				"error_conn": {
					Name:         "error_conn",
					Plugin:       "hub.steampipe.io/plugins/turbot/aws@latest",
					PluginAlias:  "aws",
					ImportSchema: "enabled",
					Error:        assert.AnError,
				},
			},
			currentConnectionState: ConnectionStateMap{},
			wantCount:              1,
			wantMissingPlugins:     0,
			wantError:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, missingPlugins, result := GetRequiredConnectionStateMap(tt.connectionMap, tt.currentConnectionState)

			// Should return result even if there are warnings
			assert.NotNil(t, got)
			assert.Len(t, got, tt.wantCount)
			assert.Len(t, missingPlugins, tt.wantMissingPlugins)

			// Check that connections are in the expected state
			for name, conn := range tt.connectionMap {
				state, ok := got[name]
				assert.True(t, ok, "connection %s should be in result", name)
				if state != nil {
					assert.Equal(t, conn.Name, state.ConnectionName)
					assert.Equal(t, conn.Plugin, state.Plugin)

					// Connections with errors should have error state
					if conn.Error != nil {
						assert.Equal(t, constants.ConnectionStateError, state.State)
					}
				}
			}

			// Check error expectation
			if tt.wantError {
				assert.Error(t, result.Error)
			}
		})
	}
}
