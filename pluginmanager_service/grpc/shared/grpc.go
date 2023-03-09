package shared

import (
	"context"

	pb "github.com/turbot/steampipe/pluginmanager_service/grpc/proto"
)

// GRPCClient is an implementation of PluginManager service that talks over GRPC.
type GRPCClient struct {
	// Proto client use to make the grpc service calls.
	client pb.PluginManagerClient
	// this context is created by the plugin package, and is canceled when the
	// plugin process ends.
	ctx context.Context
}

func (c *GRPCClient) Get(req *pb.GetRequest) (*pb.GetResponse, error) {
	return c.client.Get(c.ctx, req)
}

func (c *GRPCClient) Shutdown(req *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	return c.client.Shutdown(c.ctx, req)
}

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	pb.UnimplementedPluginManagerServer
	// This is the real implementation
	Impl PluginManager
}

func (m *GRPCServer) Get(_ context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	return m.Impl.Get(req)
}

func (m *GRPCServer) Shutdown(_ context.Context, req *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	return m.Impl.Shutdown(req)
}
