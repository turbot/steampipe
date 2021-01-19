package connection_config

import (
	"log"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/turbot/steampipe-plugin-sdk/grpc"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	pluginshared "github.com/turbot/steampipe-plugin-sdk/grpc/shared"
	"github.com/turbot/steampipe-plugin-sdk/logging"
)

type ConnectionPluginOptions struct {
	PluginFQN      string
	ConnectionName string
	DisableLogger  bool
}
type ConnectionPlugin struct {
	ConnectionName string
	PluginName     string
	Plugin         *grpc.PluginClient
	Schema         *proto.Schema
}

func CreateConnectionPlugin(options *ConnectionPluginOptions) (*ConnectionPlugin, error) {

	remoteSchema := options.PluginFQN
	connectionName := options.ConnectionName
	disableLogger := options.DisableLogger

	log.Printf("[DEBUG] createConnectionPlugin name %s, remoteSchema %s \n", connectionName, remoteSchema)
	pluginPath, err := GetPluginPath(remoteSchema)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] found pluginPath %s\n", pluginPath)

	// launch the plugin process.
	// create the plugin map
	pluginMap := map[string]plugin.Plugin{
		remoteSchema: &pluginshared.WrapperPlugin{},
	}
	loggOpts := &hclog.LoggerOptions{Name: "plugin"}
	// HACK avoid logging if the plugin is being invoked by refreshConnections
	if disableLogger {
		loggOpts.Exclude = func(hclog.Level, string, ...interface{}) bool { return true }
	}
	logger := logging.NewLogger(loggOpts)

	cmd := exec.Command(pluginPath)
	// pass env to command
	cmd.Env = os.Environ()
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginMap,
		// this failed when running from extension
		//Cmd:              exec.Command("sh", "-c", pluginPath),
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(remoteSchema)
	if err != nil {
		return nil, err
	}
	// We should have a stub plugin now
	p := raw.(pluginshared.WrapperPluginClient)
	pluginClient := &grpc.PluginClient{
		Name:   remoteSchema,
		Path:   pluginPath,
		Client: client,
		Stub:   p,
	}
	schemaResponse, err := pluginClient.Stub.GetSchema(&proto.GetSchemaRequest{})
	if err != nil {
		pluginClient.Client.Kill()
		return nil, err
	}
	schema := schemaResponse.Schema

	// now create ConnectionPlugin object and add to map
	c := &ConnectionPlugin{ConnectionName: connectionName, PluginName: remoteSchema, Plugin: pluginClient, Schema: schema}
	return c, nil
}
