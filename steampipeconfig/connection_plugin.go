package steampipeconfig

import (
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/grpc"
	sdkpb "github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	sdkpluginshared "github.com/turbot/steampipe-plugin-sdk/grpc/shared"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/plugin_manager"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/options"
	"github.com/turbot/steampipe/utils"
)

// ConnectionPlugin :: structure representing an instance of a plugin
// NOTE: currently this corresponds to a single connection, i.e. we have 1 plugin instance per connection
type ConnectionPlugin struct {
	ConnectionName    string
	ConnectionConfig  string
	ConnectionOptions *options.Connection
	PluginName        string
	Plugin            *sdkgrpc.PluginClient
	Schema            *sdkpb.Schema
}

// CreateConnectionPlugin :: instantiate a plugin for a connection, fetch schema and send connection config
func CreateConnectionPlugin(connection *modconfig.Connection, disableLogger bool) (*ConnectionPlugin, error) {
	pluginName := connection.Plugin
	connectionName := connection.Name
	connectionConfig := connection.Config
	connectionOptions := connection.Options

	log.Printf("[TRACE] CreateConnectionPlugin connection: '%s', pluginName: '%s'", connectionName, pluginName)

	pluginManager, err := plugin_manager.GetPluginManager()
	// run locally - for debugging
	//pluginManager, err := getPluginManager()
	if err != nil {
		return nil, err
	}

	getResponse, err := pluginManager.Get(&pb.GetRequest{Connection: connectionName, DisableLogger: disableLogger})
	if err != nil {
		log.Printf("[TRACE] plugin manager failed to get reattach config for connection '%s': %s", connectionName, err)
		return nil, err
	}

	log.Printf("[TRACE] plugin manager returned reconnect config for connection '%s' - pid %d",
		connectionName, getResponse.Reattach.Pid)

	// launch the plugin process.
	pluginClient, err := attachToPlugin(getResponse, pluginName, disableLogger)
	if err != nil {
		log.Printf("[TRACE] failed to attach to plugin for connection '%s' - pid %d : %s",
			connectionName, getResponse.Reattach.Pid, err)
		return nil, err
	}
	if err = SetConnectionConfig(connectionName, connectionConfig, pluginClient); err != nil {
		pluginClient.Client.Kill()
		return nil, err
	}

	schemaResponse, err := pluginClient.Stub.GetSchema(&sdkpb.GetSchemaRequest{})
	if err != nil {
		pluginClient.Client.Kill()
		return nil, utils.HandleGrpcError(err, connectionName, "getSchema")
	}
	schema := schemaResponse.Schema

	// now create ConnectionPlugin object and add to map
	c := &ConnectionPlugin{
		ConnectionName:    connectionName,
		ConnectionConfig:  connectionConfig,
		ConnectionOptions: connectionOptions,
		PluginName:        pluginName,
		Plugin:            pluginClient,
		Schema:            schema,
	}
	return c, nil
}

func attachToPlugin(getResponse *pb.GetResponse, pluginName string, disableLogger bool) (*sdkgrpc.PluginClient, error) {
	// create the plugin map
	pluginMap := map[string]plugin.Plugin{
		pluginName: &sdkpluginshared.WrapperPlugin{},
	}
	// avoid logging if the plugin is being invoked by refreshConnections
	loggOpts := &hclog.LoggerOptions{Name: "plugin"}
	if disableLogger {
		loggOpts.Exclude = func(hclog.Level, string, ...interface{}) bool { return true }
	}
	logger := logging.NewLogger(loggOpts)

	// create grpc client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  sdkpluginshared.Handshake,
		Plugins:          pluginMap,
		Reattach:         getResponse.Reattach.Convert(),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})

	// connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// request the plugin
	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		return nil, err
	}
	// we should have a stub plugin now
	p := raw.(sdkpluginshared.WrapperPluginClient)
	pluginClient := &sdkgrpc.PluginClient{
		Name:   pluginName,
		Client: client,
		Stub:   p,
	}
	return pluginClient, nil
}

func getPluginManager() (*plugin_manager.PluginManager, error) {
	steampipeConfig, err := LoadConnectionConfig()
	if err != nil {
		return nil, err
	}
	// build config map
	configMap := make(map[string]*pb.ConnectionConfig)
	for k, v := range steampipeConfig.Connections {
		configMap[k] = &pb.ConnectionConfig{
			Plugin:          v.Plugin,
			PluginShortName: v.PluginShortName,
			Config:          v.Config,
		}
	}
	return plugin_manager.NewPluginManager(configMap), nil
}

// SetConnectionConfig sends the connection config
func SetConnectionConfig(connectionName string, connectionConfig string, pluginClient *sdkgrpc.PluginClient) error {
	req := &sdkpb.SetConnectionConfigRequest{
		ConnectionName:   connectionName,
		ConnectionConfig: connectionConfig,
	}

	_, err := pluginClient.Stub.SetConnectionConfig(req)
	if err != nil {
		// create a new cleaner error, ignoring Not Implemented errors for backwards compatibility
		return utils.HandleGrpcError(err, connectionName, "SetConnectionConfig")
	}
	return nil
}
