package steampipeconfig

import (
	"testing"

	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

// TestNewConnectionPlugin removed - trivial constructor test
// Only validated field assignment and map initialization
// Documented in cleanup report

func TestConnectionPlugin_AddConnection(t *testing.T) {
	cp := NewConnectionPlugin("test", "test-plugin", nil, nil)

	tests := []struct {
		name           string
		connectionName string
		config         string
		connectionType string
	}{
		{
			name:           "simple_connection",
			connectionName: "conn1",
			config:         `{"key": "value"}`,
			connectionType: "plugin",
		},
		{
			name:           "connection_with_empty_config",
			connectionName: "conn2",
			config:         "",
			connectionType: "plugin",
		},
		{
			name:           "aggregator_connection",
			connectionName: "agg1",
			config:         "",
			connectionType: "aggregator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := len(cp.ConnectionMap)
			cp.addConnection(tt.connectionName, tt.config, tt.connectionType)

			if len(cp.ConnectionMap) != initialCount+1 {
				t.Errorf("Expected ConnectionMap to have %d entries, got %d", initialCount+1, len(cp.ConnectionMap))
			}

			connData, exists := cp.ConnectionMap[tt.connectionName]
			if !exists {
				t.Fatalf("Connection %s not found in ConnectionMap", tt.connectionName)
			}

			if connData.Name != tt.connectionName {
				t.Errorf("Expected connection name %s, got %s", tt.connectionName, connData.Name)
			}
			if connData.Config != tt.config {
				t.Errorf("Expected config %s, got %s", tt.config, connData.Config)
			}
			if connData.Type != tt.connectionType {
				t.Errorf("Expected type %s, got %s", tt.connectionType, connData.Type)
			}
		})
	}
}

// TestConnectionPluginData removed - trivial struct field assignment test
// Only validated that fields matched assigned values
// Documented in cleanup report

func TestHandleGetFailures(t *testing.T) {
	// Create a minimal global config for testing
	GlobalConfig = &SteampipeConfig{
		Connections:      make(map[string]*modconfig.SteampipeConnection),
		PluginsInstances: make(map[string]*plugin.Plugin),
	}

	tests := []struct {
		name               string
		getResponse        *proto.GetResponse
		expectedWarnings   int
		expectedFailedConn int
	}{
		{
			name: "no_failures",
			getResponse: &proto.GetResponse{
				FailureMap: map[string]string{},
			},
			expectedWarnings:   0,
			expectedFailedConn: 0,
		},
		{
			name: "generic_failure",
			getResponse: &proto.GetResponse{
				FailureMap: map[string]string{
					"plugin1": "failed to start",
				},
			},
			expectedWarnings:   1,
			expectedFailedConn: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &RefreshConnectionResult{}
			handleGetFailures(tt.getResponse, res, []*modconfig.SteampipeConnection{})

			if len(res.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(res.Warnings))
			}
		})
	}
}

func TestFullConnectionPluginMap(t *testing.T) {
	// Create connection plugins
	cp1 := NewConnectionPlugin("plugin1", "plugin1-instance", nil, nil)
	cp1.addConnection("conn1", "config1", "plugin")
	cp1.addConnection("conn2", "config2", "plugin")

	cp2 := NewConnectionPlugin("plugin2", "plugin2-instance", nil, nil)
	cp2.addConnection("conn3", "config3", "plugin")

	// Create sparse map (only requesting conn1 and conn3)
	sparseMap := map[string]*ConnectionPlugin{
		"conn1": cp1,
		"conn3": cp2,
	}

	// Get full map
	fullMap := fullConnectionPluginMap(sparseMap)

	// Should have all 3 connections
	if len(fullMap) != 3 {
		t.Errorf("Expected 3 connections in full map, got %d", len(fullMap))
	}

	// Verify all connections are present
	expectedConnections := []string{"conn1", "conn2", "conn3"}
	for _, connName := range expectedConnections {
		if _, exists := fullMap[connName]; !exists {
			t.Errorf("Expected connection %s not found in full map", connName)
		}
	}

	// Verify conn1 and conn2 point to cp1
	if fullMap["conn1"] != cp1 || fullMap["conn2"] != cp1 {
		t.Error("conn1 and conn2 should point to the same plugin")
	}

	// Verify conn3 points to cp2
	if fullMap["conn3"] != cp2 {
		t.Error("conn3 should point to cp2")
	}
}
