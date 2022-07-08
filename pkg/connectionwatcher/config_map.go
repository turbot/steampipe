package connectionwatcher

import (
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func NewConnectionConfigMap(connectionMap map[string]*modconfig.Connection) map[string]*sdkproto.ConnectionConfig {
	configMap := make(map[string]*sdkproto.ConnectionConfig)
	for k, v := range connectionMap {
		configMap[k] = &sdkproto.ConnectionConfig{
			Plugin:          v.Plugin,
			PluginShortName: v.PluginShortName,
			Config:          v.Config,
		}
	}
	return configMap
}
