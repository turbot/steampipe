package pluginmanager_service

import (
	"context"
	"crypto/md5"
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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sethvargo/go-retry"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	sdkshared "github.com/turbot/steampipe-plugin-sdk/v5/grpc/shared"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/connection"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/pluginmanager_service/grpc"
	pb "github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/proto"
	pluginshared "github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/shared"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

// PluginManager is the implementation of grpc.PluginManager
type PluginManager struct {
	pb.UnimplementedPluginManagerServer

	// map of running plugins keyed by plugin label
	runningPluginMap map[string]*runningPlugin
	// map of connection configs, keyed by plugin label
	// this is populated at startup and updated when a connection config change is detected
	pluginConnectionConfigMap map[string][]*sdkproto.ConnectionConfig
	// map of connection configs, keyed by connection name
	// this is populated at startup and updated when a connection config change is detected
	connectionConfigMap connection.ConnectionConfigMap
	// map of max cache size, keyed by plugin label
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

	// map of user configured rate limiter maps, keyed by plugin label
	// NOTE: this is populated from config
	userLimiters connection.PluginLimiterMap
	// map of plugin configured rate limiter maps  (keyed by plugin label)
	// NOTE: if this is nil, that means the steampipe_rate_limiter tables has not been populated yet -
	// the first time we refresh connections we must load all plugins and fetch their rate limiter defs
	pluginLimiters connection.PluginLimiterMap

	// map of plugin configs (keyed by plugin label)
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

func (m *PluginManager) Get(req *pb.GetRequest) (_ *pb.GetResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = sperr.ToError(r, sperr.WithMessage("unexpected error encountered"))
		}
	}()
	log.Printf("[TRACE] PluginManager Get %p", req)
	defer log.Printf("[TRACE] PluginManager Get DONE %p", req)

	resp := &pb.GetResponse{
		ReattachMap: make(map[string]*pb.ReattachConfig),
		FailureMap:  make(map[string]string),
	}

	// build a map of plugins to connection config for requested connections, and a lookup of the requested connections
	plugins, requestedConnectionsLookup, err := m.buildRequiredPluginMap(req)
	if err != nil {
		return resp, err
	}

	log.Printf("[TRACE] PluginManager Get, connections: '%s'\n", req.Connections)
	for pluginLabel, connectionConfigs := range plugins {
		// ensure plugin is running
		reattach, err := m.ensurePlugin(pluginLabel, connectionConfigs, req)
		if err != nil {
			log.Printf("[WARN] PluginManager Get failed for %s: %s (%p)", pluginLabel, err.Error(), resp)
			resp.FailureMap[pluginLabel] = sperr.WrapWithMessage(err, "failed to start '%s'", pluginLabel).Error()
		} else {
			log.Printf("[TRACE] PluginManager Get succeeded for %s, pid %d (%p)", pluginLabel, reattach.Pid, resp)

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

// build a map of plugins to connection config for requested connections, keyed by plugin label,
// and a lookup of the requested connections
func (m *PluginManager) buildRequiredPluginMap(req *pb.GetRequest) (map[string][]*sdkproto.ConnectionConfig, map[string]struct{}, error) {
	var plugins = make(map[string][]*sdkproto.ConnectionConfig)
	// also make a map of target connections - used when assigning resuts to the response
	var requestedConnectionsLookup = make(map[string]struct{}, len(req.Connections))
	for _, connectionName := range req.Connections {
		// store connection in requested connection map
		requestedConnectionsLookup[connectionName] = struct{}{}

		connectionConfig, err := m.getConnectionConfig(connectionName)
		if err != nil {
			return nil, nil, err
		}
		pluginLabel := connectionConfig.PluginInstance
		// if we have not added this plugin label, add it now
		if _, addedPlugin := plugins[pluginLabel]; !addedPlugin {
			// now get ALL connection configs for this plugin
			// (not just the requested connections)
			plugins[pluginLabel] = m.pluginConnectionConfigMap[pluginLabel]
		}
	}
	return plugins, requestedConnectionsLookup, nil
}

func (m *PluginManager) Pool() *pgxpool.Pool {
	return m.pool
}

func (m *PluginManager) RefreshConnections(*pb.RefreshConnectionsRequest) (*pb.RefreshConnectionsResponse, error) {
	resp := &pb.RefreshConnectionsResponse{}
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
func (m *PluginManager) OnConnectionConfigChanged(configMap connection.ConnectionConfigMap, plugins map[string]*modconfig.Plugin) {
	m.mut.Lock()
	defer m.mut.Unlock()

	names := utils.SortedMapKeys(configMap)
	log.Printf("[TRACE] OnConnectionConfigChanged: %s", strings.Join(names, ","))

	err := m.handleConnectionConfigChanges(configMap)
	if err != nil {
		log.Printf("[WARN] handleConnectionConfigChanges failed: %s", err.Error())
	}

	// update our plugin configs
	m.plugins = plugins
	err = m.handleUserLimiterChanges(plugins)
	if err != nil {
		log.Printf("[WARN] handleUserLimiterChanges failed: %s", err.Error())
	}
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
	m.startPluginWg.Wait()

	// close our pool
	log.Printf("[INFO] PluginManager closing pool")
	m.pool.Close()

	m.mut.RLock()
	defer func() {
		m.mut.RUnlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	// kill all plugins in pluginMultiConnectionMap
	for _, p := range m.runningPluginMap {
		log.Printf("[INFO] Kill plugin %s (%p)", p.pluginLabel, p.client)
		m.killPlugin(p)
	}

	return &pb.ShutdownResponse{}, nil
}

func (m *PluginManager) killPlugin(p *runningPlugin) {
	log.Println("[DEBUG] PluginManager killPlugin start")
	defer log.Println("[DEBUG] PluginManager killPlugin complete")

	if p.client == nil {
		log.Printf("[WARN] plugin %s has no client - cannot kill client", p.pluginLabel)
		// shouldn't happen but has been observed in error situations
		return
	}
	log.Printf("[INFO] PluginManager killing plugin %s (%v)", p.pluginLabel, p.reattach.Pid)
	p.client.Kill()
}

func (m *PluginManager) ensurePlugin(pluginLabel string, connectionConfigs []*sdkproto.ConnectionConfig, req *pb.GetRequest) (reattach *pb.ReattachConfig, err error) {
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
	if m.shuttingDown() {
		return nil, fmt.Errorf("plugin manager is shutting down")
	}

	log.Printf("[TRACE] PluginManager ensurePlugin %s (%p)", pluginLabel, req)

	err = retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		reattach, err = m.startPluginIfNeeded(pluginLabel, connectionConfigs, req)
		return err
	})

	return
}

func (m *PluginManager) startPluginIfNeeded(pluginLabel string, connectionConfigs []*sdkproto.ConnectionConfig, req *pb.GetRequest) (*pb.ReattachConfig, error) {
	// is this plugin already running
	// lock access to plugin map
	m.mut.RLock()
	startingPlugin, ok := m.runningPluginMap[pluginLabel]
	m.mut.RUnlock()

	if ok {
		log.Printf("[TRACE] startPluginIfNeeded got running plugin (%p)", req)

		// wait for plugin to process connection config, and verify it is running
		err := m.waitForPluginLoad(startingPlugin, req)
		if err == nil {
			// so plugin has loaded - we are done
			log.Printf("[TRACE] waitForPluginLoad succeeded %s (%p)", pluginLabel, req)
			return startingPlugin.reattach, nil
		}
		log.Printf("[TRACE] waitForPluginLoad failed %s (%p)", err.Error(), req)

		// just return the error
		return nil, err
	}

	// so the plugin is NOT loaded or loading
	// fall through to plugin startup
	log.Printf("[INFO] plugin %s NOT started or starting - start now (%p)", pluginLabel, req)

	return m.startPlugin(pluginLabel, connectionConfigs, req)
}

func (m *PluginManager) startPlugin(pluginLabel string, connectionConfigs []*sdkproto.ConnectionConfig, req *pb.GetRequest) (_ *pb.ReattachConfig, err error) {
	log.Printf("[DEBUG] startPlugin %s (%p) start", pluginLabel, req)
	defer log.Printf("[DEBUG] startPlugin %s (%p) end", pluginLabel, req)

	// add a new running plugin to pluginMultiConnectionMap
	// (if someone beat us to it and added a starting plugin before we get the write lock,
	// this will return a retryable error)
	startingPlugin, err := m.addRunningPlugin(pluginLabel)
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
			delete(m.runningPluginMap, pluginLabel)
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
	client, err := m.startPluginProcess(pluginLabel, connectionConfigs)
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

func (m *PluginManager) addRunningPlugin(pluginLabel string) (*runningPlugin, error) {
	// add a new running plugin to pluginMultiConnectionMap
	// this is a placeholder so no other thread tries to create start this plugin

	// acquire write lock
	m.mut.Lock()
	defer m.mut.Unlock()
	log.Printf("[TRACE] add running plugin for %s (if someone didn't beat us to it)", pluginLabel)

	// check someone else has beaten us to it (there is a race condition to starting a plugin)
	if _, ok := m.runningPluginMap[pluginLabel]; ok {
		log.Printf("[TRACE] re checked map and found a starting plugin - return retryable error so we wait for this plugin")
		// if so, just retry, which will wait for the loading plugin
		return nil, retry.RetryableError(fmt.Errorf("another client has already started the plugin"))
	}

	// create the running plugin
	startingPlugin := &runningPlugin{
		pluginLabel: pluginLabel,
		initialized: make(chan struct{}),
		failed:      make(chan struct{}),
	}
	// write back
	m.runningPluginMap[pluginLabel] = startingPlugin

	log.Printf("[INFO] written running plugin to map")

	return startingPlugin, nil
}

func (m *PluginManager) startPluginProcess(pluginLabel string, connectionConfigs []*sdkproto.ConnectionConfig) (*plugin.Client, error) {
	// retrieve the plugin config
	pluginConfig := m.plugins[pluginLabel]
	// must be there (if no explicit config was specified, we create a default)
	if pluginConfig == nil {
		panic(fmt.Sprintf("no plugin config is stored for plugin label %s", pluginLabel))
	}

	imageRef := pluginConfig.GetImageRef()
	log.Printf("[INFO] ************ start plugin: %s, label: %s ********************\n", imageRef, pluginConfig.Instance)

	// NOTE: pass pluginConfig.Source as the pluginAlias
	// - this is just used for the error message if we fail to load
	pluginPath, err := filepaths.GetPluginPath(imageRef, pluginConfig.Source)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] ************ plugin path %s ********************\n", pluginPath)

	// create the plugin map
	pluginMap := map[string]plugin.Plugin{
		imageRef: &sdkshared.WrapperPlugin{},
	}

	utils.LogTime("getting plugin exec hash")
	pluginChecksum, err := helpers.FileMD5Hash(pluginPath)
	if err != nil {
		return nil, err
	}
	utils.LogTime("got plugin exec hash")
	cmd := exec.Command(pluginPath)

	m.setPluginMaxMemory(pluginConfig, cmd)

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
		// attempt to retrieve error message encoded in the plugin stdout
		err := grpc.HandleStartFailure(err)
		return nil, err
	}

	return client, nil

}

func (m *PluginManager) setPluginMaxMemory(pluginConfig *modconfig.Plugin, cmd *exec.Cmd) {
	maxMemoryBytes := pluginConfig.GetMaxMemoryBytes()
	if maxMemoryBytes == 0 {
		if viper.IsSet(constants.ArgMemoryMaxMbPlugin) {
			maxMemoryBytes = viper.GetInt64(constants.ArgMemoryMaxMbPlugin) * 1024 * 1024
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
func (m *PluginManager) initializePlugin(connectionConfigs []*sdkproto.ConnectionConfig, client *plugin.Client, req *pb.GetRequest) (_ *pb.ReattachConfig, err error) {
	// extract connection names
	connectionNames := make([]string, len(connectionConfigs))
	for i, c := range connectionConfigs {
		connectionNames[i] = c.Connection
	}
	exemplarConnectionConfig := connectionConfigs[0]
	pluginName := exemplarConnectionConfig.Plugin
	pluginShortName := exemplarConnectionConfig.PluginShortName

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
		// TODO SEND NOTIFICATION TO CLI
		return nil, fmt.Errorf("plugins which do not support multiple connections (using SDK version < v4) are no longer supported. Upgrade plugin '%s", pluginName)
	}

	// provide opportunity to avoid setting connection configs if we are shutting down
	if m.shuttingDown() {
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
		err = m.setRateLimiters(pluginShortName, pluginClient)
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
func (m *PluginManager) shuttingDown() bool {
	if !m.shutdownMut.TryLock() {
		return true
	}
	m.shutdownMut.Unlock()
	return false
}

// populate map of connection configs for each plugin label
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

	// get total connection count for this pluginLabel (excluding aggregators)
	numConnections := m.nonAggregatorConnectionCount()

	log.Printf("[TRACE] PluginManager setPluginCacheSizeMap: %d %s.", numConnections, utils.Pluralize("connection", numConnections))
	log.Printf("[TRACE] Total cache size %dMb", maxCacheSizeMb)

	for pluginLabel, connections := range m.pluginConnectionConfigMap {
		var size int64 = 0
		// if no max size is set, just set all plugins to zero (unlimited)
		if maxCacheSizeMb > 0 {
			// get connection count for this pluginLabel (excluding aggregators)
			numPluginConnections := nonAggregatorConnectionCount(connections)
			size = int64(float64(numPluginConnections) / float64(numConnections) * float64(maxCacheSizeMb))
			// make this at least 1 Mb (as zero means unlimited)
			if size == 0 {
				size = 1
			}
			log.Printf("[INFO] Plugin '%s', %d %s, max cache size %dMb", pluginLabel, numPluginConnections, utils.Pluralize("connection", numPluginConnections), size)
		}

		m.pluginCacheSizeMap[pluginLabel] = size
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
	log.Printf("[TRACE] waitForPluginLoad (%p)", req)
	// TODO make this configurable
	pluginStartTimeoutSecs := 30

	// wait for the plugin to be initialized
	select {
	case <-time.After(time.Duration(pluginStartTimeoutSecs) * time.Second):
		log.Printf("[WARN] timed out waiting for %s to startup after %d seconds (%p)", p.pluginLabel, pluginStartTimeoutSecs, req)
		// do not retry
		return fmt.Errorf("timed out waiting for %s to startup after %d seconds (%p)", p.pluginLabel, pluginStartTimeoutSecs, req)
	case <-p.initialized:
		log.Printf("[TRACE] plugin initialized: pid %d (%p)", p.reattach.Pid, req)
	case <-p.failed:
		log.Printf("[TRACE] plugin pid %d failed %s (%p)", p.reattach.Pid, p.error.Error(), req)
		// get error from running plugin
		return p.error
	}

	// now double-check the plugins process IS running
	if !p.client.Exited() {
		// so the plugin is good
		log.Printf("[INFO] waitForPluginLoad: %s is now loaded and ready (%p)", p.pluginLabel, req)
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
	if r, ok := m.runningPluginMap[p.pluginLabel]; ok {
		// is the running plugin we read from the map the same as our running plugin?
		// if not, it must already have been removed by another thread - do nothing
		if r == p {
			log.Printf("[INFO] delete plugin %s from runningPluginMap (%p)", p.pluginLabel, req)
			delete(m.runningPluginMap, p.pluginLabel)
		}
	}
	m.mut.Unlock()

	// so the pid does not exist
	err := fmt.Errorf("PluginManager found pid %d for plugin '%s' in plugin map but plugin process does not exist (%p)", p.reattach.Pid, p.pluginLabel, req)
	// we need to start the plugin again - make the error retryable
	return retry.RetryableError(err)
}

// set connection config for multiple connection
// NOTE: we DO NOT set connection config for aggregator connections
func (m *PluginManager) setAllConnectionConfigs(connectionConfigs []*sdkproto.ConnectionConfig, pluginClient *sdkgrpc.PluginClient, supportedOperations *sdkproto.GetSupportedOperationsResponse) error {
	// TODO does this fail all connections if one fails
	exemplarConnectionConfig := connectionConfigs[0]
	pluginLabel := exemplarConnectionConfig.PluginInstance

	req := &sdkproto.SetAllConnectionConfigsRequest{
		Configs: connectionConfigs,
		// NOTE: set MaxCacheSizeMb to -1so that query cache is not created until we call SetCacheOptions (if supported)
		MaxCacheSizeMb: -1,
	}
	// if plugin _does not_ support setting the cache options separately, pass the max size now
	// (if it does support SetCacheOptions, it will be called after we return)
	if !supportedOperations.SetCacheOptions {
		req.MaxCacheSizeMb = m.pluginCacheSizeMap[pluginLabel]
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

func (m *PluginManager) setRateLimiters(pluginName string, pluginClient *sdkgrpc.PluginClient) error {
	log.Printf("[INFO] setRateLimiters for plugin '%s'", pluginName)
	var defs []*sdkproto.RateLimiterDefinition

	for _, l := range m.userLimiters[pluginName] {
		defs = append(defs, l.AsProto())
	}

	req := &sdkproto.SetRateLimitersRequest{Definitions: defs}

	_, err := pluginClient.SetRateLimiters(req)
	return err
}

// update the schema for the specified connection
// called from the message server after receiving a PluginMessageType_SCHEMA_UPDATED message from plugin
func (m *PluginManager) updateConnectionSchema(ctx context.Context, connectionName string) {
	log.Printf("[TRACE] updateConnectionSchema connection %s", connectionName)

	refreshResult := connection.RefreshConnections(ctx, m, connectionName)
	if refreshResult.Error != nil {
		log.Printf("[TRACE] error refreshing connections: %s", refreshResult.Error)
		return
	}

	// also send a postgres notification
	notification := steampipeconfig.NewSchemaUpdateNotification()

	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		log.Printf("[WARN] failed to send schema update notification: %s", err)
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

// getPluginExemplarConnections returns a map of keyed by plugin short name with the value an exemplar connection
func (m *PluginManager) getPluginExemplarConnections() map[string]string {
	res := make(map[string]string)
	for _, c := range m.connectionConfigMap {
		res[c.PluginShortName] = c.Connection
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
