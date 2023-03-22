package pluginmanager_service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	sdkshared "github.com/turbot/steampipe-plugin-sdk/v5/grpc/shared"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/connectionwatcher"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pluginmanager_service/grpc/proto"
	pluginshared "github.com/turbot/steampipe/pluginmanager_service/grpc/shared"
)

type runningPlugin struct {
	pluginName  string
	client      *plugin.Client
	reattach    *proto.ReattachConfig
	initialized chan struct{}
}

// PluginManager is the implementation of grpc.PluginManager
type PluginManager struct {
	proto.UnimplementedPluginManagerServer

	// map of multi connection running plugins keyed by plugin name
	pluginMultiConnectionMap map[string]*runningPlugin
	// TACTICAL
	// until a plugin has loaded we do not know if it supports multi connection or not
	// keep the runningPlugin in this map until it is loaded to avoid race condition
	// starting multiple connections for a multi-connection plugin
	// (keyed by plugin name)
	loadingPlugins map[string]*runningPlugin

	// map of ALL running plugins keyed by connection name
	connectionPluginMap map[string]*runningPlugin
	// map of connection configs, keyed by plugin name
	// NOTE - for legacy plugins, one entry in this map may correspond to multiple running plugins
	pluginConnectionConfigMap map[string][]*sdkproto.ConnectionConfig
	// map of connection configs, keyed by connection name
	connectionConfigMap connectionwatcher.ConnectionConfigMap
	// map of max cache size, keyed by plugin name
	pluginCacheSizeMap map[string]int64

	mut           sync.Mutex
	logger        hclog.Logger
	messageServer *PluginMessageServer
}

func NewPluginManager(connectionConfig map[string]*sdkproto.ConnectionConfig, logger hclog.Logger) (*PluginManager, error) {
	log.Printf("[TRACE] NewPluginManager")
	pluginManager := &PluginManager{
		logger:                   logger,
		pluginMultiConnectionMap: make(map[string]*runningPlugin),
		loadingPlugins:           make(map[string]*runningPlugin),
		connectionPluginMap:      make(map[string]*runningPlugin),
		connectionConfigMap:      connectionConfig,
		// pluginConnectionConfigMap is created by populatePluginConnectionConfigs

	}
	messageServer, err := NewPluginMessageServer(pluginManager)
	if err != nil {
		return nil, err
	}
	pluginManager.messageServer = messageServer

	// populate plugin connection config map
	pluginManager.populatePluginConnectionConfigs()
	// determine cache size for each plugin
	pluginManager.setPluginCacheSizeMap()

	return pluginManager, nil
}

// plugin interface functions

func (m *PluginManager) Serve() {
	// create a plugin map, using ourselves as the implementation
	pluginMap := map[string]plugin.Plugin{
		pluginshared.PluginName: &pluginshared.PluginManagerPlugin{Impl: m},
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginMap,
		//  enable gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

func (m *PluginManager) Get(req *proto.GetRequest) (*proto.GetResponse, error) {
	resp := &proto.GetResponse{
		ReattachMap: make(map[string]*proto.ReattachConfig),
		FailureMap:  make(map[string]string),
	}

	// build a map of plugins required
	var plugins = make(map[string][]*sdkproto.ConnectionConfig)
	for _, connectionName := range req.Connections {
		connectionConfig, err := m.getConnectionConfig(connectionName)
		if err != nil {
			return nil, err
		}
		pluginName := connectionConfig.Plugin
		plugins[pluginName] = append(plugins[pluginName], connectionConfig)
	}

	log.Printf("[TRACE] PluginManager Get, connections: '%s'\n", req.Connections)
	for pluginName, connectionConfigs := range plugins {
		// have we already tried and failed to load this plugin - if so skip
		if _, pluginAlreadyFailed := resp.FailureMap[pluginName]; pluginAlreadyFailed {
			continue
		}
		// ensure plugin is running
		err := m.ensurePlugin(pluginName, connectionConfigs)
		if err != nil {
			resp.FailureMap[pluginName] = err.Error()
		} else {
			// the running plugin will have been populated in connectionPluginMap for all connections
			// copy reattach into responses
			for _, config := range connectionConfigs {
				resp.ReattachMap[config.Connection] = m.connectionPluginMap[config.Connection].reattach
			}
		}
	}

	return resp, nil
}

// OnConnectionConfigChanged is the callback function invoked by the connection watcher when the config changed
func (m *PluginManager) OnConnectionConfigChanged(configMap connectionwatcher.ConnectionConfigMap) {
	m.mut.Lock()
	defer m.mut.Unlock()

	names := utils.SortedMapKeys(configMap)
	log.Printf("[TRACE] OnConnectionConfigChanged: %s", strings.Join(names, ","))

	err := m.handleConnectionConfigChanges(configMap)
	if err != nil {
		log.Printf("[WARN] handleConnectionConfigChanges returned error: %s", err.Error())
	}

}

// OnConnectionsChanged is the callback function invoked by the connection watcher when connections are added or removed
func (m *PluginManager) OnConnectionsChanged(ctx context.Context, refreshResult *steampipeconfig.RefreshConnectionResult) {
	notification := refreshResult.Updates.AsNotification()
	m.notifySchemaChange(ctx, notification)
}

func (m *PluginManager) Shutdown(*proto.ShutdownRequest) (resp *proto.ShutdownResponse, err error) {
	log.Printf("[INFO] PluginManager Shutdown")

	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	for name, p := range m.connectionPluginMap {
		if p.client == nil {
			log.Printf("[WARN] plugin %s has no client - cannot kill", name)
			// shouldn't happen but has been observed in error situations
			continue
		}
		log.Printf("[INFO] PluginManager killing plugin %s (%v)", name, p.reattach.Pid)
		p.client.Kill()
	}
	return &proto.ShutdownResponse{}, nil
}

func (m *PluginManager) handleConnectionConfigChanges(newConfigMap map[string]*sdkproto.ConnectionConfig) error {
	// now determine whether there are any new or deleted connections
	addedConnections, deletedConnections, changedConnections := m.connectionConfigMap.Diff(newConfigMap)

	requestMap := make(map[string]*sdkproto.UpdateConnectionConfigsRequest)

	// for deleted connections, remove from plugins and pluginConnectionConfigs
	m.handleDeletedConnections(deletedConnections, requestMap)

	// for new connections, add to plugins and pluginConnectionConfigs
	m.handleAddedConnections(addedConnections, requestMap)
	// for updated connections just add to request map
	m.handleUpdatedConnections(changedConnections, requestMap)
	// update connectionConfigMap
	m.connectionConfigMap = newConfigMap

	// rebuild pluginConnectionConfigMap
	m.populatePluginConnectionConfigs()

	// now send UpdateConnectionConfigs for all update plugins
	return m.sendUpdateConnectionConfigs(requestMap)
}

func (m *PluginManager) sendUpdateConnectionConfigs(requestMap map[string]*sdkproto.UpdateConnectionConfigsRequest) error {
	var errors []error
	for plugin, req := range requestMap {
		runningPlugin, pluginAlreadyRunning := m.pluginMultiConnectionMap[plugin]
		// TODO what if the plugin crashed - should we restart here?
		// if the plugin is not running (or is not multi connection, so is not in this map), return
		if !pluginAlreadyRunning {
			continue
		}

		pluginClient, err := sdkgrpc.NewPluginClient(runningPlugin.client, plugin)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		err = pluginClient.UpdateConnectionConfigs(req)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return error_helpers.CombineErrors(errors...)
}

// this mutates requestMap
func (m *PluginManager) handleAddedConnections(addedConnections map[string][]*sdkproto.ConnectionConfig, requestMap map[string]*sdkproto.UpdateConnectionConfigsRequest) {
	// for new connections, add to plugins , pluginConnectionConfigs and connectionConfig
	// (but only if the plugin is already started - if not we do nothing here - refreshConnections will start the plugin)
	for p, connections := range addedConnections {
		// find the existing running plugin for this plugin
		// if this plugins is NOT running (or is not multi connection), skip here - we will start it when running refreshConnections
		runningPlugin, pluginAlreadyRunning := m.pluginMultiConnectionMap[p]
		if !pluginAlreadyRunning {
			log.Printf("[TRACE] handleAddedConnections - plugin '%s' has been added to connection config and is not running - doing nothing here as it will be started by refreshConnections", p)
			continue
		}

		// get or create req for this plugin
		req, ok := requestMap[p]
		if !ok {
			req = &sdkproto.UpdateConnectionConfigsRequest{}
		}

		for _, connection := range connections {
			// add this connection to the running plugin
			runningPlugin.reattach.AddConnection(connection.Connection)

			// add to updateConnectionConfigsRequest
			req.Added = append(req.Added, connection)

			// add this connection to connection-running plugin map
			m.connectionPluginMap[connection.Connection] = runningPlugin
		}
		// write back to map
		requestMap[p] = req
	}
}

// this mutates requestMap
func (m *PluginManager) handleDeletedConnections(deletedConnections map[string][]*sdkproto.ConnectionConfig, requestMap map[string]*sdkproto.UpdateConnectionConfigsRequest) {
	for p, connections := range deletedConnections {
		runningPlugin, pluginAlreadyRunning := m.pluginMultiConnectionMap[p]
		if !pluginAlreadyRunning {
			continue
		}

		// get or create req for this plugin
		req, ok := requestMap[p]
		if !ok {
			req = &sdkproto.UpdateConnectionConfigsRequest{}
		}

		for _, connection := range connections {
			// remove this connection from the running plugin
			runningPlugin.reattach.RemoveConnection(connection.Connection)

			// add to updateConnectionConfigsRequest
			req.Deleted = append(req.Deleted, connection)

			// remove this connection from connection plugin map
			delete(m.connectionPluginMap, connection.Connection)
		}
		// write back to map
		requestMap[p] = req
	}
}

// this mutates requestMap
func (m *PluginManager) handleUpdatedConnections(updatedConnections map[string][]*sdkproto.ConnectionConfig, requestMap map[string]*sdkproto.UpdateConnectionConfigsRequest) {
	// for new connections, add to plugins , pluginConnectionConfigs and connectionConfig
	// (but only if the plugin is already started - if not we do nothing here - refreshConnections will start the plugin)
	for p, connections := range updatedConnections {
		// get or create req for this plugin
		req, ok := requestMap[p]
		if !ok {
			req = &sdkproto.UpdateConnectionConfigsRequest{}
		}

		for _, connection := range connections {
			// add to updateConnectionConfigsRequest
			req.Changed = append(req.Changed, connection)
		}
		// write back to map
		requestMap[p] = req
	}
}

func (m *PluginManager) getConnectionConfig(connectionName string) (*sdkproto.ConnectionConfig, error) {
	connectionConfig, ok := m.connectionConfigMap[connectionName]
	if !ok {
		return nil, fmt.Errorf("plugin manager: no connection config loaded for connection '%s'", connectionName)
	}
	return connectionConfig, nil
}

func (m *PluginManager) ensurePlugin(pluginName string, connectionConfigs []*sdkproto.ConnectionConfig) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	for _, connectionConfig := range connectionConfigs {
		connectionName := connectionConfig.Connection

		log.Printf("[TRACE] PluginManager ensurePlugin %s connection '%s'\n", pluginName, connectionName)

		// reason for starting the plugin (if we need to)
		var reason string

		// is this plugin already running
		// lock access to plugin map
		m.mut.Lock()
		runningPlugin := m.isPluginRunning(connectionName, pluginName)
		// do we now have a plugin?
		if runningPlugin != nil {
			// unlock access to map to allow other ensurePlugin calls to proceed
			m.mut.Unlock()
			var reattach *proto.ReattachConfig

			// wait for plugin to load, verify it is running and check it provides the required connection
			reason, reattach, err = m.verifyLoadingPlugin(connectionName, runningPlugin)
			if err != nil {
				return err
			}
			if reason == "" {
				// so we have a reattach - if this plugin handles multiple connections, we are done
				if reattach.SupportedOperations.MultipleConnections {
					return nil
				}
				// this must be a legacy plugin - keep looping
				continue
			}

			// so we have not yet found a compatible plugin

			// NOTE: re-lock the mutex before falling through to addLoadingPlugin
			m.mut.Lock()

			// TACTICAL there is a race condition here - multiple threads may be here at the same time
			// check whether another thread has one and started loading the required plugin
			// recheck the connection map
			p, ok := m.connectionPluginMap[connectionName]
			if ok {
				// unlock before calling verifyLoadingPlugin
				m.mut.Unlock()

				log.Printf("[TRACE] after waiting for plugin %s to load, and discovering it does not support connection %s, found a loading plugin in connectionPluginMap, so using that", pluginName, connectionName)
				reason, reattach, err = m.verifyLoadingPlugin(connectionName, p)
				if err != nil {
					return err
				}

				if reason == "" {
					log.Printf("[TRACE] now we have one")
					// so we have a reattach - if this plugin handles multiple connections, we are done
					if reattach.SupportedOperations.MultipleConnections {
						return nil
					}
					// this must be a legacy plugin - keep looping
					continue
				}
				// relock
				m.mut.Lock()
			}

			//  either the pid does not exist or the plugin has exited

		} else {
			// so the plugin is NOT loaded or loading - this is the first time anyone has requested this plugin
			reason = fmt.Sprintf("PluginManager %p plugin %s (%s) NOT started or starting - start now", m, pluginName, connectionName)
		}

		// to get here, for whatever reason, we need to start the plugin

		// NOTE: at this point, m.mut is locked
		// put in a placeholder so no other thread tries to create start this plugin
		m.addLoadingPlugin(connectionName, pluginName)

		// unlock access to map
		m.mut.Unlock()

		// NOTE: It is an error to try to start a plugin which is already running
		// this may happen if the file watcher has been triggered by a connection being added for an existing plugin
		// if this happened, the plugin manager should ALREADY have called UpdateConnectionConfig to send the config
		// for the new connection to the plugin

		// fall through to plugin startup
		// log the startup reason
		log.Printf("[TRACE] %s", reason)
		// so we need to start the plugin
		client, reattach, err := m.startPlugin(connectionName)
		if err != nil {
			m.mut.Lock()
			delete(m.connectionPluginMap, connectionName)
			m.mut.Unlock()

			log.Println("[TRACE] startPlugin failed with", err)
			return err
		}

		// store the client to our map
		m.storePluginToMap(connectionName, client, reattach)

		log.Printf("[TRACE] PluginManager ensurePlugin complete, returning reattach config with PID: %d", reattach.Pid)
	}
	// and return
	return nil
}

func (m *PluginManager) isPluginRunning(connectionName string, pluginName string) *runningPlugin {
	p, ok := m.connectionPluginMap[connectionName]
	if ok {
		log.Printf("[TRACE] connection %s found in connectionPluginMap\n", connectionName)
		return p
	}
	// so there is no entry in connectionPluginMap for this connection - check whether there is an entry in either
	// - pluginMultiConnectionMap (indicating this is a multi connection plugin which has been loaded for another connection
	// - loadingPlugins (indicating this is a plugin which is still loading and we do not yet know if it supports multi connection
	p, ok = m.pluginMultiConnectionMap[pluginName]
	if ok {
		log.Printf("[TRACE] %s found in pluginMultiConnectionMap\n", pluginName)
		return p
	}
	p, ok = m.loadingPlugins[pluginName]
	if ok {
		log.Printf("[TRACE] %s found in loadingPlugins\n", pluginName)
		return p
	}

	return nil
}

// wait for plugin to load, verify it is running and check it provides the required connection
func (m *PluginManager) verifyLoadingPlugin(connectionName string, p *runningPlugin) (string, *proto.ReattachConfig, error) {
	var reason string
	// so we have a plugin in our map for this connection - is it started?
	err := m.waitForPluginLoad(p)
	if err != nil {
		return "", nil, err
	}
	log.Printf("[TRACE] connection %s is loaded, check for running PID", connectionName)

	// ok so the plugin _should_ now be running

	// check if this plugin provides this connection
	// this should always be the case for multiconnection plugins but may not be the case for legacy plugins
	reattach := p.reattach
	if helpers.StringSliceContains(p.reattach.Connections, connectionName) {
		// now check if the plugins process IS running
		exists, _ := utils.PidExists(int(reattach.Pid))
		if exists {
			// so the plugin is good
			log.Printf("[TRACE] PluginManager found '%s' in map, pid %d", connectionName, reattach.Pid)
			return "", reattach, nil
		} else {
			// otherwise we need to start the plugin again -  update reason
			reason = fmt.Sprintf("PluginManager found pid %d for connection '%s' in plugin map but plugin process does not exist - killing client and removing from map", reattach.Pid, connectionName)
		}
	} else {
		// so the plugin does not support this connection (must be a legacy plugin)

		// update reason
		reason = fmt.Sprintf("plugin %s does NOT provide connection %s", p.reattach.Plugin, connectionName)
	}
	return reason, nil, nil
}

func (m *PluginManager) addLoadingPlugin(connectionName string, pluginName string) {
	// add a new running plugin to both connectionPluginMap and pluginMap
	// NOTE: m.mut must be locked before calling this
	p := &runningPlugin{
		pluginName:  pluginName,
		initialized: make(chan struct{}, 1),
	}
	m.connectionPluginMap[connectionName] = p
	// also add to loadingPlugins
	m.loadingPlugins[pluginName] = p
}

// create reattach config for plugin, store to map for all connections and close initialized channel
func (m *PluginManager) storePluginToMap(connection string, client *plugin.Client, reattach *proto.ReattachConfig) {
	// lock access to map
	m.mut.Lock()
	defer m.mut.Unlock()

	// a RunningPlugin in initializing state will already have been put into the Plugins map
	// populate its properties
	p := m.connectionPluginMap[connection]
	p.client = client
	p.reattach = reattach

	// store fully initialised runningPlugin to pluginMap
	if reattach.SupportedOperations.MultipleConnections {
		m.pluginMultiConnectionMap[reattach.Plugin] = p
	}
	// remove from loadingPlugins
	delete(m.loadingPlugins, reattach.Plugin)
	// NOTE: if this plugin supports multiple connections, reattach.Connections will be a list of all connections
	// provided by this plugin
	// add map entries for all other connections using this plugin (all pointing to same RunningPlugin)
	for _, c := range reattach.Connections {
		m.connectionPluginMap[c] = p
	}
	// mark as initialized
	close(p.initialized)
}

// populate map of connection configs for each plugin
func (m *PluginManager) populatePluginConnectionConfigs() {
	m.pluginConnectionConfigMap = make(map[string][]*sdkproto.ConnectionConfig)
	for _, config := range m.connectionConfigMap {
		m.pluginConnectionConfigMap[config.Plugin] = append(m.pluginConnectionConfigMap[config.Plugin], config)
	}
}

// populate map of connection configs for each plugin
func (m *PluginManager) setPluginCacheSizeMap() {
	m.pluginCacheSizeMap = make(map[string]int64, len(m.pluginConnectionConfigMap))

	// read the env var setting cache size
	maxCacheSizeMb, _ := strconv.Atoi(os.Getenv(constants.EnvCacheMaxSize))

	// get total connection count for this plugin (excluding aggregators)
	numConnections := m.nonAggregatorConnectionCount()

	log.Printf("[TRACE] PluginManager setPluginCacheSizeMap: %d %s.", numConnections, utils.Pluralize("connection", numConnections))
	log.Printf("[TRACE] Total cache size %dMb", maxCacheSizeMb)

	for plugin, connections := range m.pluginConnectionConfigMap {
		var size int64 = 0
		// if no max size is set, just set all plugins to zero (unlimited)
		if maxCacheSizeMb > 0 {
			// get connection count for this plugin (excluding aggregators)
			numPluginConnections := nonAggregatorConnectionCount(connections)
			size = int64(float64(numPluginConnections) / float64(numConnections) * float64(maxCacheSizeMb))
			// make this at least 1 Mb (as zero means unlimited)
			if size == 0 {
				size = 1
			}
			log.Printf("[WARN] Plugin '%s', %d %s, max cache size %dMb", plugin, numPluginConnections, utils.Pluralize("connection", numPluginConnections), size)
		}

		m.pluginCacheSizeMap[plugin] = size
	}
}

func (m *PluginManager) startPlugin(connectionName string) (_ *plugin.Client, _ *proto.ReattachConfig, err error) {
	log.Printf("[TRACE] ************ start plugin %s ********************\n", connectionName)

	// get connection config
	connectionConfig, err := m.getConnectionConfig(connectionName)
	if err != nil {
		return nil, nil, err
	}

	pluginPath, err := filepaths.GetPluginPath(connectionConfig.Plugin, connectionConfig.PluginShortName)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("[TRACE] ************ plugin path %s ********************\n", pluginPath)

	// create the plugin map
	pluginName := connectionConfig.Plugin
	pluginMap := map[string]plugin.Plugin{
		pluginName: &sdkshared.WrapperPlugin{},
	}

	utils.LogTime("getting plugin exec hash")
	pluginChecksum, err := helpers.FileMD5Hash(pluginPath)
	if err != nil {
		return nil, nil, err
	}
	utils.LogTime("got plugin exec hash")
	cmd := exec.Command(pluginPath)
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  sdkshared.Handshake,
		Plugins:          pluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		SecureConfig: &plugin.SecureConfig{
			Checksum: pluginChecksum,
			Hash:     md5.New(),
		},
		// pass our logger to the plugin client to ensure plugin logs end up in logfile
		Logger: m.logger,
	})

	if _, err := client.Start(); err != nil {
		return nil, nil, err
	}

	// ensure we shut down in case of failure
	defer func() {
		if err != nil {
			// we failed - shut down the plugin again
			client.Kill()
		}
	}()

	// get the supported operations
	pluginClient, err := sdkgrpc.NewPluginClient(client, pluginName)
	if err != nil {
		return nil, nil, err
	}

	// fetch the supported operations
	supportedOperations, _ := pluginClient.GetSupportedOperations()
	// ignore errors  - just create an empty support structure if needed
	if supportedOperations == nil {
		supportedOperations = &sdkproto.GetSupportedOperationsResponse{}
	}

	log.Printf("[TRACE] supportedOperations: %v", supportedOperations)
	var connectionNames = []string{connectionName}

	if supportedOperations.MultipleConnections {
		// send the connection config for all connections for this plugin
		// this returns a list of all connections provided by this plugin
		connectionNames, err = m.setAllConnectionConfigs(pluginClient, pluginName)
	} else {
		// send the connection config using legacy single connection function
		err = m.setSingleConnectionConfig(pluginClient, connectionName)
	}
	if err != nil {
		log.Printf("[WARN] failed to set connection config: %s", err.Error())
		return nil, nil, err
	}

	reattach := proto.NewReattachConfig(pluginName, client.ReattachConfig(), proto.SupportedOperationsFromSdk(supportedOperations), connectionNames)

	//
	// if this plugin has a dynamic schema, add connections to message server
	schema, err := pluginClient.GetSchema(connectionName)
	if err != nil {
		log.Printf("[WARN] failed to set fetch schema: %s", err.Error())
		return nil, nil, err
	}

	if schema.Mode == sdkplugin.SchemaModeDynamic {
		_ = m.messageServer.AddConnection(pluginClient, pluginName, connectionNames...)
	}
	// TODO KAI how to handle error here
	return client, reattach, nil
}

func (m *PluginManager) waitForPluginLoad(p *runningPlugin) error {
	pluginStartTimeoutSecs := 5

	select {
	case <-p.initialized:
		log.Printf("[TRACE] initialized: %d", p.reattach.Pid)
		log.Printf("[TRACE] initialized: %d", p.reattach.Pid)
		return nil

	case <-time.After(time.Duration(pluginStartTimeoutSecs) * time.Second):
		return fmt.Errorf("timed out waiting for %s to startup after %d seconds", p.pluginName, pluginStartTimeoutSecs)
	}
}

func (m *PluginManager) getConnectionsForPlugin(pluginName string) []string {
	var res = make([]string, len(m.pluginConnectionConfigMap[pluginName]))
	for i, c := range m.pluginConnectionConfigMap[pluginName] {
		res[i] = c.Connection
	}
	return res
}

// set connection config for multiple connection, for compatible plugins
// NOTE: we DO NOT set connection config for aggregator connections
func (m *PluginManager) setAllConnectionConfigs(pluginClient *sdkgrpc.PluginClient, pluginName string) ([]string, error) {
	configs, ok := m.pluginConnectionConfigMap[pluginName]
	if !ok {
		// should never happen
		return nil, fmt.Errorf("no config loaded for plugin '%s'", pluginName)
	}
	req := &sdkproto.SetAllConnectionConfigsRequest{
		Configs:        configs,
		MaxCacheSizeMb: m.pluginCacheSizeMap[pluginName],
	}
	// build list of connections
	connections := make([]string, len(configs))
	for i, config := range configs {
		connections[i] = config.Connection
	}
	_, err := pluginClient.SetAllConnectionConfigs(req)
	return connections, err
}

// set connection config for single connection, for legacy plugins)
func (m *PluginManager) setSingleConnectionConfig(pluginClient *sdkgrpc.PluginClient, connectionName string) error {
	connectionConfig, err := m.getConnectionConfig(connectionName)
	if err != nil {
		return err
	}
	// set the connection config
	req := &sdkproto.SetConnectionConfigRequest{
		ConnectionName:   connectionName,
		ConnectionConfig: connectionConfig.Config,
	}

	return pluginClient.SetConnectionConfig(req)
}

// update the schema for the specified connection
// called from the message server after receiving a PluginMessageType_SCHEMA_UPDATED message from plugin
func (m *PluginManager) updateConnectionSchema(ctx context.Context, connectionName string) {
	log.Printf("[TRACE] updateConnectionSchema connection %s", connectionName)

	refreshResult := db_local.RefreshConnectionAndSearchPaths(ctx, connectionName)
	if refreshResult.Error != nil {
		log.Printf("[TRACE] error refreshing connections: %s", refreshResult.Error)
		return
	}

	// also send a postgres notification
	notification := steampipeconfig.NewSchemaUpdateNotification([]string{connectionName}, nil)

	m.notifySchemaChange(ctx, notification)
}

// send a postgres notification that the schema has chganged
func (m *PluginManager) notifySchemaChange(ctx context.Context, notification any) {
	conn, err := db_local.CreateLocalDbConnection(ctx, &db_local.CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {

	}

	notificationBytes, err := json.Marshal(notification)
	if err != nil {
		log.Printf("[TRACE] error marshalling Postgres notification: %s", err.Error())
		return
	}

	log.Printf("[TRACE] Send update notification")

	sql := fmt.Sprintf("select pg_notify('%s', $1)", constants.PostgresNotificationChannel)
	_, err = conn.Exec(context.Background(), sql, notificationBytes)
	if err != nil {
		log.Printf("[WARN] Error sending notification: %s", err)
	}
}

func (m *PluginManager) nonAggregatorConnectionCount() int {
	res := 0
	for _, connections := range m.pluginConnectionConfigMap {
		res += nonAggregatorConnectionCount(connections)
	}
	return res
}

func nonAggregatorConnectionCount(connections []*sdkproto.ConnectionConfig) int {
	res := 0
	for _, c := range connections {
		if len(c.ChildConnections) == 0 {
			res++
		}
	}
	return res

}
