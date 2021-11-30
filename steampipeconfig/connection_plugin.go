package steampipeconfig

import (
	"fmt"
	"io"
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
	"github.com/turbot/steampipe/utils"
)

// ConnectionPlugin is a structure representing an instance of a plugin
// NOTE: this corresponds to a single steampipe connection,
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
func CreateConnectionPlugin(connection *modconfig.Connection) (*ConnectionPlugin, error) {

	res, err := CreateConnectionPlugins([]*modconfig.Connection{connection}, nil)
	if err != nil {
		return nil, err
	}
	return res[connection.Name], nil
}

func CreateConnectionPlugins(connections []*modconfig.Connection, connectionState ConnectionDataMap) (res map[string]*ConnectionPlugin, err error) {
	defer func() {
		// TOODO
		//if err != nil {
		//	// prefix error with the plugin name
		//	err = fmt.Errorf("failed to start plugin '%s': %s", connection.PluginShortName, err)
		//}
	}()

	connectionNames := make([]string, len(connections))
	for i, connection := range connections {
		connectionNames[i] = connection.Name
	}

	res = make(map[string]*ConnectionPlugin, len(connections))

	log.Printf("[TRACE] CreateConnectionPlugin creating %d connections", len(connectionNames))

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
	getResponse, err := pluginManager.Get(&proto.GetRequest{Connections: connectionNames})
	if err != nil {

	}

	var errors []error
	var clients = make(map[string]*sdkgrpc.PluginClient)

	for _, connection := range connections {
		pluginName := connection.Plugin
		connectionName := connection.Name
		connectionConfig := connection.Config
		connectionOptions := connection.Options

		reattach := getResponse.ReattachMap[connectionName]
		log.Printf("[TRACE] plugin manager returned reattach config for connection '%s' - pid %d, reattach %v",
			connectionName, reattach.Pid, reattach)
		if reattach.Pid == 0 {
			log.Printf("[WARN] plugin manager returned nil PID for %s", connectionName)
			return nil, fmt.Errorf("plugin manager returned nil PID for %s", connectionName)
		}

		// attach to the plugin process
		pluginClient, err := attachToPlugin(reattach.Convert(), pluginName)
		if err != nil {
			log.Printf("[TRACE] failed to attach to plugin for connection '%s' - pid %d: %s",
				connectionName, reattach.Pid, err)
			errors = append(errors, err)
			continue
		}
		// save client so it can be reused when fetching schemas
		clients[connectionName] = pluginClient

		// set the connection config
		req := &sdkproto.SetConnectionConfigRequest{
			ConnectionName:   connectionName,
			ConnectionConfig: connectionConfig,
		}

		if err = pluginClient.SetConnectionConfig(req); err != nil {
			log.Printf("[TRACE] failed to set connection config for connection '%s' - pid %d: %s",
				connectionName, reattach.Pid, err)
			errors = append(errors, err)
			continue
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
			SupportedOperations: supportedOperations,
		}
		log.Printf("[TRACE] created connection plugin for connection: '%s', pluginName: '%s'", connectionName, pluginName)

		res[connectionName] = c
	}
	if len(errors) > 0 {
		return nil, utils.CombineErrors(errors...)
	}

	// now get schemas

	// we will only need to fetch the schema once for each plugin (apart from plugins with dynamic schema)
	// build a ConnectionSchemaMap object for the connections we are updating
	// - we can use this to identify the minimal set of schemas we need to fetch
	// if only one connection was passed, load the schema
	connectionSchemas, err := NewConnectionSchemaMapForConnections(connectionNames, connectionState)
	if err != nil {
		return nil, err
	}

	// for every connection with unique schema, fetch the schema and then set in all connections which share this schema
	for _, c := range connectionSchemas.UniqueSchemas() {
		// fetch the plugin schema
		pluginClient := clients[c]
		schema, err := pluginClient.GetSchema()
		if err != nil {
			log.Printf("[TRACE] failed to get schema for connection '%s': %s", c, err)
			errors = append(errors, err)
			continue
		}
		// now set this schema for all connections which share it
		for _, connectionUsingSchema := range connectionSchemas[c] {
			res[connectionUsingSchema].Schema = schema
		}
	}

	if len(errors) > 0 {
		return nil, utils.CombineErrors(errors...)
	}

	return res, nil
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
	loggOpts := &hclog.LoggerOptions{Name: "plugin", Output: io.Discard}
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
