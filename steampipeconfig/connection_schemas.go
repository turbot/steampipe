package steampipeconfig

import "github.com/turbot/steampipe-plugin-sdk/plugin"

type ConnectionSchemas struct {
	// map of connection to all connections with the same schema
	SchemaMap map[string][]string
}

func NewConnectionSchemas(connectionNames []string, connectionState ConnectionDataMap) (*ConnectionSchemas, error) {
	res := &ConnectionSchemas{
		SchemaMap: make(map[string][]string),
	}

	// if there is only 1 connection, just return a map containing it
	if len(connectionNames) == 1 {
		res.SchemaMap[connectionNames[0]] = connectionNames
		return res, nil
	}

	// map of plugin name to first connection which uses it
	pluginMap := make(map[string]string)

	for _, connectionName := range connectionNames {
		connection, ok := GlobalConfig.Connections[connectionName]
		if !ok {
			continue
		}

		p := connection.Plugin

		// look for this plugin in the map - read out the first conneciton which uses it
		connectionForPlugin, ok := pluginMap[p]
		// if the plugin does NOT appear in the plugin map,
		// this is the first connection schema that uses this plugin
		thisIsFirstConnectionForPlugin := !ok

		// do we have a state for this connection - if so determine whether this is a dynamic schema
		var connectionData *ConnectionData
		if connectionState != nil {
			connectionData = connectionState[connectionName]
		}
		dynamicSchema := connectionData != nil && connectionData.SchemaMode == plugin.SchemaModeDynamic
		shouldAddSchema := thisIsFirstConnectionForPlugin || dynamicSchema

		// if we have not handled this plugin before, or it is a dynamic schema
		if shouldAddSchema {
			pluginMap[p] = connectionName
			// add a new entry in the schema map
			res.SchemaMap[connectionName] = []string{connectionName}
		} else {
			// just update list of connections using same schema
			res.SchemaMap[connectionForPlugin] = append(res.SchemaMap[connectionForPlugin], connectionName)
		}
	}
	return res, nil
}

// UniqueSchemas returns the unique schemas for all loaded connections
func (c *ConnectionSchemas) UniqueSchemas() []string {
	res := make([]string, len(c.SchemaMap))
	idx := 0
	for c := range c.SchemaMap {
		res[idx] = c
		idx++
	}
	return res
}
