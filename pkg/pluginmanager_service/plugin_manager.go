package pluginmanager_service

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/sethvargo/go-retry"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	sdkshared "github.com/turbot/steampipe-plugin-sdk/v5/grpc/shared"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/connection"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	pb "github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/proto"
	pluginshared "github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/shared"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

type runningPlugin struct {
	pluginName  string
	client      *plugin.Client
	reattach    *pb.ReattachConfig
	initialized chan struct{}
	failed      chan struct{}
	error       error
}

// PluginManager is the implementation of grpc.PluginManager
type PluginManager struct {
	pb.UnimplementedPluginManagerServer

	// map of multi connection running plugins keyed by plugin name
	pluginMultiConnectionMap map[string]*runningPlugin

	// TACTICAL
	// until a plugin has loaded we do not know if it supports multi connection or not
	// keep the runningPlugin in this map until it is loaded to avoid race condition
	// starting multiple connections for a multi-connection plugin
	// (keyed by plugin name)
	//loadingPlugins map[string]*runningPlugin

	// map of ALL running plugins keyed by connection name
	//connectionPluginMap map[string]*runningPlugin

	// map of connection configs, keyed by plugin name
	pluginConnectionConfigMap map[string][]*sdkproto.ConnectionConfig
	// map of connection configs, keyed by connection name
	connectionConfigMap connection.ConnectionConfigMap
	// map of max cache size, keyed by plugin name
	pluginCacheSizeMap map[string]int64

	// map lock
	mut sync.RWMutex

	// shutdown syncronozation
	// do not start any plugins while shutting down
	shutdownMut sync.Mutex
	// do not shutdown until all plugins have loaded
	startPluginWg sync.WaitGroup

	logger        hclog.Logger
	messageServer *PluginMessageServer
}

func NewPluginManager(connectionConfig map[string]*sdkproto.ConnectionConfig, logger hclog.Logger) (*PluginManager, error) {
	log.Printf("[TRACE] NewPluginManager")
	pluginManager := &PluginManager{
		logger:                   logger,
		pluginMultiConnectionMap: make(map[string]*runningPlugin),
		//loadingPlugins:           make(map[string]*runningPlugin),
		//connectionPluginMap:      make(map[string]*runningPlugin),
		connectionConfigMap: connectionConfig,
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

func (m *PluginManager) Get(req *pb.GetRequest) (*pb.GetResponse, error) {
	resp := &pb.GetResponse{
		ReattachMap: make(map[string]*pb.ReattachConfig),
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

	// TODO make sure we get all connections for each plugin????

	log.Printf("[TRACE] PluginManager Get, connections: '%s'\n", req.Connections)
	for pluginName, connectionConfigs := range plugins {
		// have we already tried and failed to load this plugin - if so skip
		if _, pluginAlreadyFailed := resp.FailureMap[pluginName]; pluginAlreadyFailed {
			continue
		}
		// ensure plugin is running
		reattach, err := m.ensurePlugin(pluginName, connectionConfigs)
		if err != nil {
			resp.FailureMap[pluginName] = err.Error()
		} else {
			// the running plugin will have been populated in connectionPluginMap for all connections
			// copy reattach into responses
			for _, config := range connectionConfigs {
				resp.ReattachMap[config.Connection] = reattach
			}
		}
	}

	return resp, nil
}

func (m *PluginManager) RefreshConnections(*pb.RefreshConnectionsRequest) (*pb.RefreshConnectionsResponse, error) {
	resp := &pb.RefreshConnectionsResponse{}
	refreshResult := connection.RefreshConnections(context.Background())
	if refreshResult.Error != nil {
		return nil, refreshResult.Error
	}

	return resp, nil
}

// OnConnectionConfigChanged is the callback function invoked by the connection watcher when the config changed
func (m *PluginManager) OnConnectionConfigChanged(configMap connection.ConnectionConfigMap) {
	m.mut.Lock()
	defer m.mut.Unlock()

	names := utils.SortedMapKeys(configMap)
	log.Printf("[TRACE] OnConnectionConfigChanged: %s", strings.Join(names, ","))

	err := m.handleConnectionConfigChanges(configMap)
	if err != nil {
		log.Printf("[WARN] handleConnectionConfigChanges returned error: %s", err.Error())
	}

}

func (m *PluginManager) Shutdown(*pb.ShutdownRequest) (resp *pb.ShutdownResponse, err error) {
	log.Printf("[INFO] PluginManager Shutdown")

	// lock shutdownMut before waiting for startPluginWg
	// this enables us to exit from ensurePlugin early if needed
	m.shutdownMut.Lock()
	log.Printf("[TRACE] locked shutdownMut, waiting for startPluginWg")
	m.startPluginWg.Wait()
	log.Printf("[TRACE] waited for startPluginWg, locking mut")
	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	// kill all plugins in pluginMultiConnectionMap
	for _, p := range m.pluginMultiConnectionMap {

		log.Printf("[INFO] Kill plugin %s (%p)", p.pluginName, p.client)
		m.killPlugin(p)
	}

	return &pb.ShutdownResponse{}, nil
}

func (m *PluginManager) killPlugin(p *runningPlugin) {
	if p.client == nil {
		log.Printf("[WARN] plugin %s has no client - cannot kill client", p.pluginName)
		// shouldn't happen but has been observed in error situations
		return
	}
	log.Printf("[INFO] PluginManager killing plugin %s (%v)", p.pluginName, p.reattach.Pid)
	p.client.Kill()
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

		// add to updateConnectionConfigsRequest
		req.Changed = append(req.Changed, connections...)
		// write back to map
		requestMap[p] = req
	}
}

func (m *PluginManager) getConnectionConfig(connectionName string) (*sdkproto.ConnectionConfig, error) {
	connectionConfig, ok := m.connectionConfigMap[connectionName]
	if !ok {
		return nil, fmt.Errorf("connection '%s' does not exist in connection config", connectionName)
	}
	return connectionConfig, nil
}

func (m *PluginManager) ensurePlugin(pluginName string, connectionConfigs []*sdkproto.ConnectionConfig) (reattach *pb.ReattachConfig, err error) {
	backoff := retry.WithMaxRetries(5, retry.NewConstant(10*time.Millisecond))

	// ensure we do not shutdown until this has finished
	m.startPluginWg.Add(1)
	defer func() {
		m.startPluginWg.Done()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	// do not install a plugin while shutting down
	if m.shuttingDown() {
		return nil, fmt.Errorf("plugin manager is shutting down")
	}

	log.Printf("[INFO] PluginManager ensurePlugin %s", pluginName)

	err = retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		reattach, err = m.startPluginIfNeeded(pluginName, connectionConfigs)
		return err
	})

	return
}

func (m *PluginManager) startPluginIfNeeded(pluginName string, connectionConfigs []*sdkproto.ConnectionConfig) (*pb.ReattachConfig, error) {
	// is this plugin already running
	// lock access to plugin map
	log.Printf("[INFO] startPluginIfNeeded getting lock")
	m.mut.RLock()
	log.Printf("[INFO] startPluginIfNeeded got lock reading map")
	startingPlugin, ok := m.pluginMultiConnectionMap[pluginName]
	m.mut.RUnlock()
	log.Printf("[INFO] startPluginIfNeeded released lock")

	if ok {
		log.Printf("[INFO] got running plugin")

		// wait for plugin to process connection config, and verify it is running
		err := m.waitForPluginLoad(startingPlugin)
		if err == nil {
			// so plugin has loaded - we are done
			log.Printf("[INFO] found running plugin %s", pluginName)
			return startingPlugin.reattach, nil
		}
		log.Printf("[INFO] waitForPluginLoad failed %s", err.Error())

		// just return the error
		return nil, err
	}

	// so the plugin is NOT loaded or loading
	// fall through to plugin startup
	log.Printf("[WARN] PluginManager %p plugin %s NOT started or starting - start now", m, pluginName)

	return m.startPlugin(pluginName, connectionConfigs)
}

func (m *PluginManager) startPlugin(pluginName string, connectionConfigs []*sdkproto.ConnectionConfig) (_ *pb.ReattachConfig, err error) {
	// add a new running plugin to pluginMultiConnectionMap
	startingPlugin, err := m.addRunningPlugin(pluginName)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] release lock ")

	log.Printf("[INFO] start plugin")
	// now start the process
	client, err := m.startPluginProcess(pluginName, connectionConfigs)

	defer func() {
		if err != nil {
			m.mut.Lock()
			m.mut.Unlock()

			// delete from map
			delete(m.pluginMultiConnectionMap, pluginName)
			// close failed chan
			close(startingPlugin.failed)

			log.Println("[WARN] startPluginProcess failed with", err)
		}
	}()
	reattach, err := m.initializePlugin(connectionConfigs, client)

	log.Printf("[INFO] assign reattach and client")
	startingPlugin.reattach = reattach
	startingPlugin.client = client
	log.Printf("[INFO] assign reattached and client")

	// TODO INVESTIGATE connectionConfigMap
	log.Printf("[INFO] store connection configs")
	m.mut.Lock()
	for _, connectionConfig := range connectionConfigs {
		m.connectionConfigMap[connectionConfig.Connection] = connectionConfig
	}
	// unlock
	m.mut.Unlock()
	log.Printf("[INFO] stored connection configs")

	log.Printf("[INFO] close init chan")
	// close initialized chan
	close(startingPlugin.initialized)
	log.Printf("[INFO] closed init chan")

	log.Printf("[TRACE] PluginManager ensurePlugin complete, returning reattach config with PID: %d", reattach.Pid)

	// and return
	return reattach, nil
}

func (m *PluginManager) addRunningPlugin(pluginName string) (*runningPlugin, error) {
	// add a new running plugin to pluginMultiConnectionMap
	// this is a placeholder so no other thread tries to create start this plugin

	// acquire write lock
	m.mut.Lock()
	defer m.mut.Unlock()
	log.Printf("[INFO] starting plugin %s (if someone didn't beat us to it)", pluginName)

	log.Printf("[INFO] startPlugin got lock ")
	log.Printf("[INFO] rechecking ")

	// check someone else has beaten us to it (there is a race condition to staring a plugin)
	if _, ok := m.pluginMultiConnectionMap[pluginName]; ok {
		log.Printf("[INFO] re checked map and found a starting plugin - retrying")
		// if so, just retry, which will wait for the loading plugin
		return nil, retry.RetryableError(fmt.Errorf("another client has already started the plugin"))
	}

	// create the running plugin
	startingPlugin := &runningPlugin{
		pluginName:  pluginName,
		initialized: make(chan struct{}),
		failed:      make(chan struct{}),
	}
	// write back
	m.pluginMultiConnectionMap[pluginName] = startingPlugin
	log.Printf("[INFO] written running plugin to map")

	return startingPlugin, nil
}

func (m *PluginManager) startPluginProcess(pluginName string, connectionConfigs []*sdkproto.ConnectionConfig) (*plugin.Client, error) {
	log.Printf("[INFO] ************ start plugin %s ********************\n", pluginName)

	exemplarConnectionConfig := connectionConfigs[0]
	pluginPath, err := filepaths.GetPluginPath(pluginName, exemplarConnectionConfig.PluginShortName)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] ************ plugin path %s ********************\n", pluginPath)

	// create the plugin map
	pluginMap := map[string]plugin.Plugin{
		pluginName: &sdkshared.WrapperPlugin{},
	}

	utils.LogTime("getting plugin exec hash")
	pluginChecksum, err := helpers.FileMD5Hash(pluginPath)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return client, nil

}

// set the connection configs and build a ReattachConfig
func (m *PluginManager) initializePlugin(connectionConfigs []*sdkproto.ConnectionConfig, client *plugin.Client) (_ *pb.ReattachConfig, err error) {
	log.Printf("[WARN] initializePlugin pid %d", client.ReattachConfig().Pid)
	// ensure we shut down in case of failure
	defer func() {
		if err != nil {
			// we failed - shut down the plugin again
			client.Kill()
		}
	}()

	// extract connection names
	connectionNames := make([]string, len(connectionConfigs))
	for i, c := range connectionConfigs {
		connectionNames[i] = c.Connection
	}
	exemplarConnectionConfig := connectionConfigs[0]
	pluginName := exemplarConnectionConfig.Plugin

	// get the supported operations
	pluginClient, err := sdkgrpc.NewPluginClient(client, pluginName)
	if err != nil {
		return nil, err
	}

	// fetch the supported operations
	supportedOperations, _ := pluginClient.GetSupportedOperations()
	//// ignore errors  - just create an empty support structure if needed
	if supportedOperations == nil {
		supportedOperations = &sdkproto.GetSupportedOperationsResponse{}
	}

	// provide opportunity to avoid setting connection configs if we are shutting down
	if m.shuttingDown() {
		log.Printf("[INFO] aborting plugin %s startup - plugin manager is shutting down", pluginName)
		client.Kill()
		return nil, fmt.Errorf("plugin manager is shutting down")
	}

	// send the connection config for all connections for this plugin
	// this returns a list of all connections provided by this plugin
	err = m.setAllConnectionConfigs(connectionConfigs, pluginClient, supportedOperations)
	if err != nil {
		log.Printf("[WARN] failed to set connection config for %s: %s", pluginName, err.Error())
		return nil, err
	}

	// if this plugin supports setting cache options, do so
	if supportedOperations.SetCacheOptions {
		err = m.setCacheOptions(pluginClient)
		if err != nil {
			log.Printf("[WARN] failed to set cache options for %s: %s", pluginName, err.Error())
			return nil, err
		}
	}

	reattach := pb.NewReattachConfig(pluginName, client.ReattachConfig(), pb.SupportedOperationsFromSdk(supportedOperations), connectionNames)

	// if this plugin has a dynamic schema, add connections to message server
	err = m.notifyNewDynamicSchemas(pluginClient, exemplarConnectionConfig, connectionNames)
	if err != nil {
		// TODO how to handle error here
		client.Kill()
		// send err down  running plugin error channel
		return nil, err
	}

	log.Printf("[WARN] initializePlugin complete pid %d", client.ReattachConfig().Pid)
	return reattach, nil
}

// return whether the plugin manager is shutting down
func (m *PluginManager) shuttingDown() bool {
	m.mut.Lock()
	defer m.mut.Unlock()

	if !m.shutdownMut.TryLock() {
		return true
	}
	m.shutdownMut.Unlock()
	return false
}

//func (m *PluginManager) isPluginRunning(pluginName string) *runningPlugin {
//	p, ok := m.pluginMultiConnectionMap[pluginName]
//	if ok {
//		log.Printf("[TRACE] connection %s found in connectionPluginMap\n", connectionName)
//		return p
//	}
//	// so there is no entry in connectionPluginMap for this connection - check whether there is an entry in either
//	// - pluginMultiConnectionMap (indicating this is a multi connection plugin which has been loaded for another connection
//	// - loadingPlugins (indicating this is a plugin which is still loading and we do not yet know if it supports multi connection
//	p, ok = m.pluginMultiConnectionMap[pluginName]
//	if ok {
//		log.Printf("[TRACE] %s found in pluginMultiConnectionMap\n", pluginName)
//		return p
//	}
//	p, ok = m.loadingPlugins[pluginName]
//	if ok {
//		log.Printf("[TRACE] %s found in loadingPlugins\n", pluginName)
//		return p
//	}
//
//	return nil
//}

func (m *PluginManager) sigKillPlugin(reattach *pb.ReattachConfig) {
	log.Printf("[WARN] sending SIGKILL to plugin %s (pid %d)", reattach.Plugin, reattach.Pid)
	// kill to be certain
	err := syscall.Kill(int(reattach.Pid), syscall.SIGKILL)
	if err != nil {
		log.Printf("[WARN] failed to kill process %d: %s", reattach.Pid, err.Error())
	}
}

//func (m *PluginManager) addLoadingPlugin(connectionName string, pluginName string) {
//	// add a new running plugin to both connectionPluginMap and pluginMap
//	// NOTE: m.mut must be locked before calling this
//	p := &runningPlugin{
//		pluginName:  pluginName,
//		initialized: make(chan struct{}, 1),
//	}
//	m.connectionPluginMap[connectionName] = p
//	// also add to loadingPlugins
//	m.loadingPlugins[pluginName] = p
//}

//// create reattach config for plugin, store to map for all connections and close initialized channel
//func (m *PluginManager) storePluginToMap(connection string, client *plugin.Client, reattach *pb.ReattachConfig) {
//	// lock access to map
//	m.mut.Lock()
//	defer m.mut.Unlock()
//
//	// a RunningPlugin in initializing state will already have been put into the Plugins map
//	// populate its properties
//	p := m.connectionPluginMap[connection]
//	p.client = client
//	p.reattach = reattach
//
//	// store fully initialised runningPlugin to pluginMap
//	if reattach.SupportedOperations.MultipleConnections {
//		log.Printf("[INFO] store fully initialised runningPlugin to pluginMap %s (%p)", p.pluginName, p.client)
//		m.pluginMultiConnectionMap[reattach.Plugin] = p
//	}
//	// remove from loadingPlugins
//	delete(m.loadingPlugins, reattach.Plugin)
//	// NOTE: if this plugin supports multiple connections, reattach.Connections will be a list of all connections
//	// provided by this plugin
//	// add map entries for all other connections using this plugin (all pointing to same RunningPlugin)
//	for _, c := range reattach.Connections {
//		m.connectionPluginMap[c] = p
//	}
//	// mark as initialized
//	close(p.initialized)
//}

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

func (m *PluginManager) notifyNewDynamicSchemas(pluginClient *sdkgrpc.PluginClient, exemplarConnectionConfig *sdkproto.ConnectionConfig, connectionNames []string) error {
	// fetch the schema for the first connection so we know if it is dynamic
	schema, err := pluginClient.GetSchema(exemplarConnectionConfig.Connection)
	if err != nil {
		log.Printf("[WARN] failed to set fetch schema for %s: %s", exemplarConnectionConfig, err.Error())
		return err
	}
	if schema.Mode == sdkplugin.SchemaModeDynamic {
		_ = m.messageServer.AddConnection(pluginClient, exemplarConnectionConfig.Plugin, connectionNames...)
	}
	return nil
}

func (m *PluginManager) waitForPluginLoad(p *runningPlugin) error {
	log.Printf("[INFO] waitForPluginLoad")
	// TODO make this configurable
	pluginStartTimeoutSecs := 20

	// wait for the plugin to be initialized
	select {
	case <-time.After(time.Duration(pluginStartTimeoutSecs) * time.Second):
		log.Printf("[WARN] timed out waiting for %s to startup after %d seconds", p.pluginName, pluginStartTimeoutSecs)
		// do not retry
		return fmt.Errorf("timed out waiting for %s to startup after %d seconds", p.pluginName, pluginStartTimeoutSecs)
	case <-p.initialized:
		log.Printf("[TRACE] initialized: %d", p.reattach.Pid)
	case <-p.failed:
		log.Printf("[INFO] waitForPluginLoad")
		// get error from running plugin
		return p.error
		log.Printf("[WARN] initialized: %d", p.reattach.Pid)
	}

	// now double check the plugins process IS running
	exists, _ := utils.PidExists(int(p.reattach.Pid))
	if exists {
		// so the plugin is good
		log.Printf("[TRACE] waitForPluginLoad: %s is now loaded and ready", p.pluginName)
		return nil
	}

	// remove this plugin from the map
	m.mut.Lock()
	// check only needed for logging
	if _, ok := m.pluginMultiConnectionMap[p.pluginName]; ok {
		delete(m.pluginMultiConnectionMap, p.pluginName)
	}
	m.mut.Unlock()

	// so the pid does not exist
	// kill to be on the safe side
	//m.sigKillPlugin(p.reattach)

	err := fmt.Errorf("PluginManager found pid %d for plugin '%s' in plugin map but plugin process does not exist", p.reattach.Pid, p.pluginName)
	// we need to start the plugin again - make the error retryable
	return retry.RetryableError(err)
}

//func (m *PluginManager) getConnectionsForPlugin(pluginName string) []string {
//	var res = make([]string, len(m.pluginConnectionConfigMap[pluginName]))
//	for i, c := range m.pluginConnectionConfigMap[pluginName] {
//		res[i] = c.Connection
//	}
//	return res
//}

// set connection config for multiple connection
// NOTE: we DO NOT set connection config for aggregator connections
func (m *PluginManager) setAllConnectionConfigs(connectionConfigs []*sdkproto.ConnectionConfig, pluginClient *sdkgrpc.PluginClient, supportedOperations *sdkproto.GetSupportedOperationsResponse) error {
	exemplarConnectionConfig := connectionConfigs[0]
	pluginName := exemplarConnectionConfig.Plugin

	req := &sdkproto.SetAllConnectionConfigsRequest{
		Configs: connectionConfigs,
		// NOTE: set MaxCacheSizeMb to -1so that query cache is not created until we call SetCacheOptions (if supported)
		MaxCacheSizeMb: -1,
	}
	// if plugin _does not_ support setting the cache options separately, pass the max size now
	// (if it does support SetCacheOptions, it will be called after we return)
	if !supportedOperations.SetCacheOptions {
		req.MaxCacheSizeMb = m.pluginCacheSizeMap[pluginName]
	}

	_, err := pluginClient.SetAllConnectionConfigs(req)
	return err
}

func (m *PluginManager) setCacheOptions(pluginClient *sdkgrpc.PluginClient) error {
	req := &sdkproto.SetCacheOptionsRequest{
		Enabled:   viper.GetBool(constants.ArgServiceCacheEnabled),
		Ttl:       viper.GetInt64(constants.ArgCacheMaxTtl),
		MaxSizeMb: viper.GetInt64(constants.ArgMaxCacheSizeMb),
	}
	_, err := pluginClient.SetCacheOptions(req)
	return err
}

// update the schema for the specified connection
// called from the message server after receiving a PluginMessageType_SCHEMA_UPDATED message from plugin
func (m *PluginManager) updateConnectionSchema(ctx context.Context, connectionName string) {
	log.Printf("[TRACE] updateConnectionSchema connection %s", connectionName)

	refreshResult := connection.RefreshConnections(ctx, connectionName)
	if refreshResult.Error != nil {
		log.Printf("[TRACE] error refreshing connections: %s", refreshResult.Error)
		return
	}

	// also send a postgres notification
	notification := steampipeconfig.NewSchemaUpdateNotification(steampipeconfig.PgNotificationSchemaUpdate)

	conn, err := db_local.CreateLocalDbConnection(ctx, &db_local.CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		log.Printf("[WARN] failed to send schema update notification: %s", err)
	}

	err = db_local.SendPostgresNotification(ctx, conn, notification)
	if err != nil {
		log.Printf("[WARN] failed to send schema update notification: %s", err)
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
