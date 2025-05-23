package proto

import (
	"slices"

	"github.com/hashicorp/go-plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

func NewReattachConfig(pluginName string, src *plugin.ReattachConfig, supportedOperations *SupportedOperations, connections []string) *ReattachConfig {
	return &ReattachConfig{
		Plugin:          pluginName,
		Protocol:        string(src.Protocol),
		ProtocolVersion: int64(src.ProtocolVersion),
		Addr: &NetAddr{
			Network: src.Addr.Network(),
			Address: src.Addr.String(),
		},
		Pid:                 int64(src.Pid),
		SupportedOperations: supportedOperations,
		Connections:         connections,
	}
}

// Convert converts from a protobuf reattach config to a plugin.ReattachConfig
func (r *ReattachConfig) Convert() *plugin.ReattachConfig {
	return &plugin.ReattachConfig{
		Protocol:        plugin.Protocol(r.Protocol),
		ProtocolVersion: int(r.ProtocolVersion),
		Addr: &SimpleAddr{
			NetworkString: r.Addr.Network,
			AddressString: r.Addr.Address,
		},
		Pid: int(r.Pid),
	}
}

func (r *ReattachConfig) AddConnection(connection string) {
	if !slices.Contains(r.Connections, connection) {
		r.Connections = append(r.Connections, connection)
	}
}
func (r *ReattachConfig) RemoveConnection(connection string) {
	existingConnections := r.Connections
	r.Connections = nil
	for _, existingConnections := range existingConnections {
		if existingConnections != connection {
			r.Connections = append(r.Connections, existingConnections)
		}
	}
}

func (r *ReattachConfig) UpdateConnections(configs []*proto.ConnectionConfig) {
	r.Connections = make([]string, len(configs))
	for i, c := range configs {
		r.Connections[i] = c.Connection
	}
}
