package steampipeconfig

import "github.com/turbot/steampipe-plugin-sdk/plugin"

type ConnectionSchemas struct {
	// map of connection to all connections with the same schema
	SchemaMap map[string][]string
}

func NewConnectionSchemas() (*ConnectionSchemas, error) {
	res := &ConnectionSchemas{
		SchemaMap: make(map[string][]string),
	}

	// get list of all connection names
	allConnections := GlobalConfig.ConnectionNames()
	// load the connection state file
	connectionState, err := GetConnectionState(allConnections)
	if err != nil {
		return nil, err
	}

	// map of plugin name to first conneciton which uses it
	pluginMap := make(map[string]string)
	for name, connectionData := range connectionState {
		// look for this plugin in the map - read out th efirst conneciton which uses it
		connectionForPlugin, ok := pluginMap[connectionData.Plugin]
		// if the plugin does NOT appear in the plugin map,
		// this is the first connection schema that uses this plugin
		thisIsFirstConnectionForPlugin := !ok
		dynamicSchema := connectionData.SchemaMode == plugin.SchemaModeDynamic
		shouldAddSchema := thisIsFirstConnectionForPlugin || dynamicSchema
		// TODO think about/test bootstrapping
		// if we have not handled this plugin before, or it is a dynamic schema
		if shouldAddSchema {
			pluginMap[connectionData.Plugin] = name
			// add a new entry in the schema map
			res.SchemaMap[name] = []string{name}
		} else {
			// just update list of connections using same schema
			res.SchemaMap[connectionForPlugin] = append(res.SchemaMap[connectionForPlugin], name)
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
