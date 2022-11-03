package steampipeconfig

import (
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// ConnectionSchemaMap is a map of connection to all connections with the same schema
type ConnectionSchemaMap map[string][]string

// NewConnectionSchemaMap creates a ConnectionSchemaMap for all configured connections
// it uses the current connection state to determine if a connection has a dynamic schema
// (NOTE: this will no work for newly added plugins which will not have a state yet
// - which is why CreateConnectionPlugins loads the schemas for each new plugin
// and calls NewConnectionSchemaMapForConnections directly, passing the schema modes)
func NewConnectionSchemaMap() (ConnectionSchemaMap, error) {
	connectionNames := GlobalConfig.ConnectionNames()
	connectionState, err := GetConnectionState(connectionNames)
	if err != nil {
		return nil, err
	}

	res := make(ConnectionSchemaMap)

	// if there is only 1 connection, just return a map containing it
	if len(connectionNames) == 1 {
		res[connectionNames[0]] = connectionNames
		return res, nil
	}

	// build a map of connection name to schema mode
	schemaModeMap := make(map[string]string, len(connectionState))
	for connectionName, connectionData := range connectionState {
		if connectionData.SchemaMode != "" {
			schemaModeMap[connectionName] = connectionData.SchemaMode
		}
	}
	// now build the ConnectionSchemaMap
	return NewConnectionSchemaMapForConnections(GlobalConfig.ConnectionList(), schemaModeMap, connectionState), nil

}

func NewConnectionSchemaMapForConnections(connections []*modconfig.Connection, schemaModeMap map[string]string, connectionState ConnectionDataMap) ConnectionSchemaMap {
	var res = make(ConnectionSchemaMap)
	// map of plugin name to first connection which uses it
	pluginMap := make(map[string]string)
	for _, connection := range connections {
		// if this does not exist in state, skip it
		if _, ok := connectionState[connection.Name]; !ok {
			continue
		}

		p := connection.Plugin

		// look for this plugin in the map - read out the first conneciton which uses it
		connectionForPlugin, ok := pluginMap[p]
		// if the plugin does NOT appear in the plugin map,
		// this is the first connection schema that uses this plugin
		thisIsFirstConnectionForPlugin := !ok

		// so determine whether this is a dynamic schema
		dynamicSchema := schemaModeMap[connection.Name] == plugin.SchemaModeDynamic
		shouldAddSchema := thisIsFirstConnectionForPlugin || dynamicSchema

		// if we have not handled this plugin before, or it is a dynamic schema
		if shouldAddSchema {
			pluginMap[p] = connection.Name
			// add a new entry in the schema map
			res[connection.Name] = []string{connection.Name}
		} else {
			// just update list of connections using same schema
			res[connectionForPlugin] = append(res[connectionForPlugin], connection.Name)
		}
	}
	return res
}

// UniqueSchemas returns the unique schemas for all loaded connections
func (c ConnectionSchemaMap) UniqueSchemas() []string {
	res := make([]string, len(c))
	idx := 0
	for c := range c {
		res[idx] = c
		idx++
	}
	return res
}
