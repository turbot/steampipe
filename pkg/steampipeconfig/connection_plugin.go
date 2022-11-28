package steampipeconfig

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-plugin"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"github.com/turbot/steampipe/pluginmanager"
	"github.com/turbot/steampipe/pluginmanager_service/grpc/proto"
)

type ConnectionPluginData struct {
	Name    string
	Config  string
	Options *options.Connection
	Schema  *sdkproto.Schema
}

// ConnectionPlugin is a structure representing an instance of a plugin
// for non-legacy plugins, each plugin instance supportds multiple connections
// the config, options and schema for each connection is stored in  ConnectionMap
type ConnectionPlugin struct {
	// map of connection data (name, config, options)
	// keyed by connection name
	ConnectionMap       map[string]*ConnectionPluginData
	PluginName          string
	PluginClient        *sdkgrpc.PluginClient
	SupportedOperations *proto.SupportedOperations
}

func (p ConnectionPlugin) addConnection(name string, config string, connectionOptions *options.Connection) {
	p.ConnectionMap[name] = &ConnectionPluginData{
		Name:    name,
		Config:  config,
		Options: connectionOptions,
	}
}

func (p ConnectionPlugin) IncludesConnection(name string) bool {
	_, ok := p.ConnectionMap[name]
	return ok
}

// GetSchema returns the cached schema if it is static, or if it is dynamic, refetch it
func (p ConnectionPlugin) GetSchema(connectionName string) (*sdkproto.Schema, error) {
	connectionData, ok := p.ConnectionMap[connectionName]
	if ok {
		// if the schema mode is static, return the cached schema
		if connectionData.Schema.Mode == sdkplugin.SchemaModeStatic {
			return connectionData.Schema, nil
		}
	}
	// otherwise this is a dynamic schema - refetch it
	// we need to do this in case it has changed (for example as a result of a file watching event)
	schema, err := p.PluginClient.GetSchema(connectionName)
	if err != nil {
		log.Printf("[TRACE] failed to get schema for connection '%s': %s", connectionName, err)
		return nil, err
	}
	// update schema in our map
	connectionData.Schema = schema

	return schema, nil
}

func NewConnectionPlugin(pluginName string, pluginClient *sdkgrpc.PluginClient, supportedOperations *proto.SupportedOperations) *ConnectionPlugin {
	return &ConnectionPlugin{
		PluginName:          pluginName,
		PluginClient:        pluginClient,
		SupportedOperations: supportedOperations,
		ConnectionMap:       make(map[string]*ConnectionPluginData)}
}

// CreateConnectionPlugins instantiates plugins for specified connections, and fetches schemas
func CreateConnectionPlugins(connectionsToCreate []*modconfig.Connection) (requestedConnectionPluginMap map[string]*ConnectionPlugin, res *RefreshConnectionResult) {
	res = &RefreshConnectionResult{}
	requestedConnectionPluginMap = make(map[string]*ConnectionPlugin)
	if len(connectionsToCreate) == 0 {
		return
	}
	log.Printf("[TRACE] CreateConnectionPlugin creating %d connections", len(connectionsToCreate))

	// build result map, keyed by connection name
	requestedConnectionPluginMap = make(map[string]*ConnectionPlugin, len(connectionsToCreate))
	// build list of connection names to pass to plugin manager 'get'
	connectionNames := make([]string, len(connectionsToCreate))
	for i, connection := range connectionsToCreate {
		connectionNames[i] = connection.Name
	}

	// get plugin manager
	pluginManager, err := pluginmanager.GetPluginManager()
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

	// if there were any failures, display them
	for plugin, failure := range getResponse.FailureMap {
		res.AddWarning(fmt.Sprintf("failed to start plugin '%s': %s", plugin, failure))
	}

	// now create or retrieve a connection plugin for each connection

	// NOTE: multiple connections may use the same plugin
	// store a map of multi connection plugins, keyed by connection name
	multiConnectionPlugins := make(map[string]*ConnectionPlugin)

	for _, connection := range connectionsToCreate {
		// is this connection provided by a plugin we have already instantiated?
		if existingConnectionPlugin, ok := multiConnectionPlugins[connection.Plugin]; ok {
			log.Printf("[TRACE] CreateConnectionPlugins - connection %s is provided by existing connectionPlugin %s - reusing", connection.Name, connection.Plugin)
			// store the existing connection plugin in the result map
			requestedConnectionPluginMap[connection.Name] = existingConnectionPlugin
			continue
		}

		// do we have a reattach config for this connection's plugin
		if _, ok := getResponse.ReattachMap[connection.Name]; !ok {
			log.Printf("[TRACE] CreateConnectionPlugins skipping connection '%s', plugin '%s' as plugin manager failed to start it", connection.Name, connection.Plugin)
			continue
		}

		// so we have a reattach - create a connection plugin
		reattach := getResponse.ReattachMap[connection.Name]
		// if this is a legacy aggregator connection, skip - we do not instantiate connectionPlugins for these
		if connection.Type == modconfig.ConnectionTypeAggregator && !reattach.SupportedOperations.MultipleConnections {
			log.Printf("[TRACE] %s is a legacy aggregator connection - NOT creating a ConnectionPlugin", connection.Name)
			continue
		}

		connectionPlugin, err := createConnectionPlugin(connection, reattach)
		if err != nil {
			res.AddWarning(fmt.Sprintf("failed to start plugin '%s': %s", connection.PluginShortName, err))
			continue
		}
		requestedConnectionPluginMap[connection.Name] = connectionPlugin
		if connectionPlugin.SupportedOperations.MultipleConnections {
			// if it supports multiple connections, store in multiConnectionPlugins too
			multiConnectionPlugins[connection.Plugin] = connectionPlugin
		}
	}
	log.Printf("[TRACE] all connection plugins created, populating schemas")

	// now get populate schemas for all these connection plugins
	if err := populateConnectionPluginSchemas(requestedConnectionPluginMap); err != nil {
		res.Error = err
		return nil, res
	}

	log.Printf("[TRACE] populate schemas complete")

	return requestedConnectionPluginMap, res
}

func populateConnectionPluginSchemas(requestedConnectionPluginMap map[string]*ConnectionPlugin) error {

	// requestedConnectionPluginMap is a map of connection plugins, keyed by connection name
	// the connection names which are the keys of this map are the connections
	// which were _requested_ in the parent CreateConnectionPlugins call (i.e. not necessarily all connections)
	// NOTE: the connection plugins may provide  _more_ connections that those requested
	// - we need to populate the schema for _all_ of them

	// build a map keyed by _all_ connection names provided by the connection plugins
	connectionPluginMap := fullConnectionPluginMap(requestedConnectionPluginMap)

	var errors []error

	// build map of the static schemas, keyed by plugin
	staticSchemas := make(map[string]*sdkproto.Schema)

	log.Printf("[TRACE] populateConnectionPluginSchemas")

	for connectionName, connectionPlugin := range connectionPluginMap {
		log.Printf("[TRACE] populateConnectionPluginSchemas: connectionName: %s", connectionName)
		// does this plugin  exist in the static schema map?
		schema, ok := staticSchemas[connectionPlugin.PluginName]
		if !ok {
			log.Printf("[TRACE] schema does not exist in list of static schemas, fetching")
			log.Printf("[TRACE] GetSchema %s", connectionName)

			// if not, fetch the schema
			var err error
			schema, err = connectionPlugin.PluginClient.GetSchema(connectionName)
			if err != nil {
				log.Printf("[TRACE] failed to get schema for connection '%s': %s", connectionName, err)
				errors = append(errors, err)
				continue
			}

			log.Printf("[TRACE] got schema, mode: %s", schema.Mode)
			// if the schema is static, add to static schema map
			if schema.Mode == sdkplugin.SchemaModeStatic {
				staticSchemas[connectionPlugin.PluginName] = schema
			}
		}

		log.Printf("[TRACE] add schema to connection map for connection name %s, len %d", connectionName, len(schema.Schema))

		// set the schema on the connection plugin
		connectionPlugin.ConnectionMap[connectionName].Schema = schema

	}
	if len(errors) > 0 {
		return error_helpers.CombineErrors(errors...)
	}
	return nil

	// we will only need to fetch the schema once for each plugin (apart from plugins with dynamic schema)

	// map of plugin schemas - add schema for each plugin with a static schema

	//

	// we first new to build a map of schemas for each plugin
	//
	////log.Printf("[TRACE] populateConnectionPluginSchemas for %d connections", len(connections))
	//staticSchemaMap, dynamicSchemaMap, err := buildPluginSchemaMap(connectionNames, pluginMap)
	//if err != nil {
	//	return err
	//}
	//
	//// now build a map of connection to schema mode
	//schemaModeMap := buildSchemaModeMap(connectionPluginMap, pluginSchemaMap)
	//
	//// now build a ConnectionSchemaMap object for the connections we are updating
	//// - we can use this to identify the minimal set of schemas we need to fetch
	//connectionSchemaMap := NewConnectionSchemaMapForConnections(connections, schemaModeMap)
	//
	//// for every connection with unique schema, fetch the schema and then set in all connections which share this schema
	//for _, c := range connectionSchemaMap.UniqueSchemas() {
	//	connectionPlugin, ok := connectionPluginMap[c]
	//	if !ok {
	//		// we must have had issues loading this plugin
	//		continue
	//	}
	//	// retrieve the plugin schema from the schema map
	//	pluginName := connectionPlugin.PluginName
	//	schema := pluginSchemaMap[pluginName]
	//	// now set this schema for all connections which share it
	//	for _, connectionUsingSchema := range connectionSchemaMap[c] {
	//		connectionPluginMap[connectionUsingSchema].ConnectionMap[connectionUsingSchema].Schema = schema
	//	}
	//}
	//log.Printf("[WARN] populateConnectionPluginSchemas 5")
	//return nil
}

// given a map of connection names to the connectionPlugins which proivide them,
// return a map of _all_ connections provided by the connection plugins
func fullConnectionPluginMap(sparseConnectionPluginMap map[string]*ConnectionPlugin) map[string]*ConnectionPlugin {
	// sparseConnectionPluginMap is a map of ConnectionPlugins keyed by connection name
	// NOTE: the connection plugins may provide  _more_ connections than the keys of the map
	connectionNameMap := make(map[string]*ConnectionPlugin)

	for _, connectionPlugin := range sparseConnectionPluginMap {
		for connectionName := range connectionPlugin.ConnectionMap {
			connectionNameMap[connectionName] = connectionPlugin
		}
	}

	return connectionNameMap
}

//func populateMultiConnectionPluginSchemas(connections []*modconfig.Connection, connectionPluginMap map[string]*ConnectionPlugin, multiConnectionPlugins map[string]*ConnectionPlugin) error {
//
//	// we will only need to fetch the schema once for each plugin (apart from plugins with dynamic schema)
//	// we first new to build a map of schemas for each plugin
//
//	pluginSchemaMap, err := buildPluginSchemaMap(connectionPluginMap)
//	if err != nil {
//		return err
//	}
//
//	// now build a map of connection to schema mode
//	schemaModeMap := buildSchemaModeMap(connectionPluginMap, pluginSchemaMap)
//
//	// now build a ConnectionSchemaMap object for the connections we are updating
//	// - we can use this to identify the minimal set of schemas we need to fetch
//	connectionSchemaMap := NewConnectionSchemaMapForConnections(connections, schemaModeMap)
//
//	// for every connection with unique schema, fetch the schema and then set in all connections which share this schema
//	for _, c := range connectionSchemaMap.UniqueSchemas() {
//		connectionPlugin, ok := connectionPluginMap[c]
//		if !ok {
//			// we must have had issues loading this plugin
//			continue
//		}
//		// retrieve the plugin schema from the schema map
//		pluginName := connectionPlugin.PluginName
//		schema := pluginSchemaMap[pluginName]
//		// now set this schema for all connections which share it
//		for _, connectionUsingSchema := range connectionSchemaMap[c] {
//			connectionPluginMap[connectionUsingSchema].Schema = schema
//		}
//	}
//	return nil
//}

// build a map of plugin schemas, keyed by plugin name
// NOTE: if a plugin has dynamic schema, we will later need to fetch the schema for _all_ connections
//func buildPluginSchemaMap(pluginMap *connectionPluginMap) (error) {
//	var errors []error
//
//	// build map of the static schemas, keyed by plugin
//	staticSchemas := make(map[string]*sdkproto.Schema)
//
//	for connectionName, connectionPlugin := range pluginMap.connectionNameMap {
//		// does this plugin  exist in the static schema map?
//		schema, ok := staticSchemas[connectionPlugin.PluginName]
//		if !ok {
//			// if not, fetch the schema
//			var err error
//			schema, err = connectionPlugin.PluginClient.GetSchema(connectionName)
//			if err != nil {
//				log.Printf("[TRACE] failed to get schema for connection '%s': %s", connectionName, err)
//				errors = append(errors, err)
//				continue
//			}
//
//			// if the schema is static, add to static schema map
//			if schema.Mode == sdkplugin.SchemaModeStatic {
//				staticSchemas[connectionPlugin.PluginName] = schema
//			}
//		}
//
//		// set the schema on the connection plugin
//		connectionPlugin.ConnectionMap[connectionName].Schema = schema
//
//	}
//	if len(errors) > 0 {
//		return utils.CombineErrors(errors...)
//	}
//	return nil
//}
//
//func buildPluginMaps(connectionPluginMap map[string]*ConnectionPlugin) (pluginConnectionMap map[string][]string, pluginNameMap map[string]*ConnectionPlugin) {
//	pluginConnectionMap = make(map[string][]string)
//	pluginNameMap = make(map[string]*ConnectionPlugin)
//	for _, connectionPlugin := range connectionPluginMap {
//		connectionNames := connectionPlugin.GetConnectionNames()
//		pluginConnectionMap[connectionPlugin.PluginName] = append(pluginConnectionMap[connectionPlugin.PluginName], connectionNames...)
//		if _, ok := pluginNameMap[connectionPlugin.PluginName]; !ok {
//			pluginNameMap[connectionPlugin.PluginName] = connectionPlugin
//		}
//	}
//	return
//}
//
//func buildSchemaModeMap(connectionPluginMap map[string]*ConnectionPlugin, pluginSchemaMap map[string]*sdkproto.Schema) map[string]string {
//	schemaModeMap := make(map[string]string, len(connectionPluginMap))
//
//	for connectionName, connectionPlugin := range connectionPluginMap {
//		schema := pluginSchemaMap[connectionPlugin.PluginName]
//		schemaModeMap[connectionName] = schema.Mode
//	}
//	return schemaModeMap
//}

func createConnectionPlugin(connection *modconfig.Connection, reattach *proto.ReattachConfig) (*ConnectionPlugin, error) {

	log.Printf("[TRACE] createConnectionPlugin for connection %s", connection.Name)
	pluginName := connection.Plugin
	connectionName := connection.Name
	connectionConfig := connection.Config
	connectionOptions := connection.Options

	log.Printf("[TRACE] plugin manager returned reattach config for connection '%s' - pid %d",
		connectionName, reattach.Pid)
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

	log.Printf("[TRACE] plugin client created for %s", pluginName)

	// now create ConnectionPlugin object return
	connectionPlugin := NewConnectionPlugin(pluginName, pluginClient, reattach.SupportedOperations)

	// if multiple connections are NOT supported, add the config for our one and only connection
	if reattach.SupportedOperations == nil || !reattach.SupportedOperations.MultipleConnections {
		log.Printf("[TRACE] multiple connections NOT supported - adding single connection '%s' to ConnectionPlugin", connectionName)
		connectionPlugin.addConnection(connectionName, connectionConfig, connectionOptions)
	} else {
		log.Printf("[TRACE] multiple connections ARE supported - adding all connections to ConnectionPlugin: %v", reattach.Connections)
		// now identify all connections serviced by this plugin
		for _, c := range reattach.Connections {
			log.Printf("[TRACE] adding connection %s", c)

			// NOTE: use GlobalConfig to access connection config
			// we assume this has been populated either by the hub (if this is being invoked from the fdw) or the CLI
			config, ok := GlobalConfig.Connections[c]
			if !ok {
				return nil, fmt.Errorf("no connection config loaded for '%s'", c)
			}
			connectionPlugin.addConnection(c, config.Config, config.Options)
		}
	}

	log.Printf("[TRACE] created connection plugin for connection: '%s', pluginName: '%s'", connectionName, pluginName)
	return connectionPlugin, nil
}

// use the reattach config to create a PluginClient for the plugin
func attachToPlugin(reattach *plugin.ReattachConfig, pluginName string) (*sdkgrpc.PluginClient, error) {
	return sdkgrpc.NewPluginClientFromReattach(reattach, pluginName)
}
