package connectionwatcher

import (
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type ConnectionConfigMap map[string]*sdkproto.ConnectionConfig

func NewConnectionConfigMap(connectionMap map[string]*modconfig.Connection) ConnectionConfigMap {
	configMap := make(ConnectionConfigMap)
	for k, v := range connectionMap {
		configMap[k] = &sdkproto.ConnectionConfig{
			Connection:            v.Name,
			Plugin:                v.Plugin,
			PluginShortName:       v.PluginShortName,
			Config:                v.Config,
			ChildConnections:      v.GetResolveConnectionNames(),
			TableAggregationSpecs: v.TableAggregationSpecs.ToProto(),
		}
	}

	return configMap
}

func (m ConnectionConfigMap) Diff(otherMap ConnectionConfigMap) (addedConnections, deletedConnections, changedConnections map[string][]*sdkproto.ConnectionConfig) {
	// results are maps os  connections keyed by plugin
	addedConnections = make(map[string][]*sdkproto.ConnectionConfig)
	deletedConnections = make(map[string][]*sdkproto.ConnectionConfig)
	changedConnections = make(map[string][]*sdkproto.ConnectionConfig)

	// TODO if anything other than the plugin specific connection config has changed,
	// treat as a deletion and addition of a new connection
	// https://github.com/turbot/steampipe/issues/2348

	for name, connection := range m {
		if otherConnection, ok := otherMap[name]; !ok {
			deletedConnections[connection.Plugin] = append(deletedConnections[connection.Plugin], connection)
		} else {
			// check for changes

			// special case - if the plugin has changed, treat this as a deletion and a re-add
			if connection.Plugin != otherConnection.Plugin {
				addedConnections[otherConnection.Plugin] = append(addedConnections[otherConnection.Plugin], otherConnection)
				deletedConnections[connection.Plugin] = append(deletedConnections[connection.Plugin], connection)
			} else {
				if !connection.Equals(otherConnection) {
					changedConnections[connection.Plugin] = append(changedConnections[connection.Plugin], otherConnection)
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
