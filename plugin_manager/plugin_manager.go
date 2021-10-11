package plugin_manager

import (
	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

// PluginManager is the real implementation of grpc.PluginManager
type PluginManager struct {
	pb.UnimplementedPluginManagerServer
}

func (m PluginManager) GetPlugin(req *pb.GetPluginRequest) (resp *pb.GetPluginResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()
	return &pb.GetPluginResponse{
		Protocol:        "FOO",
		ProtocolVersion: 0,
		Pid:             1234,
	}, nil
}

func (m PluginManager) Serve() {
	// create a plugin mapo, using ourselves as the implementation
	pluginMap := map[string]plugin.Plugin{
		pluginshared.PluginName: &pluginshared.PluginManagerPlugin{Impl: m},
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginMap,
		//  enable gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
