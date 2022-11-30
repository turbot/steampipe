// Package shared contains shared data between the host and plugins.
package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	pb "github.com/turbot/steampipe/pluginmanager_service/grpc/proto"
	"google.golang.org/grpc"
)

const PluginName = "steampipe_plugin_manager"

// PluginMap is a ma of the plugins supported, _without the implementation_
// this used to create a GRPC client
var PluginMap = map[string]plugin.Plugin{
	PluginName: &PluginManagerPlugin{},
}

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	MagicCookieKey:   "PLUGIN_MANAGER_MAGIC_COOKIE",
	MagicCookieValue: "really-complex-permanent-string-value",
}

// PluginManager is the interface for the plugin manager service
type PluginManager interface {
	Get(req *pb.GetRequest) (*pb.GetResponse, error)
	Shutdown(req *pb.ShutdownRequest) (*pb.ShutdownResponse, error)
}

// PluginManagerPlugin is the implementation of plugin.GRPCServer so we can serve/consume this.
type PluginManagerPlugin struct {
	// GRPCPlugin must still implement the Stub interface
	plugin.Plugin
	// Concrete implementation
	Impl PluginManager
}

func (p *PluginManagerPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	//fmt.Println("GRPCServer")
	pb.RegisterPluginManagerServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient returns a GRPCClient, called by Dispense
func (p *PluginManagerPlugin) GRPCClient(ctx context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: pb.NewPluginManagerClient(c), ctx: ctx}, nil
}
