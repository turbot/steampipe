package pluginmanager_service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sethvargo/go-retry"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/utils"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	sdkshared "github.com/turbot/steampipe-plugin-sdk/v5/grpc/shared"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/connection"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
	pluginshared "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/shared"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// PluginManager is the implementation of grpc.PluginManager
type PluginManager struct {
	pb.UnimplementedPluginManagerServer

	// map of running plugins keyed by plugin instance
	runningPluginMap map[string]*runningPlugin
	// map of connection configs, keyed by plugin instance
	// this is populated at startup and updated when a connection config change is detected
	pluginConnectionConfigMap map[string][]*sdkproto.ConnectionConfig
	// map of connection configs, keyed by connection name
	// this is populated at startup and updated when a connection config change is detected
	connectionConfigMap connection.ConnectionConfigMap
	// map of max cache size, keyed by plugin instance
	pluginCacheSizeMap map[string]int64

	// mut protects concurrent access to plugin manager state (runningPluginMap, connectionConfigMap, etc.)
	//
	// LOCKING PATTERN TO PREVENT DEADLOCKS:
	// - Functions that acquire mut.Lock() and call other methods MUST only call *Internal versions
	// - Public methods that need locking: acquire lock → call internal version → release lock
	// - Internal methods: assume caller holds lock, never acquire lock themselves
	//
	// Example:
	//   func (m *PluginManager) SomeMethod() {
	//       m.mut.Lock()
	//       defer m.mut.Unlock()
	//       return m.someMethodInternal()
	//   }
	//   func (m *PluginManager) someMethodInternal() {
	//       // NOTE: caller must hold m.mut lock
	//       // ... implementation without locking ...
	//   }
	//
	// Functions with internal/external versions:
	// - refreshRateLimiterTable / refreshRateLimiterTableInternal
	// - updateRateLimiterStatus / updateRateLimiterStatusInternal
	// - setRateLimiters / setRateLimitersInternal
	// - getPluginsWithChangedLimiters / getPluginsWithChangedLimitersInternal
	mut sync.RWMutex

	// shutdown synchronization
	// do not start any plugins while shutting down
	shutdownMut  sync.RWMutex
	shuttingDown bool
	// do not shutdown until all plugins have loaded
	startPluginWg sync.WaitGroup

	logger        hclog.Logger
	messageServer *PluginMessageServer

	// map of user configured rate limiter maps, keyed by plugin instance
	// NOTE: this is populated from config
	userLimiters connection.PluginLimiterMap
	// map of plugin configured rate limiter maps  (keyed by plugin instance)
	// NOTE: if this is nil, that means the steampipe_rate_limiter tables has not been populated yet -
	// the first time we refresh connections we must load all plugins and fetch their rate limiter defs
	pluginLimiters connection.PluginLimiterMap

	// map of plugin configs (keyed by plugin instance)
	plugins connection.PluginMap

	pool *pgxpool.Pool
}

func NewPluginManager(ctx context.Context, connectionConfig map[string]*sdkproto.ConnectionConfig, pluginConfigs connection.PluginMap, logger hclog.Logger) (*PluginManager, error) {
	log.Printf("[INFO] NewPluginManager")
	pluginManager := &PluginManager{
		logger:              logger,
		runningPluginMap:    make(map[string]*runningPlugin),
		connectionConfigMap: connectionConfig,
		userLimiters:        pluginConfigs.ToPluginLimiterMap(),
		plugins:             pluginConfigs,
	}

	pluginManager.messageServer = &PluginMessageServer{pluginManager: pluginManager}

	// populate plugin connection config map
	pluginManager.populatePluginConnectionConfigs()
	// determine cache size for each plugin
	pluginManager.setPluginCacheSizeMap()

	// create a connection pool to connection refresh
	// in testing, a size of 20 seemed optimal
	poolsize := 20
	pool, err := db_local.CreateConnectionPool(ctx, &db_local.CreateDbOptions{Username: constants.DatabaseSuperUser}, poolsize)
	if err != nil {
		return nil, err
	}
	pluginManager.pool = pool

	if err := pluginManager.initialiseRateLimiterDefs(ctx); err != nil {
		return nil, err
	}

	if err := pluginManager.initialisePluginColumns(ctx); err != nil {
		return nil, err
	}
	return pluginManager, nil
}

// plugin interface functions

func (m *PluginManager) Serve() {
	// create a plugin map, using ourselves as the implementation
	pluginMap := map[string]goplugin.Plugin{
		pluginshared.PluginName: &pluginshared.PluginManagerPlugin{Impl: m},
	}
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginMap,
		//  enable gRPC serving for this plugin...
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}

func (m *PluginManager) Get(req *pb.GetRequest) (_ *pb.GetResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = sperr.ToError(r, sperr.WithMessage("unexpected error encountered"))
		}
	}()
	log.Printf("[TRACE] PluginManager Get %p", req)
	defer log.Printf("[TRACE] PluginManager Get DONE %p", req)

	resp := newGetResponse()

	// build a map of plugins to connection config for requested connections, and a lookup of the requested connections
	plugins, requestedConnectionsLookup, err := m.buildRequiredPluginMap(req)
	if err != nil {
		return resp.GetResponse, err
	}

	log.Printf("[TRACE] PluginManager Get, connections: '%s'\n", req.Connections)
	var pluginWg sync.WaitGroup
	for pluginInstance, connectionConfigs := range plugins {
		m.ensurePluginAsync(req, resp, pluginInstance, connectionConfigs, requestedConnectionsLookup, &pluginWg)
	}
	pluginWg.Wait()

	log.Printf("[TRACE] PluginManager Get DONE")
	return resp.GetResponse, nil
}

func (m *PluginManager) ensurePluginAsync(req *pb.GetRequest, resp *getResponse, pluginInstance string, connectionConfigs []*sdkproto.ConnectionConfig, requestedConnectionsLookup map[string]struct{}, pluginWg *sync.WaitGroup) {
	pluginWg.Add(1)
	go func() {
		defer pluginWg.Done()
		// ensure plugin is running
		reattach, err := m.ensurePlugin(pluginInstance, connectionConfigs, req)
		if err != nil {
			log.Printf("[WARN] PluginManager Get failed for %s: %s (%p)", pluginInstance, err.Error(), resp)
			resp.AddFailure(pluginInstance, err.Error())
		} else {
			log.Printf("[TRACE] PluginManager Get succeeded for %s, pid %d (%p)", pluginInstance, reattach.Pid, resp)

			// assign reattach for requested connections
			// (NOTE: connectionConfigs contains ALL connections for the plugin)
			for _, config := range connectionConfigs {
				// if this connection was requested, copy reattach into responses
				if _, connectionWasRequested := requestedConnectionsLookup[config.Connection]; connectionWasRequested {
					resp.AddReattach(config.Connection, reattach)
				}
			}
		}
	}()
}

// build a map of plugins to connection config for requested connections, keyed by plugin instance,
// and a lookup of the requested connections
func (m *PluginManager) buildRequiredPluginMap(req *pb.GetRequest) (map[string][]*sdkproto.ConnectionConfig, map[string]struct{}, error) {
	var plugins = make(map[string][]*sdkproto.ConnectionConfig)
	// also make a map of target connections - used when assigning results to the response
	var requestedConnectionsLookup = make(map[string]struct{}, len(req.Connections))
	for _, connectionName := range req.Connections {
		// store connection in requested connection map
		requestedConnectionsLookup[connectionName] = struct{}{}

		connectionConfig, err := m.getConnectionConfig(connectionName)
		if err != nil {
			return nil, nil, err
		}
		pluginInstance := connectionConfig.PluginInstance
		// if we have not added this plugin instance, add it now
		if _, addedPlugin := plugins[pluginInstance]; !addedPlugin {
			// now get ALL connection configs for this plugin
			// (not just the requested connections)
			plugins[pluginInstance] = m.pluginConnectionConfigMap[pluginInstance]
		}
	}
	return plugins, requestedConnectionsLookup, nil
}

func (m *PluginManager) Pool() *pgxpool.Pool {
	return m.pool
}

func (m *PluginManager) RefreshConnections(*pb.RefreshConnectionsRequest) (*pb.RefreshConnectionsResponse, error) {
	log.Printf("[INFO] PluginManager RefreshConnections")

	resp := &pb.RefreshConnectionsResponse{}

	log.Printf("[INFO] calling RefreshConnections asyncronously")

	go m.doRefresh()
	return resp, nil
}

func (m *PluginManager) doRefresh() {
	refreshResult := connection.RefreshConnections(context.Background(), m)
	if refreshResult.Error != nil {
		// NOTE: the RefreshConnectionState will already have sent a notification to the CLI
		log.Printf("[WARN] RefreshConnections failed with error: %s", refreshResult.Error.Error())
	}
}

// OnConnectionConfigChanged is the callback function invoked by the connection watcher when the config changed
func (m *PluginManager) OnConnectionConfigChanged(ctx context.Context, configMap connection.ConnectionConfigMap, plugins map[string]*plugin.Plugin) {
	log.Printf("[DEBUG] OnConnectionConfigChanged: acquiring lock")
	m.mut.Lock()
	defer m.mut.Unlock()
	log.Printf("[DEBUG] OnConnectionConfigChanged: lock acquired")

	log.Printf("[TRACE] OnConnectionConfigChanged: connections: %s plugin instances: %s", strings.Join(utils.SortedMapKeys(configMap), ","), strings.Join(utils.SortedMapKeys(plugins), ","))

	log.Printf("[DEBUG] OnConnectionConfigChanged: calling handleConnectionConfigChanges")
	if err := m.handleConnectionConfigChanges(ctx, configMap); err != nil {
		log.Printf("[WARN] handleConnectionConfigChanges failed: %s", err.Error())
	}
	log.Printf("[DEBUG] OnConnectionConfigChanged: handleConnectionConfigChanges complete")

	// update our plugin configs
	log.Printf("[DEBUG] OnConnectionConfigChanged: calling handlePluginInstanceChanges")
	if err := m.handlePluginInstanceChanges(ctx, plugins); err != nil {
		log.Printf("[WARN] handlePluginInstanceChanges failed: %s", err.Error())
	}
	log.Printf("[DEBUG] OnConnectionConfigChanged: handlePluginInstanceChanges complete")

	log.Printf("[DEBUG] OnConnectionConfigChanged: calling handleUserLimiterChanges")
	if err := m.handleUserLimiterChanges(ctx, plugins); err != nil {
		log.Printf("[WARN] handleUserLimiterChanges failed: %s", err.Error())
	}
	log.Printf("[DEBUG] OnConnectionConfigChanged: handleUserLimiterChanges complete")
	log.Printf("[DEBUG] OnConnectionConfigChanged: about to release lock and return")
}

func (m *PluginManager) GetConnectionConfig() connection.ConnectionConfigMap {
	return m.connectionConfigMap
}

func (m *PluginManager) Shutdown(*pb.ShutdownRequest) (resp *pb.ShutdownResponse, err error) {
	log.Printf("[INFO] PluginManager Shutdown")
	defer log.Printf("[INFO] PluginManager Shutdown complete")

	// lock shutdownMut before waiting for startPluginWg
	// this enables us to exit from ensurePlugin early if needed
	m.shutdownMut.Lock()
	m.shuttingDown = true
	m.shutdownMut.Unlock()
	m.startPluginWg.Wait()

	// close our pool
	if m.pool != nil {
		log.Printf("[INFO] PluginManager closing pool")
		m.pool.Close()
	}

	m.mut.RLock()
	defer func() {
		m.mut.RUnlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	// kill all plugins in pluginMultiConnectionMap
	for _, p := range m.runningPluginMap {
		log.Printf("[INFO] Kill plugin %s (%p)", p.pluginInstance, p.client)
		m.killPlugin(p)
	}

	return &pb.ShutdownResponse{}, nil
}

func (m *PluginManager) killPlugin(p *runningPlugin) {
	log.Println("[DEBUG] PluginManager killPlugin start")
	defer log.Println("[DEBUG] PluginManager killPlugin complete")

	if p.client == nil {
		log.Printf("[WARN] plugin %s has no client - cannot kill client", p.pluginInstance)
		// shouldn't happen but has been observed in error situations
		return
	}
	log.Printf("[INFO] PluginManager killing plugin %s (%v)", p.pluginInstance, p.reattach.Pid)
	p.client.Kill()
}

func (m *PluginManager) ensurePlugin(pluginInstance string, connectionConfigs []*sdkproto.ConnectionConfig, req *pb.GetRequest) (reattach *pb.ReattachConfig, err error) {
	/* call startPluginIfNeeded within a retry block
	 we will retry if:
	 - we enter the plugin startup flow, but discover another process has beaten us to it an is starting the plugin already
	 - plugin initialization fails
	- there was a runningPlugin entry in our map but the pid did not exist
	  (i.e we thought the plugin was running, but it was not)
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
	if m.isShuttingDown() {
		return nil, fmt.Errorf("plugin manager is shutting down")
	}

	log.Printf("[TRACE] PluginManager ensurePlugin %s (%p)", pluginInstance, req)

	err = retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		reattach, err = m.startPluginIfNeeded(pluginInstance, connectionConfigs, req)
		return err
	})

	return
}

func (m *PluginManager) startPluginIfNeeded(pluginInstance string, connectionConfigs []*sdkproto.ConnectionConfig, req *pb.GetRequest) (*pb.ReattachConfig, error) {
	// is this plugin already running
	// lock access to plugin map
	m.mut.RLock()
	startingPlugin, ok := m.runningPluginMap[pluginInstance]
	m.mut.RUnlock()

	if ok {
		log.Printf("[TRACE] startPluginIfNeeded got running plugin (%p)", req)

		// wait for plugin to process connection config, and verify it is running
		err := m.waitForPluginLoad(startingPlugin, req)
		if err == nil {
			// so plugin has loaded - we are done

			// NOTE: ensure the connections assigned to this plugin are correct
			// (may be out of sync if a connection is being added)
			m.mut.Lock()
			startingPlugin.reattach.UpdateConnections(connectionConfigs)
			m.mut.Unlock()

			log.Printf("[TRACE] waitForPluginLoad succeeded %s (%p)", pluginInstance, req)
			return startingPlugin.reattach, nil
		}
		log.Printf("[TRACE] waitForPluginLoad failed %s (%p)", err.Error(), req)

		// just return the error
		return nil, err
	}

	// so the plugin is NOT loaded or loading
	// fall through to plugin startup
	log.Printf("[INFO] plugin %s NOT started or starting - start now (%p)", pluginInstance, req)

	return m.startPlugin(pluginInstance, connectionConfigs, req)
}

func (m *PluginManager) startPlugin(pluginInstance string, connectionConfigs []*sdkproto.ConnectionConfig, req *pb.GetRequest) (_ *pb.ReattachConfig, err error) {
	log.Printf("[DEBUG] startPlugin %s (%p) start", pluginInstance, req)
	defer log.Printf("[DEBUG] startPlugin %s (%p) end", pluginInstance, req)

	// add a new running plugin to pluginMultiConnectionMap
	// (if someone beat us to it and added a starting plugin before we get the write lock,
	// this will return a retryable error)
	startingPlugin, err := m.addRunningPlugin(pluginInstance)
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
			delete(m.runningPluginMap, pluginInstance)
			// set error on running plugin
			startingPlugin.error = err

			// close failed chan to signal to anyone waiting for the plugin to startup that it failed
			close(startingPlugin.failed)

			log.Printf("[INFO] startPluginProcess failed: %s (%p)", err.Error(), req)
			// kill the client
			if startingPlugin.client != nil {
				log.Printf("[INFO] failed pid: %d (%p)", startingPlugin.client.ReattachConfig().Pid, req)
				startingPlugin.client.Kill()
			}

			m.mut.Unlock()
		}
	}()

	// OK so now proceed with plugin startup

	log.Printf("[INFO] start plugin (%p)", req)
	// now start the process
	client, err := m.startPluginProcess(pluginInstance, connectionConfigs)
	if err != nil {
		// do not retry - no reason to think this will fix itself
		return nil, err
	}

	startingPlugin.client = client

	// set the connection configs and build a ReattachConfig
	reattach, err := m.initializePlugin(connectionConfigs, client, req)
	if err != nil {
		log.Printf("[WARN] initializePlugin failed: %s (%p)", err.Error(), req)
		return nil, err
	}
	startingPlugin.reattach = reattach

	// close initialized chan to advertise that this plugin is ready
	close(startingPlugin.initialized)

	log.Printf("[INFO] PluginManager ensurePlugin complete, returning reattach config with PID: %d (%p)", reattach.Pid, req)

	// and return
	return reattach, nil
}

func (m *PluginManager) addRunningPlugin(pluginInstance string) (*runningPlugin, error) {
	// add a new running plugin to pluginMultiConnectionMap
	// this is a placeholder so no other thread tries to create start this plugin

	// acquire write lock
	m.mut.Lock()
	defer m.mut.Unlock()
	log.Printf("[TRACE] add running plugin for %s (if someone didn't beat us to it)", pluginInstance)

	// check someone else has beaten us to it (there is a race condition to starting a plugin)
	if _, ok := m.runningPluginMap[pluginInstance]; ok {
		log.Printf("[TRACE] re checked map and found a starting plugin - return retryable error so we wait for this plugin")
		// if so, just retry, which will wait for the loading plugin
		return nil, retry.RetryableError(fmt.Errorf("another client has already started the plugin"))
	}

	// get the config for this instance
	pluginConfig := m.plugins[pluginInstance]
	if pluginConfig == nil {
		// not expected
		return nil, sperr.New("plugin manager has no config for plugin instance %s", pluginInstance)
	}
	// create the running plugin
	startingPlugin := &runningPlugin{
		pluginInstance: pluginInstance,
		imageRef:       pluginConfig.Plugin,
		initialized:    make(chan struct{}),
		failed:         make(chan struct{}),
	}
	// write back
	m.runningPluginMap[pluginInstance] = startingPlugin

	log.Printf("[INFO] written running plugin to map")

	return startingPlugin, nil
}

func (m *PluginManager) startPluginProcess(pluginInstance string, connectionConfigs []*sdkproto.ConnectionConfig) (*goplugin.Client, error) {
	// retrieve the plugin config
	pluginConfig := m.plugins[pluginInstance]
	// must be there (if no explicit config was specified, we create a default)
	if pluginConfig == nil {
		panic(fmt.Sprintf("no plugin config is stored for plugin instance %s", pluginInstance))
	}

	imageRef := pluginConfig.Plugin
	log.Printf("[INFO] ************ start plugin: %s, label: %s ********************\n", imageRef, pluginConfig.Instance)

	// NOTE: pass pluginConfig.Alias as the pluginAlias
	// - this is just used for the error message if we fail to load
	pluginPath, err := filepaths.GetPluginPath(imageRef, pluginConfig.Alias)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] ************ plugin path %s ********************\n", pluginPath)

	// create the plugin map
	pluginMap := map[string]goplugin.Plugin{
		imageRef: &sdkshared.WrapperPlugin{},
	}

	cmd := exec.Command(pluginPath)
	m.setPluginMaxMemory(pluginConfig, cmd)

	pluginStartTimeoutDuration := time.Duration(viper.GetInt64(pconstants.ArgPluginStartTimeout)) * time.Second
	log.Printf("[TRACE] %s pluginStartTimeoutDuration: %s", pluginPath, pluginStartTimeoutDuration)

	client := goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig:  sdkshared.Handshake,
		Plugins:          pluginMap,
		Cmd:              cmd,
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
		StartTimeout:     pluginStartTimeoutDuration,

		// pass our logger to the plugin client to ensure plugin logs end up in logfile
		Logger: m.logger,
	})

	if _, err := client.Start(); err != nil {
		// attempt to retrieve error message encoded in the plugin stdout
		err := grpc.HandleStartFailure(err)
		return nil, err
	}

	return client, nil

}

func (m *PluginManager) setPluginMaxMemory(pluginConfig *plugin.Plugin, cmd *exec.Cmd) {
	maxMemoryBytes := pluginConfig.GetMaxMemoryBytes()
	if maxMemoryBytes == 0 {
		if viper.IsSet(pconstants.ArgMemoryMaxMbPlugin) {
			maxMemoryBytes = viper.GetInt64(pconstants.ArgMemoryMaxMbPlugin) * 1024 * 1024
		}
	}
	if maxMemoryBytes != 0 {
		log.Printf("[INFO] Setting max memory for plugin '%s' to %d Mb", pluginConfig.Instance, maxMemoryBytes/(1024*1024))
		// set GOMEMLIMIT for the plugin command env
		// TODO should I check for GOMEMLIMIT or does this just override
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOMEMLIMIT=%d", maxMemoryBytes))
	}
}

// set the connection configs and build a ReattachConfig
func (m *PluginManager) initializePlugin(connectionConfigs []*sdkproto.ConnectionConfig, client *goplugin.Client, req *pb.GetRequest) (_ *pb.ReattachConfig, err error) {
	// extract connection names
	connectionNames := make([]string, len(connectionConfigs))
	for i, c := range connectionConfigs {
		connectionNames[i] = c.Connection
	}
	exemplarConnectionConfig := connectionConfigs[0]
	pluginName := exemplarConnectionConfig.Plugin
	pluginInstance := exemplarConnectionConfig.PluginInstance

	log.Printf("[INFO] initializePlugin %s pid %d (%p)", pluginName, client.ReattachConfig().Pid, req)

	// build a client
	pluginClient, err := sdkgrpc.NewPluginClient(client, pluginName)
	if err != nil {
		return nil, err
	}

	// fetch the supported operations
	supportedOperations, _ := pluginClient.GetSupportedOperations()
	// ignore errors  - just create an empty support structure if needed
	if supportedOperations == nil {
		supportedOperations = &sdkproto.GetSupportedOperationsResponse{}
	}
	// if this plugin does not support multiple connections, we no longer support it
	if !supportedOperations.MultipleConnections {
		return nil, fmt.Errorf("%s", error_helpers.PluginSdkCompatibilityError)
	}

	// provide opportunity to avoid setting connection configs if we are shutting down
	if m.isShuttingDown() {
		log.Printf("[INFO] aborting plugin %s initialization - plugin manager is shutting down", pluginName)
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

	// if this plugin supports setting cache options, do so
	if supportedOperations.RateLimiters {
		err = m.setRateLimiters(pluginInstance, pluginClient)
		if err != nil {
			log.Printf("[WARN] failed to set rate limiters for %s: %s", pluginName, err.Error())
			return nil, err
		}
	}

	reattach := pb.NewReattachConfig(pluginName, client.ReattachConfig(), pb.SupportedOperationsFromSdk(supportedOperations), connectionNames)

	// if this plugin has a dynamic schema, add connections to message server
	err = m.notifyNewDynamicSchemas(pluginClient, exemplarConnectionConfig, connectionNames)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] initializePlugin complete pid %d", client.ReattachConfig().Pid)
	return reattach, nil
}

// return whether the plugin manager is shutting down
func (m *PluginManager) isShuttingDown() bool {
	m.shutdownMut.RLock()
	defer m.shutdownMut.RUnlock()
	return m.shuttingDown
}

// populate map of connection configs for each plugin instance
func (m *PluginManager) populatePluginConnectionConfigs() {
	m.pluginConnectionConfigMap = make(map[string][]*sdkproto.ConnectionConfig)
	for _, config := range m.connectionConfigMap {
		m.pluginConnectionConfigMap[config.PluginInstance] = append(m.pluginConnectionConfigMap[config.PluginInstance], config)
	}
}

// populate map of connection configs for each plugin
func (m *PluginManager) setPluginCacheSizeMap() {
	m.pluginCacheSizeMap = make(map[string]int64, len(m.pluginConnectionConfigMap))

	// read the env var setting cache size
	maxCacheSizeMb, _ := strconv.Atoi(os.Getenv(constants.EnvCacheMaxSize))

	// get total connection count for this pluginInstance (excluding aggregators)
	numConnections := m.nonAggregatorConnectionCount()

	log.Printf("[TRACE] PluginManager setPluginCacheSizeMap: %d %s.", numConnections, utils.Pluralize("connection", numConnections))
	log.Printf("[TRACE] Total cache size %dMb", maxCacheSizeMb)

	for pluginInstance, connections := range m.pluginConnectionConfigMap {
		var size int64 = 0
		// if no max size is set, just set all plugins to zero (unlimited)
		if maxCacheSizeMb > 0 {
			// get connection count for this pluginInstance (excluding aggregators)
			numPluginConnections := nonAggregatorConnectionCount(connections)
			size = int64(float64(numPluginConnections) / float64(numConnections) * float64(maxCacheSizeMb))
			// make this at least 1 Mb (as zero means unlimited)
			if size == 0 {
				size = 1
			}
			log.Printf("[INFO] Plugin '%s', %d %s, max cache size %dMb", pluginInstance, numPluginConnections, utils.Pluralize("connection", numPluginConnections), size)
		}

		m.pluginCacheSizeMap[pluginInstance] = size
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

	pluginConfig := m.plugins[p.pluginInstance]
	if pluginConfig == nil {
		// not expected
		return sperr.New("plugin manager has no config for plugin instance %s", p.pluginInstance)
	}
	pluginStartTimeoutSecs := pluginConfig.GetStartTimeout()
	if pluginStartTimeoutSecs == 0 {
		if viper.IsSet(pconstants.ArgPluginStartTimeout) {
			pluginStartTimeoutSecs = viper.GetInt64(pconstants.ArgPluginStartTimeout)
		}
	}

	log.Printf("[TRACE] waitForPluginLoad: waiting %d seconds (%p)", pluginStartTimeoutSecs, req)

	// wait for the plugin to be initialized
	select {
	case <-time.After(time.Duration(pluginStartTimeoutSecs) * time.Second):
		log.Printf("[WARN] timed out waiting for %s to startup after %d seconds (%p)", p.pluginInstance, pluginStartTimeoutSecs, req)
		// do not retry
		return fmt.Errorf("timed out waiting for %s to startup after %d seconds (%p)", p.pluginInstance, pluginStartTimeoutSecs, req)
	case <-p.initialized:
		log.Printf("[TRACE] plugin initialized: pid %d (%p)", p.reattach.Pid, req)
	case <-p.failed:
		// reattach may be nil if plugin failed before it was set
		if p.reattach != nil {
			log.Printf("[TRACE] plugin pid %d failed %s (%p)", p.reattach.Pid, p.error.Error(), req)
		} else {
			log.Printf("[TRACE] plugin %s failed before reattach was set: %s (%p)", p.pluginInstance, p.error.Error(), req)
		}
		// get error from running plugin
		return p.error
	}

	// now double-check the plugins process IS running
	if !p.client.Exited() {
		// so the plugin is good
		log.Printf("[INFO] waitForPluginLoad: %s is now loaded and ready (%p)", p.pluginInstance, req)
		return nil
	}

	// so even though our data structure indicates the plugin is running, the client says the underlying pid has exited
	// - it must have terminated for some reason
	log.Printf("[INFO] waitForPluginLoad: pid %d exists in runningPluginMap but pid has exited (%p)", p.reattach.Pid, req)

	// remove this plugin from the map
	// NOTE: multiple thread may be trying to remove the failed plugin from the map
	// - and then someone will add a new running plugin when the startup is retried
	// So we must check the pid before deleting
	m.mut.Lock()
	if r, ok := m.runningPluginMap[p.pluginInstance]; ok {
		// is the running plugin we read from the map the same as our running plugin?
		// if not, it must already have been removed by another thread - do nothing
		if r == p {
			log.Printf("[INFO] delete plugin %s from runningPluginMap (%p)", p.pluginInstance, req)
			delete(m.runningPluginMap, p.pluginInstance)
		}
	}
	m.mut.Unlock()

	// so the pid does not exist
	err := fmt.Errorf("PluginManager found pid %d for plugin '%s' in plugin map but plugin process does not exist (%p)", p.reattach.Pid, p.pluginInstance, req)
	// we need to start the plugin again - make the error retryable
	return retry.RetryableError(err)
}

// set connection config for multiple connection
// NOTE: we DO NOT set connection config for aggregator connections
func (m *PluginManager) setAllConnectionConfigs(connectionConfigs []*sdkproto.ConnectionConfig, pluginClient *sdkgrpc.PluginClient, supportedOperations *sdkproto.GetSupportedOperationsResponse) error {
	// TODO does this fail all connections if one fails
	exemplarConnectionConfig := connectionConfigs[0]
	pluginInstance := exemplarConnectionConfig.PluginInstance

	req := &sdkproto.SetAllConnectionConfigsRequest{
		Configs: connectionConfigs,
		// NOTE: set MaxCacheSizeMb to -1so that query cache is not created until we call SetCacheOptions (if supported)
		MaxCacheSizeMb: -1,
	}
	// if plugin _does not_ support setting the cache options separately, pass the max size now
	// (if it does support SetCacheOptions, it will be called after we return)
	if !supportedOperations.SetCacheOptions {
		req.MaxCacheSizeMb = m.pluginCacheSizeMap[pluginInstance]
	}

	_, err := pluginClient.SetAllConnectionConfigs(req)
	return err
}

func (m *PluginManager) setCacheOptions(pluginClient *sdkgrpc.PluginClient) error {
	req := &sdkproto.SetCacheOptionsRequest{
		Enabled:   viper.GetBool(pconstants.ArgServiceCacheEnabled),
		Ttl:       viper.GetInt64(pconstants.ArgCacheMaxTtl),
		MaxSizeMb: viper.GetInt64(pconstants.ArgMaxCacheSizeMb),
	}
	_, err := pluginClient.SetCacheOptions(req)
	return err
}

func (m *PluginManager) setRateLimiters(pluginInstance string, pluginClient *sdkgrpc.PluginClient) error {
	m.mut.RLock()
	defer m.mut.RUnlock()
	return m.setRateLimitersInternal(pluginInstance, pluginClient)
}

func (m *PluginManager) setRateLimitersInternal(pluginInstance string, pluginClient *sdkgrpc.PluginClient) error {
	// NOTE: caller must hold m.mut lock (at least RLock)
	log.Printf("[INFO] setRateLimiters for plugin '%s'", pluginInstance)
	var defs []*sdkproto.RateLimiterDefinition

	for _, l := range m.userLimiters[pluginInstance] {
		defs = append(defs, RateLimiterAsProto(l))
	}

	req := &sdkproto.SetRateLimitersRequest{Definitions: defs}

	_, err := pluginClient.SetRateLimiters(req)
	return err
}

// update the schema for the specified connection
// called from the message server after receiving a PluginMessageType_SCHEMA_UPDATED message from plugin
func (m *PluginManager) updateConnectionSchema(ctx context.Context, connectionName string) {
	log.Printf("[INFO] updateConnectionSchema connection %s", connectionName)

	// check if pool is nil before attempting to refresh connections
	if m.pool == nil {
		log.Printf("[WARN] cannot update connection schema: pool is nil")
		return
	}

	refreshResult := connection.RefreshConnections(ctx, m, connectionName)
	if refreshResult.Error != nil {
		log.Printf("[TRACE] error refreshing connections: %s", refreshResult.Error)
		return
	}

	// also send a postgres notification
	notification := steampipeconfig.NewSchemaUpdateNotification()

	if m.pool == nil {
		log.Printf("[WARN] cannot send schema update notification: pool is nil")
		return
	}
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		log.Printf("[WARN] failed to send schema update notification: %s", err)
		return
	}
	defer conn.Release()

	err = db_local.SendPostgresNotification(ctx, conn.Conn(), notification)
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

// getPluginExemplarConnections returns a map of keyed by plugin full name with the value an exemplar connection
func (m *PluginManager) getPluginExemplarConnections() map[string]string {
	res := make(map[string]string)
	for _, c := range m.connectionConfigMap {
		res[c.Plugin] = c.Connection
	}
	return res
}

func (m *PluginManager) tableExists(ctx context.Context, schema, table string) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS (
    SELECT FROM 
        pg_tables
    WHERE 
        schemaname = '%s' AND 
        tablename  = '%s'
    );`, schema, table)

	row := m.pool.QueryRow(ctx, query)
	var exists bool
	err := row.Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil
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
