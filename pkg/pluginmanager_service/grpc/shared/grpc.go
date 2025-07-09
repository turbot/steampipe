package shared

import (
	"context"

	"github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

// GRPCClient is an implementation of PluginManager service that talks over GRPC.
type GRPCClient struct {
	// Proto client use to make the grpc service calls.
	client proto.PluginManagerClient
	// this context is created by the plugin package, and is canceled when the
	// plugin process ends.
	ctx context.Context
}

func (c *GRPCClient) Get(req *proto.GetRequest) (*proto.GetResponse, error) {
	return c.client.Get(c.ctx, req)
}
func (c *GRPCClient) RefreshConnections(req *proto.RefreshConnectionsRequest) (*proto.RefreshConnectionsResponse, error) {
	return c.client.RefreshConnections(c.ctx, req)
}

func (c *GRPCClient) Shutdown(req *proto.ShutdownRequest) (*proto.ShutdownResponse, error) {
	return c.client.Shutdown(c.ctx, req)
}

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	proto.UnimplementedPluginManagerServer
	// This is the real implementation
	Impl PluginManager
}

func (m *GRPCServer) Get(_ context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	return m.Impl.Get(req)
}
func (m *GRPCServer) RefreshConnections(_ context.Context, req *proto.RefreshConnectionsRequest) (*proto.RefreshConnectionsResponse, error) {
	return m.Impl.RefreshConnections(req)
}

func (m *GRPCServer) Shutdown(_ context.Context, req *proto.ShutdownRequest) (*proto.ShutdownResponse, error) {
	return m.Impl.Shutdown(req)
}
