package main

import (
	"github.com/hashicorp/go-plugin"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.PluginMap,
		//  enable gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
