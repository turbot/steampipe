package main

import (
	"github.com/turbot/go-kit/helpers"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
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
	return nil, nil
}
