package steampipeconfig

import (
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/turbot/steampipe-plugin-sdk/grpc"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	pluginshared "github.com/turbot/steampipe-plugin-sdk/grpc/shared"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/options"
)

// ConnectionPlugin :: structure representing an instance of a plugin
// NOTE: currently this corresponds to a single connection, i.e. we have 1 plugin instance per connection
type ConnectionPlugin struct {
	ConnectionName    string
	ConnectionConfig  string
	ConnectionOptions *options.Connection
	PluginName        string
	Plugin            *grpc.PluginClient
	Schema            *proto.Schema
}

// CreateConnectionPlugin :: instantiate a plugin for a connection, fetch schema and send connection config
func CreateConnectionPlugin(connection *modconfig.Connection, disableLogger bool) (*ConnectionPlugin, error) {
	remoteSchema := connection.Plugin
	connectionName := connection.Name
	connectionConfig := connection.Config
	conectionOptions := connection.Options

	pluginPath, err := GetPluginPath(connection)
	if err != nil {
		return nil, err
	}

	// launch the plugin process.
	// create the plugin map
	pluginMap := map[string]plugin.Plugin{
		remoteSchema: &pluginshared.WrapperPlugin{},
	}
	loggOpts := &hclog.LoggerOptions{Name: "plugin"}
	// avoid logging if the plugin is being invoked by refreshConnections
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
	if err = SetConnectionConfig(connectionName, connectionConfig, pluginClient); err != nil {
		pluginClient.Client.Kill()
		return nil, err
	}

	schemaResponse, err := pluginClient.Stub.GetSchema(&proto.GetSchemaRequest{})
	if err != nil {
		pluginClient.Client.Kill()
		return nil, HandleGrpcError(err, connectionName, "GetSchema")
	}
	schema := schemaResponse.Schema

	// now create ConnectionPlugin object and add to map
	c := &ConnectionPlugin{
		ConnectionName:    connectionName,
		ConnectionConfig:  connectionConfig,
		ConnectionOptions: conectionOptions,
		PluginName:        remoteSchema,
		Plugin:            pluginClient,
		Schema:            schema,
	}
	return c, nil
}

// SetConnectionConfig sends the connection config
func SetConnectionConfig(connectionName string, connectionConfig string, pluginClient *grpc.PluginClient) error {
	req := &proto.SetConnectionConfigRequest{
		ConnectionName:   connectionName,
		ConnectionConfig: connectionConfig,
	}

	_, err := pluginClient.Stub.SetConnectionConfig(req)
	if err != nil {
		// create a new cleaner error, ignoring Not Implemented errors for backwards compatibility
		return HandleGrpcError(err, connectionName, "SetConnectionConfig")
	}
	return nil
}
