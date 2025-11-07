package helpers

import (
	"testing"

	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// NewTestConfig creates a minimal valid config for testing
func NewTestConfig() *steampipeconfig.SteampipeConfig {
	return steampipeconfig.NewSteampipeConfig("")
}

// NewTestConfigWithConnection creates a test config with a single connection
func NewTestConfigWithConnection(t *testing.T, connectionName string) *steampipeconfig.SteampipeConfig {
	t.Helper()

	config := NewTestConfig()
	connection := NewTestConnection(connectionName)
	config.Connections[connectionName] = connection

	return config
}

// NewTestConnection creates a test connection
func NewTestConnection(name string) *modconfig.SteampipeConnection {
	return &modconfig.SteampipeConnection{
		Name:         name,
		PluginAlias:  "test",
		Plugin:       "test",
		Type:         "",         // empty for regular connection
		ImportSchema: "enabled",
		Connections:  make(map[string]*modconfig.SteampipeConnection),
	}
}

// NewTestAggregatorConnection creates a test aggregator connection
func NewTestAggregatorConnection(name string, childConnectionNames []string) *modconfig.SteampipeConnection {
	return &modconfig.SteampipeConnection{
		Name:            name,
		PluginAlias:     "test",
		Plugin:          "test",
		Type:            "aggregator",
		ImportSchema:    "enabled",
		ConnectionNames: childConnectionNames,
		Connections:     make(map[string]*modconfig.SteampipeConnection),
	}
}

// AddConnectionToConfig adds a connection to an existing config
func AddConnectionToConfig(config *steampipeconfig.SteampipeConfig, connection *modconfig.SteampipeConnection) {
	config.Connections[connection.Name] = connection
}
