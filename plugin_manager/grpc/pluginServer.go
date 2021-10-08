package grpc

import (
	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

type GetPluginFunc func(req *pb.GetPluginRequest) (*pb.GetPluginResponse, error)

// PluginServer encapulates the plugin manager server
type PluginServer struct {
	pb.UnimplementedPluginManagerServer
}

func NewPluginServer() *PluginServer {
	return &PluginServer{}
}

func (s PluginServer) GetPlugin(req *pb.GetPluginRequest) (res *pb.GetPluginResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()
	return nil, nil

}

func (s PluginServer) Serve() {
	pluginMap := map[string]plugin.Plugin{
		"plugin_manager": &pluginshared.PluginManagerPlugin{Impl: s},
	}

	plugin.Serve(&plugin.ServeConfig{
		Plugins:    pluginMap,
		GRPCServer: plugin.DefaultGRPCServer,
		// A non-nil value here enables gRPC serving for this plugin...
		HandshakeConfig: pluginshared.Handshake,
	})
}
