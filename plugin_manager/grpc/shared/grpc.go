package shared

import (
	"context"

	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
)

// GRPCClient is an implementation of PluginManager service that talks over GRPC.
type GRPCClient struct {
	// Proto client use to make the grpc service calls.
	client pb.PluginManagerClient
	// this context is created by the plugin package, and is canceled when the
	// plugin process ends.
	ctx context.Context
}

func (c *GRPCClient) GetPlugin(req *pb.GetPluginRequest) (*pb.GetPluginResponse, error) {
	return c.client.GetPlugin(c.ctx, req)
}

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl PluginManager
}

func (m *GRPCServer) GetPlugin(_ context.Context, req *pb.GetPluginRequest) (*pb.GetPluginResponse, error) {
	//log.Printf("[WARN] _PluginManager_GetPlugin_Handler")
	return m.Impl.GetPlugin(req)
}
