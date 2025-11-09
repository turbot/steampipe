package mocks

import (
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
	pluginshared "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/shared"
)

// MockPluginManager is a mock implementation of PluginManager for testing
type MockPluginManager struct {
	// Function fields for configurable behavior
	GetFunc                func(req *pb.GetRequest) (*pb.GetResponse, error)
	RefreshConnectionsFunc func(req *pb.RefreshConnectionsRequest) (*pb.RefreshConnectionsResponse, error)
	ShutdownFunc           func(req *pb.ShutdownRequest) (*pb.ShutdownResponse, error)

	// Track calls
	GetCalls                []GetCall
	RefreshConnectionsCalls int
	ShutdownCalls           int
}

// GetCall tracks a single Get call
type GetCall struct {
	Connections []string
}

// Ensure MockPluginManager implements pluginshared.PluginManager
var _ pluginshared.PluginManager = (*MockPluginManager)(nil)

// Get implements pluginshared.PluginManager
func (m *MockPluginManager) Get(req *pb.GetRequest) (*pb.GetResponse, error) {
	m.GetCalls = append(m.GetCalls, GetCall{
		Connections: req.Connections,
	})
	if m.GetFunc != nil {
		return m.GetFunc(req)
	}
	return &pb.GetResponse{}, nil
}

// RefreshConnections implements pluginshared.PluginManager
func (m *MockPluginManager) RefreshConnections(req *pb.RefreshConnectionsRequest) (*pb.RefreshConnectionsResponse, error) {
	m.RefreshConnectionsCalls++
	if m.RefreshConnectionsFunc != nil {
		return m.RefreshConnectionsFunc(req)
	}
	return &pb.RefreshConnectionsResponse{}, nil
}

// Shutdown implements pluginshared.PluginManager
func (m *MockPluginManager) Shutdown(req *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	m.ShutdownCalls++
	if m.ShutdownFunc != nil {
		return m.ShutdownFunc(req)
	}
	return &pb.ShutdownResponse{}, nil
}
