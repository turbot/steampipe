package connection

import (
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type ConnectionConfigMap map[string]*sdkproto.ConnectionConfig

func NewConnectionConfigMap(connectionMap map[string]*modconfig.Connection) ConnectionConfigMap {
	configMap := make(ConnectionConfigMap)
	for k, v := range connectionMap {
		configMap[k] = &sdkproto.ConnectionConfig{
			Connection:       v.Name,
			Plugin:           v.Plugin,
			PluginShortName:  v.PluginAlias,
			Config:           v.Config,
			ChildConnections: v.GetResolveConnectionNames(),
			PluginInstance:   v.PluginInstance,
		}
	}

	return configMap
}

func (m ConnectionConfigMap) Diff(otherMap ConnectionConfigMap) (addedConnections, deletedConnections, changedConnections map[string][]*sdkproto.ConnectionConfig) {
	// results are maps of connections keyed by plugin label
	addedConnections = make(map[string][]*sdkproto.ConnectionConfig)
	deletedConnections = make(map[string][]*sdkproto.ConnectionConfig)
	changedConnections = make(map[string][]*sdkproto.ConnectionConfig)

	for name, connection := range m {
		if otherConnection, ok := otherMap[name]; !ok {
			deletedConnections[connection.PluginInstance] = append(deletedConnections[connection.PluginInstance], connection)
		} else {
			// check for changes

			// special case - if the plugin has changed, treat this as a deletion and a re-add
			if connection.PluginInstance != otherConnection.Plugin {
				addedConnections[otherConnection.Plugin] = append(addedConnections[otherConnection.Plugin], otherConnection)
				deletedConnections[connection.PluginInstance] = append(deletedConnections[connection.PluginInstance], connection)
			} else {
				if !connection.Equals(otherConnection) {
					changedConnections[connection.PluginInstance] = append(changedConnections[connection.PluginInstance], otherConnection)
				}
			}
		}
	}

	for otherName, otherConnection := range otherMap {
		if _, ok := m[otherName]; !ok {
			addedConnections[otherConnection.Plugin] = append(addedConnections[otherConnection.Plugin], otherConnection)
		}
	}

	return
}
