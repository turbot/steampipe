package steampipeconfig

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
)

// ConnectionSchemaMap is a map of connection to all connections with the same schema
// key is exemplar connection and value is all connections with same schema
type ConnectionSchemaMap map[string][]string

// NewConnectionSchemaMap creates a ConnectionSchemaMap for all configured connections
// this is a map keyed by exemplar connection with the value the connections which have the same schema
// it uses the current connection state to determine if a connection has a dynamic schema
func NewConnectionSchemaMap(ctx context.Context, connectionStateMap ConnectionStateMap, searchPath []string) ConnectionSchemaMap {
	statushooks.SetStatus(ctx, "Loading connection state…")

	// res is a map of exemplar connections to all the connections with the same schema
	var res = make(ConnectionSchemaMap)

	//if there is only 1 connection, just return a map containing it
	if len(connectionStateMap) == 1 {
		for connectionName := range connectionStateMap {
			res[connectionName] = []string{connectionName}
		}
		return res
	}

	// ask the connection state for the first search path connection for each plugin
	firstConnections := connectionStateMap.GetFirstSearchPathConnectionForPlugins(searchPath)

	// map of plugin name to first connection which uses it
	pluginMap := connectionStateMap.GetPluginToConnectionMap()

	for _, exemplarConnectionName := range firstConnections {
		exemplarConnectionState := connectionStateMap[exemplarConnectionName]
		// if this is a dynamic schema, there will be no connections with the same schema
		if exemplarConnectionState.SchemaMode == plugin.SchemaModeDynamic {
			res[exemplarConnectionName] = nil
		} else {
			var connectionsWithSameSchema []string
			// add all connections for this plugin (apart from exemplar)
			for _, connectionForPlugin := range pluginMap[exemplarConnectionState.Plugin] {
				// do not copy exemplar
				if connectionForPlugin == exemplarConnectionName {
					continue
				}
				connectionState := connectionStateMap[connectionForPlugin]
				// do not include disabled connections
				if connectionState.Disabled() {
					continue
				}
				// otherwise add to list
				connectionsWithSameSchema = append(connectionsWithSameSchema, connectionForPlugin)
			}
			res[exemplarConnectionName] = connectionsWithSameSchema
		}
	}

	return res
}
