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

	// map of running plugins keyed by plugin name
	runningPluginMap map[string]*runningPlugin
	// map of connection configs, keyed by plugin name
	// this is populated at startup and updated when a connection config change is detected
	pluginConnectionConfigMap map[string][]*sdkproto.ConnectionConfig
	// map of connection configs, keyed by connection name
	// this is populated at startup and updated when a connection config change is detected
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
	log.Printf("[INFO] NewPluginManager")
	pluginManager := &PluginManager{
		logger:              logger,
		runningPluginMap:    make(map[string]*runningPlugin),
		connectionConfigMap: connectionConfig,
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

	log.Printf("[TRACE] PluginManager Get %p", req)
	defer log.Printf("[TRACE] PluginManager Get DONE %p", req)

	resp := &pb.GetResponse{
		ReattachMap: make(map[string]*pb.ReattachConfig),
		FailureMap:  make(map[string]string),
	}
	// TODO validate we have config for this plugin

	// build a map of plugins required
	var plugins = make(map[string][]*sdkproto.ConnectionConfig)
	// also make a map of target connections - used when assigning resuts to the response
	var requestedConnectionsLookup = make(map[string]struct{}, len(req.Connections))
	for _, connectionName := range req.Connections {
		connectionConfig, err := m.getConnectionConfig(connectionName)
		if err != nil {
			return nil, err
		}
		pluginName := connectionConfig.Plugin
		// now get all connection configs for this plugin
		plugins[pluginName] = m.pluginConnectionConfigMap[pluginName]
		// store in requested connection map
		requestedConnectionsLookup[connectionName] = struct{}{}
	}

	log.Printf("[TRACE] PluginManager Get, connections: '%s'\n", req.Connections)
	for pluginName, connectionConfigs := range plugins {
		// ensure plugin is running
		reattach, err := m.ensurePlugin(pluginName, connectionConfigs, req)
		if err != nil {
			log.Printf("[WARN] PluginManager Get failed for %s: %s (%p)", pluginName, err.Error(), resp)
			resp.FailureMap[pluginName] = err.Error()
		} else {
			log.Printf("[TRACE] PluginManager Get succeeded for %s, pid %d (%p)", pluginName, reattach.Pid, resp)

			// assign reattach for requested connections
			// (NOTE: connectionConfigs contains ALL connections for the plugin)
			for _, config := range connectionConfigs {
				// if this connection was requested, copy reattach into responses
				if _, connectionWasRequested := requestedConnectionsLookup[config.Connection]; connectionWasRequested {
					resp.ReattachMap[config.Connection] = reattach
				}
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
	for _, p := range m.runningPluginMap {

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
		runningPlugin, pluginAlreadyRunning := m.runningPluginMap[plugin]

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
		runningPlugin, pluginAlreadyRunning := m.runningPluginMap[p]
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
		runningPlugin, pluginAlreadyRunning := m.runningPluginMap[p]
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

func (m *PluginManager) ensurePlugin(pluginName string, connectionConfigs []*sdkproto.ConnectionConfig, req *pb.GetRequest) (reattach *pb.ReattachConfig, err error) {
	/* call startPluginIfNeeded within a retry block
	 we wil retry if:
	 - we enter the plugin startup flow, but discover another process has beaten us to it an is starting the plugin already
	 - plugin initialization fails
	- there was a runningPlugin entry in our map bu tht epid did not exist
	*/

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

	log.Printf("[INFO] PluginManager ensurePlugin %s (%p)", pluginName, req)

	err = retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		reattach, err = m.startPluginIfNeeded(pluginName, connectionConfigs, req)
		return err
	})

	return
}

func (m *PluginManager) startPluginIfNeeded(pluginName string, connectionConfigs []*sdkproto.ConnectionConfig, req *pb.GetRequest) (*pb.ReattachConfig, error) {
	// is this plugin already running
	// lock access to plugin map
	log.Printf("[INFO] startPluginIfNeeded getting lock (%p)", req)
	m.mut.RLock()
	log.Printf("[INFO] startPluginIfNeeded got lock reading map, (len %d) (%p)", len(m.runningPluginMap), req)
	startingPlugin, ok := m.runningPluginMap[pluginName]
	m.mut.RUnlock()
	log.Printf("[INFO] startPluginIfNeeded released lock (%p)", req)

	if ok {
		log.Printf("[INFO] got running plugin (%p)", req)

		// wait for plugin to process connection config, and verify it is running
		err := m.waitForPluginLoad(startingPlugin, req)
		if err == nil {
			// so plugin has loaded - we are done
			log.Printf("[INFO] waitForPluginLoad succeeded %s (%p)", pluginName, req)
			return startingPlugin.reattach, nil
		}
		log.Printf("[INFO] waitForPluginLoad failed %s (%p)", err.Error(), req)

		// just return the error
		return nil, err
	}

	// so the plugin is NOT loaded or loading
	// fall through to plugin startup
	log.Printf("[WARN] plugin %s NOT started or starting - start now (%p)", pluginName, req)

	return m.startPlugin(pluginName, connectionConfigs, req)
}

func (m *PluginManager) startPlugin(pluginName string, connectionConfigs []*sdkproto.ConnectionConfig, req *pb.GetRequest) (_ *pb.ReattachConfig, err error) {
	log.Printf("[INFO] startPlugin %s (%p)", req)
	log.Printf("[WARN] running plugin len %d (%p)", len(m.runningPluginMap), req)

	// add a new running plugin to pluginMultiConnectionMap
	// (if someone beat us to it and added a starting plugin before we get the write lock,
	// this will return a retryable error)
	startingPlugin, err := m.addRunningPlugin(pluginName)
	if err != nil {
		log.Printf("[INFO] addRunningPlugin returned error %s (%p)", err.Error(), req)
		return nil, err
	}

	log.Printf("[INFO] added running plugin (%p)", req)

	// ensure we clean up the starting plugin in case of error
	defer func() {
		if err != nil {
			m.mut.Lock()

			// delete from map
			delete(m.runningPluginMap, pluginName)
			// TODO consider OnError callbacks/subscription instead
			// set error on running plugin
			startingPlugin.error = err
			// close failed chan to signal to anyone waiting for the plugin to startup that it failed
			close(startingPlugin.failed)

			log.Println("[WARN] startPluginProcess failed: %s (%p)", err, req)
			// kill the client
			if startingPlugin.client != nil {
				log.Println("[WARN] failed pid: %d (%p)", startingPlugin.client.ReattachConfig().Pid, req)
				startingPlugin.client.Kill()
			}
			m.mut.Unlock()
		}
	}()

	// OK so now proceed with plugin startup

	log.Printf("[INFO] start plugin (%p)", req)
	// now start the process
	client, err := m.startPluginProcess(pluginName, connectionConfigs)
	if err != nil {
		// do not retry - no reason to think this will fix itself
		return nil, err
	}

	// set the connection configs and build a ReattachConfig
	reattach, err := m.initializePlugin(connectionConfigs, client)
	if err != nil {
		// if initialization failed, we DO retry - just in case it was a temporary error (network etc)
		return nil, retry.RetryableError(err)
	}

	log.Printf("[INFO] assign reattach and client")
	startingPlugin.reattach = reattach
	startingPlugin.client = client
	log.Printf("[INFO] assign reattached and client")

	log.Printf("[INFO] close init chan")
	// close initialized chan to advertise that this plugin is ready
	close(startingPlugin.initialized)
	log.Printf("[INFO] closed init chan")

	log.Printf("[INFO] PluginManager ensurePlugin complete, returning reattach config with PID: %d (%p)", reattach.Pid, req)

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

	// check someone else has beaten us to it (there is a race condition to starting a plugin)
	if _, ok := m.runningPluginMap[pluginName]; ok {
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
	m.runningPluginMap[pluginName] = startingPlugin
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
	t := time.Now()
	log.Printf("[WARN] sending SIGKILL to plugin %s (pid %d)", reattach.Plugin, reattach.Pid)
	// kill to be certain
	err := syscall.Kill(int(reattach.Pid), syscall.SIGKILL)
	if err != nil {
		log.Printf("[WARN] failed to SIGKILL process %d: %s (%dms)", reattach.Pid, err.Error(), time.Since(t).Milliseconds())
	} else {
		log.Printf("[WARN] SIGKILL process %d succeeded: (%dms)", reattach.Pid, time.Since(t).Milliseconds())
	}
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

func (m *PluginManager) waitForPluginLoad(p *runningPlugin, req *pb.GetRequest) error {
	log.Printf("[INFO] waitForPluginLoad (%p)", req)
	// TODO make this configurable
	pluginStartTimeoutSecs := 30

	// wait for the plugin to be initialized
	select {
	case <-time.After(time.Duration(pluginStartTimeoutSecs) * time.Second):
		log.Printf("[WARN] timed out waiting for %s to startup after %d seconds (%p)", p.pluginName, pluginStartTimeoutSecs, req)
		// do not retry
		return fmt.Errorf("timed out waiting for %s to startup after %d seconds (%p)", p.pluginName, pluginStartTimeoutSecs, req)
	case <-p.initialized:
		log.Printf("[TRACE] initialized: pid %d (%p)", p.reattach.Pid, req)
	case <-p.failed:
		log.Printf("[INFO] plugin pid %d failed %s (%p)", p.reattach.Pid, p.error.Error(), req)
		// get error from running plugin
		return p.error
		log.Printf("[WARN] initialized: %d (%p)", p.reattach.Pid, req)
	}

	// now double-check the plugins process IS running
	exists, err := utils.PidExists(int(p.reattach.Pid))
	if err != nil {
		log.Printf("[WARN] PidExists failed for pid %d when checking status of '%s' plugin: %s (%p)", p.reattach.Pid, p.pluginName, err.Error(), req)
	}

	if exists {
		// so the plugin is good
		log.Printf("[INFO] waitForPluginLoad: %s is now loaded and ready (%p)", p.pluginName, req)
		return nil
	}

	// so even though our data structure indicates the plugin is running, the pid is missing
	// - it must have terminated for some reason

	log.Printf("[INFO] waitForPluginLoad: pid %d exists in runningPluginMap but pid does not exist (%p)", p.reattach.Pid, req)

	// remove this plugin from the map
	// NOTE: multiple thread may be trying to remove the failed plugin from the map
	// - and then someone will add a new running plugin when the starup is retried
	// So we must check the pid before deleting
	m.mut.Lock()
	if r, ok := m.runningPluginMap[p.pluginName]; ok {
		if r == p {
			log.Printf("[INFO] delete plugin %s from runningPluginMap (%p)", p.pluginName, req)
			delete(m.runningPluginMap, p.pluginName)
		} else {
			log.Printf("[INFO] running plugin in map does not match our running plugin - must have been deleted and replaced  (%p)", req)
		}
	} else {
		log.Printf("[INFO] no plugin in running plugin map for %s (%p)", p.pluginName, req)
	}

	m.mut.Unlock()

	// so the pid does not exist
	// kill to be on the safe side
	m.sigKillPlugin(p.reattach)

	err = fmt.Errorf("PluginManager found pid %d for plugin '%s' in plugin map but plugin process does not exist (%p)", p.reattach.Pid, p.pluginName, req)
	// we need to start the plugin again - make the error retryable
	return retry.RetryableError(err)
}

// set connection config for multiple connection
// NOTE: we DO NOT set connection config for aggregator connections
func (m *PluginManager) setAllConnectionConfigs(connectionConfigs []*sdkproto.ConnectionConfig, pluginClient *sdkgrpc.PluginClient, supportedOperations *sdkproto.GetSupportedOperationsResponse) error {
	// TODO does this fail all connections if one fails
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
