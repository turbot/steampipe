package steampipeconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/plugin_manager"
	"github.com/turbot/steampipe/plugin_manager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
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
func CreateConnectionPlugin(connection *modconfig.Connection) (res *ConnectionPlugin, err error) {
	defer func() {
		if err != nil {
			// prefix error with the plugin name
			err = fmt.Errorf("failed to start plugin '%s': %s", connection.PluginShortName, err)
		}
	}()
	pluginName := connection.Plugin
	connectionName := connection.Name
	connectionConfig := connection.Config
	connectionOptions := connection.Options

	log.Printf("[TRACE] CreateConnectionPlugin connection: '%s', pluginName: '%s'", connectionName, pluginName)

	var pluginManager pluginshared.PluginManager
	if env := os.Getenv("STEAMPIPE_PLUGIN_MANAGER_DEBUG"); strings.ToLower(env) == "true" {
		// run plugin manager locally - for debugging
		log.Printf("[WARN] running plugin manager in-process for debugging")
		pluginManager, err = runPluginManagerInProcess()
	} else {
		pluginManager, err = plugin_manager.GetPluginManager()
	}
	// check the error from the plugin manager startup
	if err != nil {
		log.Printf("[WARN] failed to start plugin manager: %s", err)
		return nil, err
	}

	// ask the plugin manager for the plugin reattach config
	getResponse, err := pluginManager.Get(&proto.GetRequest{Connection: connectionName})
	if err != nil {
		log.Printf("[WARN] plugin manager failed to get reattach config for connection '%s': %s", connectionName, err)
		return nil, err
	}

	log.Printf("[TRACE] plugin manager returned reattach config for connection '%s' - pid %d",
		connectionName, getResponse.Reattach.Pid)

	// attach to the plugin process
	pluginClient, err := attachToPlugin(getResponse.Reattach.Convert(), pluginName)
	if err != nil {
		log.Printf("[TRACE] failed to attach to plugin for connection '%s' - pid %d: %s",
			connectionName, getResponse.Reattach.Pid, err)
		return nil, err
	}
	// set the connection config
	req := &sdkproto.SetConnectionConfigRequest{
		ConnectionName:   connectionName,
		ConnectionConfig: connectionConfig,
	}

	if err = pluginClient.SetConnectionConfig(req); err != nil {
		log.Printf("[TRACE] failed to set connection config: %s", err)
		return nil, err
	}

	// fetch the plugin schema
	schema, err := pluginClient.GetSchema()
	if err != nil {
		log.Printf("[TRACE] failed to get schema: %s", err)
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
	log.Printf("[TRACE] created connection plugin for connection: '%s', pluginName: '%s'", connectionName, pluginName)
	return c, nil
}

// use the reattach config to create a PluginClient for the plugin
func attachToPlugin(reattach *plugin.ReattachConfig, pluginName string) (*sdkgrpc.PluginClient, error) {
	return sdkgrpc.NewPluginClient(reattach, pluginName)
}

// function used for debugging the plugin manager
func runPluginManagerInProcess() (*plugin_manager.PluginManager, error) {
	steampipeConfig, err := LoadConnectionConfig()
	if err != nil {
		return nil, err
	}

	// discard logging from the plugin client (plugin logs will still flow through)
	loggOpts := &hclog.LoggerOptions{Name: "plugin", Output: ioutil.Discard}
	logger := logging.NewLogger(loggOpts)

	// build config map
	configMap := make(map[string]*proto.ConnectionConfig)
	for k, v := range steampipeConfig.Connections {
		configMap[k] = &proto.ConnectionConfig{
			Plugin:          v.Plugin,
			PluginShortName: v.PluginShortName,
			Config:          v.Config,
		}
	}
	return plugin_manager.NewPluginManager(configMap, logger), nil
}
