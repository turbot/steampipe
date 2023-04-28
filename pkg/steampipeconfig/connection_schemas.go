package steampipeconfig

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/statushooks"
	"golang.org/x/exp/maps"
)

// ConnectionSchemaMap is a map of connection to all connections with the same schema
type ConnectionSchemaMap map[string][]string

// NewConnectionSchemaMap creates a ConnectionSchemaMap for all configured connections
// it uses the current connection state to determine if a connection has a dynamic schema
// (NOTE: this will no work for newly added plugins which will not have a state yet
// - which is why CreateConnectionPlugins loads the schemas for each new plugin
// and calls NewConnectionSchemaMapForConnections directly, passing the schema modes)
func NewConnectionSchemaMap(ctx context.Context, conn *pgx.Conn) (ConnectionSchemaMap, error) {
	statushooks.SetStatus(ctx, "Loading connection state...")

	connectionStateMap, err := LoadConnectionState(ctx, conn)
	if err != nil {
		return nil, err
	}
	var res = make(ConnectionSchemaMap)

	//if there is only 1 connection, just return a map containing it
	if len(connectionStateMap) == 1 {
		for connectionName := range connectionStateMap {
			res[connectionName] = []string{connectionName}
		}
		return res, nil
	}

	// build a map of connection name to schema mode
	schemaModeMap := make(map[string]string, len(connectionStateMap))
	for connectionName, connectionData := range connectionStateMap {
		schemaModeMap[connectionName] = connectionData.SchemaMode
	}

	// map of plugin name to first connection which uses it
	pluginMap := make(map[string]string)
	for connectionName, connectionState := range connectionStateMap {
		p := connectionState.Plugin

		// look for this plugin in the map - read out the first conneciton which uses it
		connectionForPlugin, ok := pluginMap[p]
		// if the plugin does NOT appear in the plugin map,
		// this is the first connection schema that uses this plugin
		thisIsFirstConnectionForPlugin := !ok

		// so determine whether this is a dynamic schema
		dynamicSchema := schemaModeMap[connectionName] == plugin.SchemaModeDynamic
		shouldAddSchema := thisIsFirstConnectionForPlugin || dynamicSchema

		// if we have not handled this plugin before, or it is a dynamic schema
		if shouldAddSchema {
			pluginMap[p] = connectionName
			// add a new entry in the schema map
			res[connectionName] = []string{connectionName}
		} else {
			// just update list of connections using same schema
			res[connectionForPlugin] = append(res[connectionForPlugin], connectionName)
		}
	}
	return res, err
}

// UniqueSchemas returns the unique schemas for all loaded connections
func (c ConnectionSchemaMap) UniqueSchemas() []string {
	return maps.Keys(c)
}
