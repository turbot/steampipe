package connection

import (
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

type ConnectionConfigMap map[string]*sdkproto.ConnectionConfig

// NewConnectionConfigMap creates a map of sdkproto.ConnectionConfig keyed by connection name
// NOTE: connections in error are EXCLUDED
func NewConnectionConfigMap(connectionMap map[string]*modconfig.SteampipeConnection) ConnectionConfigMap {
	configMap := make(ConnectionConfigMap)
	for k, v := range connectionMap {
		if v.Error != nil {
			continue
		}

		configMap[k] = &sdkproto.ConnectionConfig{
			Connection:       v.Name,
			Plugin:           v.Plugin,
			PluginShortName:  v.PluginAlias,
			Config:           v.Config,
			ChildConnections: v.GetResolveConnectionNames(),
			PluginInstance:   typehelpers.SafeString(v.PluginInstance),
		}
	}

	return configMap
}

func (m ConnectionConfigMap) Diff(otherMap ConnectionConfigMap) (addedConnections, deletedConnections, changedConnections map[string][]*sdkproto.ConnectionConfig) {
	// results are maps of connections keyed by plugin instance
	addedConnections = make(map[string][]*sdkproto.ConnectionConfig)
	deletedConnections = make(map[string][]*sdkproto.ConnectionConfig)
	changedConnections = make(map[string][]*sdkproto.ConnectionConfig)

	for name, connection := range m {
		if otherConnection, ok := otherMap[name]; !ok {
			deletedConnections[connection.PluginInstance] = append(deletedConnections[connection.PluginInstance], connection)
		} else {
			// check for changes

			// special case - if the plugin has changed, treat this as a deletion and a re-add
			if connection.PluginInstance != otherConnection.PluginInstance {
				addedConnections[otherConnection.PluginInstance] = append(addedConnections[otherConnection.PluginInstance], otherConnection)
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
			addedConnections[otherConnection.PluginInstance] = append(addedConnections[otherConnection.PluginInstance], otherConnection)
		}
	}

	return
}
