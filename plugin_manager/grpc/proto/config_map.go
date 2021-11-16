package proto

import "github.com/turbot/steampipe/steampipeconfig/modconfig"

type ConnectionConfigMap map[string]*ConnectionConfig

func NewConnectionConfigMap(connectionMap map[string]*modconfig.Connection) ConnectionConfigMap {
	configMap := make(ConnectionConfigMap)
	for k, v := range connectionMap {
		configMap[k] = &ConnectionConfig{
			Plugin:          v.Plugin,
			PluginShortName: v.PluginShortName,
			Config:          v.Config,
		}
	}
	return configMap
}
