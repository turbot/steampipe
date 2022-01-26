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
	"github.com/turbot/steampipe/pluginmanager"
	"github.com/turbot/steampipe/pluginmanager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/pluginmanager/grpc/shared"
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

// CreateConnectionPlugins instantiates plugins for specified connections, fetches schemas and sends connection config
func CreateConnectionPlugins(connections ...*modconfig.Connection) (connectionPluginMap map[string]*ConnectionPlugin, res *RefreshConnectionResult) {
	res = &RefreshConnectionResult{}
	log.Printf("[TRACE] CreateConnectionPlugin creating %d connections", len(connections))

	// build result map
	connectionPluginMap = make(map[string]*ConnectionPlugin, len(connections))
	// build list of connection names to pass to plugin manager 'get'
	connectionNames := make([]string, len(connections))
	for i, connection := range connections {
		connectionNames[i] = connection.Name
	}

	// get plugin manager
	pluginManager, err := getPluginManager()
	if err != nil {
		res.Error = err
		return nil, res
	}

	// ask the plugin manager for the reattach config for all required plugins
	getResponse, err := pluginManager.Get(&proto.GetRequest{Connections: connectionNames})
	if err != nil {
		res.Error = err
		return nil, res
	}

	// now create a connection plugin for each connection
	for _, connection := range connections {
		connectionPlugin, err := createConnectionPlugin(connection, getResponse)
		if err != nil {
			res.AddWarning(fmt.Sprintf("failed to start plugin '%s': %s", connection.PluginShortName, err))
			continue
		} else {
			connectionPluginMap[connection.Name] = connectionPlugin
		}
	}
	// now get populate schemas for all these connection plugins
	// - minimising the GetSchema calls we make to the unique schemas
	if err := populateConnectionPluginSchemas(connections, connectionPluginMap); err != nil {
		res.Error = err
		return nil, res
	}

	return connectionPluginMap, res
}

func populateConnectionPluginSchemas(connections []*modconfig.Connection, connectionPluginMap map[string]*ConnectionPlugin) error {
	// we will only need to fetch the schema once for each plugin (apart from plugins with dynamic schema)
	// we first new to build a map of schemas for each plugin

	pluginSchemaMap, err := buildPluginSchemaMap(connectionPluginMap)
	if err != nil {
		return err
	}

	// now build a map of connection to schema mode
	schemaModeMap := buildSchemaModeMap(connectionPluginMap, pluginSchemaMap)

	// now build a ConnectionSchemaMap object for the connections we are updating
	// - we can use this to identify the minimal set of schemas we need to fetch
	connectionSchemaMap := NewConnectionSchemaMapForConnections(connections, schemaModeMap)

	// for every connection with unique schema, fetch the schema and then set in all connections which share this schema
	for _, c := range connectionSchemaMap.UniqueSchemas() {
		connectionPlugin, ok := connectionPluginMap[c]
		if !ok {
			// we must have had issues loading this plugin
			continue
		}
		// retrieve the plugin schema from the schema map
		pluginName := connectionPlugin.PluginName
		schema := pluginSchemaMap[pluginName]
		// now set this schema for all connections which share it
		for _, connectionUsingSchema := range connectionSchemaMap[c] {
			connectionPluginMap[connectionUsingSchema].Schema = schema
		}
	}
	return nil
}

// build a map of plugin schemas
func buildPluginSchemaMap(connectionPluginMap map[string]*ConnectionPlugin) (map[string]*sdkproto.Schema, error) {
	var errors []error
	pluginSchemaMap := make(map[string]*sdkproto.Schema)
	for _, connectionPlugin := range connectionPluginMap {
		if _, ok := pluginSchemaMap[connectionPlugin.PluginName]; !ok {
			schema, err := connectionPlugin.PluginClient.GetSchema()
			if err != nil {
				log.Printf("[TRACE] failed to get schema for connection '%s': %s", connectionPlugin.ConnectionName, err)
				errors = append(errors, err)
				continue
			}
			pluginSchemaMap[connectionPlugin.PluginName] = schema
		}
	}
	if len(errors) > 0 {
		return nil, utils.CombineErrors(errors...)
	}
	return pluginSchemaMap, nil
}

func buildSchemaModeMap(connectionPluginMap map[string]*ConnectionPlugin, pluginSchemaMap map[string]*sdkproto.Schema) map[string]string {
	schemaModeMap := make(map[string]string, len(connectionPluginMap))

	for connectionName, connectionPlugin := range connectionPluginMap {
		schema := pluginSchemaMap[connectionPlugin.PluginName]
		schemaModeMap[connectionName] = schema.Mode
	}
	return schemaModeMap
}

func createConnectionPlugin(connection *modconfig.Connection, getResponse *proto.GetResponse) (*ConnectionPlugin, error) {
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
		return nil, err
	}

	// set the connection config
	req := &sdkproto.SetConnectionConfigRequest{
		ConnectionName:   connectionName,
		ConnectionConfig: connectionConfig,
	}

	if err = pluginClient.SetConnectionConfig(req); err != nil {
		log.Printf("[TRACE] failed to set connection config for connection '%s' - pid %d: %s",
			connectionName, reattach.Pid, err)
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
		SupportedOperations: supportedOperations,
	}
	log.Printf("[TRACE] created connection plugin for connection: '%s', pluginName: '%s'", connectionName, pluginName)
	return c, nil
}

// get plugin manager
// if STEAMPIPE_PLUGIN_MANAGER_DEBUG is set, create in process - otherwise connection tro grpc plugin
func getPluginManager() (pluginshared.PluginManager, error) {
	var pluginManager pluginshared.PluginManager
	var err error
	if env := os.Getenv("STEAMPIPE_PLUGIN_MANAGER_DEBUG"); strings.ToLower(env) == "true" {
		// run plugin manager locally - for debugging
		log.Printf("[WARN] running plugin manager in-process for debugging")
		pluginManager, err = runPluginManagerInProcess()
	} else {
		pluginManager, err = pluginmanager.GetPluginManager()
	}
	// check the error from the plugin manager startup
	if err != nil {
		log.Printf("[WARN] failed to start plugin manager: %s", err)
		return nil, err
	}
	return pluginManager, nil
}

// use the reattach config to create a PluginClient for the plugin
func attachToPlugin(reattach *plugin.ReattachConfig, pluginName string) (*sdkgrpc.PluginClient, error) {
	return sdkgrpc.NewPluginClient(reattach, pluginName)
}

// function used for debugging the plugin manager
func runPluginManagerInProcess() (*pluginmanager.PluginManager, error) {
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
	return pluginmanager.NewPluginManager(configMap, logger), nil
}
