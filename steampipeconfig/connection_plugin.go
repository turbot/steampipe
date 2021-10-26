package steampipeconfig

import (
	"log"

	"github.com/hashicorp/go-plugin"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe/plugin_manager"
	"github.com/turbot/steampipe/plugin_manager/grpc/proto"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/options"
)

// ConnectionPlugin is a structure representing an instance of a plugin
// NOTE: currently this corresponds to a single steampipe connection,
// i.e. we have 1 plugin instance per steampipe connection
type ConnectionPlugin struct {
	ConnectionName      string
	ConnectionConfig    string
	ConnectionOptions   *options.Connection
	PluginName          string
	PluginClient        *sdkgrpc.PluginClient
	Schema              *sdkproto.Schema
	SupportedOperations *sdkproto.GetSupportedOperationsResponse
}

// CreateConnectionPlugin instantiates a plugin for a connection, fetches schema and sends connection config
func CreateConnectionPlugin(connection *modconfig.Connection, disableLogger bool) (*ConnectionPlugin, error) {
	pluginName := connection.Plugin
	connectionName := connection.Name
	connectionConfig := connection.Config
	connectionOptions := connection.Options

	log.Printf("[WARN] CreateConnectionPlugin connection: '%s', pluginName: '%s'", connectionName, pluginName)

	pluginManager, err := plugin_manager.GetPluginManager()
	// run locally - for debugging
	//pluginManager, err := getPluginManager()
	if err != nil {
		return nil, err
	}
	log.Printf("[WARN] got plugin manager")

	// ask the plugin manager for the plugin reattach config
	getResponse, err := pluginManager.Get(&proto.GetRequest{Connection: connectionName, DisableLogger: disableLogger})
	if err != nil {
		log.Printf("[WARN] plugin manager failed to get reattach config for connection '%s': %s", connectionName, err)
		return nil, err
	}

	log.Printf("[WARN] plugin manager returned reattach config for connection '%s' - pid %d",
		connectionName, getResponse.Reattach.Pid)

	// attach to the plugin process
	pluginClient, err := attachToPlugin(getResponse.Reattach.Convert(), pluginName, disableLogger)
	if err != nil {
		log.Printf("[WARN] failed to attach to plugin for connection '%s' - pid %d: %s",
			connectionName, getResponse.Reattach.Pid, err)
		return nil, err
	}
	// set the connection config
	req := &sdkproto.SetConnectionConfigRequest{
		ConnectionName:   connectionName,
		ConnectionConfig: connectionConfig,
	}

	if err = pluginClient.SetConnectionConfig(req); err != nil {
		pluginClient.Kill()
		return nil, err
	}

	// fetch the plugin schema
	schema, err := pluginClient.GetSchema()
	if err != nil {
		pluginClient.Kill()
		return nil, err
	}
	// fetch the supported operations
	supportedOperations, err := pluginClient.GetSupportedOperations()
	// ignore errors  - just create an empty support structure if needed
	if supportedOperations == nil {
		supportedOperations = &sdkproto.GetSupportedOperationsResponse{}
	}

	// now create ConnectionPlugin object return
	c := &ConnectionPlugin{
		ConnectionName:      connectionName,
		ConnectionConfig:    connectionConfig,
		ConnectionOptions:   connectionOptions,
		PluginName:          pluginName,
		PluginClient:        pluginClient,
		Schema:              schema,
		SupportedOperations: supportedOperations,
	}
	return c, nil
}

// use the reattach config to create a PluginClient for the plugin
func attachToPlugin(reattach *plugin.ReattachConfig, pluginName string, disableLogger bool) (*sdkgrpc.PluginClient, error) {
	return sdkgrpc.NewPluginClient(reattach, pluginName, disableLogger)
}

// function used for debugging the plugin manager
func getPluginManager() (*plugin_manager.PluginManager, error) {
	steampipeConfig, err := LoadConnectionConfig()
	if err != nil {
		return nil, err
	}
	// build config map
	configMap := make(map[string]*proto.ConnectionConfig)
	for k, v := range steampipeConfig.Connections {
		configMap[k] = &proto.ConnectionConfig{
			Plugin:          v.Plugin,
			PluginShortName: v.PluginShortName,
			Config:          v.Config,
		}
	}
	return plugin_manager.NewPluginManager(configMap), nil
}
