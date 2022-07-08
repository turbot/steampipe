package pluginmanager

import (
	"fmt"
	"github.com/spf13/viper"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v3/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
	"github.com/turbot/steampipe/pkg/constants"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	sdkshared "github.com/turbot/steampipe-plugin-sdk/v3/grpc/shared"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pluginmanager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/pluginmanager/grpc/shared"
)

type runningPlugin struct {
	client   *plugin.Client
	reattach *proto.ReattachConfig
	// does this plugin support multiple connections - requires sdk version > 4
	multiConnection bool
	initialized     chan bool
}

// PluginManager is the real implementation of grpc.PluginManager
type PluginManager struct {
	proto.UnimplementedPluginManagerServer

	Plugins map[string]*runningPlugin

	mut                     sync.Mutex
	pluginConnectionConfigs map[string][]*sdkproto.ConnectionConfig
	connectionConfig        map[string]*sdkproto.ConnectionConfig
	logger                  hclog.Logger
	cacheManager            *CacheServer
}

func NewPluginManager(connectionConfig map[string]*sdkproto.ConnectionConfig, logger hclog.Logger) (*PluginManager, error) {
	log.Printf("[TRACE] NewPluginManager")
	pluginManager := &PluginManager{
		Plugins:                 make(map[string]*runningPlugin),
		logger:                  logger,
		connectionConfig:        connectionConfig,
		pluginConnectionConfigs: make(map[string][]*sdkproto.ConnectionConfig),
	}
	maxCacheStorageMb := viper.GetInt(constants.ArgMaxCacheSizeMb)
	cacheManager, err := NewCacheServer(maxCacheStorageMb, pluginManager)
	if err != nil {
		return nil, err
	}
	pluginManager.cacheManager = cacheManager

	// populate plugin connection config map
	pluginManager.setPluginConnectionConfigs()
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
	resp := &proto.GetResponse{ReattachMap: make(map[string]*proto.ReattachConfig)}
	var errors []error
	var resultLock sync.Mutex
	var resultWg sync.WaitGroup

	log.Printf("[TRACE] PluginManager Get, connections: '%s'\n", req.Connections)
	for _, c := range req.Connections {
		resultWg.Add(1)
		go func(connectionName string) {
			reattach, err := m.getPlugin(connectionName)

			resultLock.Lock()
			if err != nil {
				errors = append(errors, err)
			} else {
				resp.ReattachMap[connectionName] = reattach
			}
			resultLock.Unlock()
			resultWg.Done()
		}(c)
	}

	resultWg.Wait()

	if len(errors) > 0 {
		return nil, utils.CombineErrors(errors...)
	}

	// TODO ADD PLUGINS TO OUR STATE FILE - JUST SERIALISE THE Plugins map?
	log.Printf("[TRACE] PluginManager Get returning %+v", resp)
	return resp, nil
}

func (m *PluginManager) getPlugin(connection string) (_ *proto.ReattachConfig, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	log.Printf("[TRACE] PluginManager getPlugin connection '%s'\n", connection)

	// reason for starting the plugin (if we need to
	var reason string

	// is this plugin already running
	// lock access to plugin map
	m.mut.Lock()
	p, ok := m.Plugins[connection]
	if ok {
		// unlock access to map
		m.mut.Unlock()

		// so we have the plugin in our map - is it started?
		err = m.waitForPluginLoad(connection, p)
		if err != nil {
			return nil, err
		}
		log.Printf("[TRACE] connection %s is loaded, check for running PID", connection)

		// ok so the plugin should now be running

		// now check if the plugins process IS running
		reattach := p.reattach
		// check the pid exists
		exists, _ := utils.PidExists(int(reattach.Pid))
		if exists {
			// so the plugin is good
			log.Printf("[TRACE] PluginManager found '%s' in map, pid %d, reattach %v", connection, reattach.Pid, reattach)

			// return the reattach config
			return reattach, nil
		}

		//  either the pid does not exist or the plugin has exited
		// remove from map
		m.mut.Lock()
		delete(m.Plugins, connection)
		m.Plugins[connection] = &runningPlugin{
			initialized: make(chan (bool), 1),
		}
		m.mut.Unlock()
		// update reason
		reason = fmt.Sprintf("PluginManager found pid %d for connection '%s' in plugin map but plugin process does not exist - killing client and removing from map", reattach.Pid, connection)

	} else {
		// so the plugin is NOT loaded or loading - this is the first time anyone has requested this connection
		// put in a placeholder so no other thread tries to create start this plugin
		m.Plugins[connection] = &runningPlugin{
			initialized: make(chan (bool), 1),
		}

		// unlock access to map
		m.mut.Unlock()
		reason = fmt.Sprintf("PluginManager %p '%s' NOT found in map  - starting", m, connection)
	}

	// fall through to plugin startup
	// log the startup reason
	log.Printf("[TRACE] %s", reason)
	// so we need to start the plugin
	client, reattach, err := m.startPlugin(connection)
	if err != nil {
		m.mut.Lock()
		delete(m.Plugins, connection)
		m.mut.Unlock()

		log.Println("[TRACE] startPlugin failed with", err)
		return nil, err
	}

	// store the client to our map
	m.storeClientToMap(connection, client, reattach)
	log.Printf("[TRACE] PluginManager getPlugin complete, returning reattach config with PID: %d", reattach.Pid)

	// open the cache stream
	m.cacheManager.AddConnection(client, connection)

	// and return
	return reattach, nil
}

// create reattach config for plugin, store to map and close initialized channel
func (m *PluginManager) storeClientToMap(connection string, client *plugin.Client, reattach *proto.ReattachConfig) {
	// lock access to map
	m.mut.Lock()
	defer m.mut.Unlock()

	// a RunningPlugin in initializing state will already have been put into the Plugins map
	// populate its properties
	p := m.Plugins[connection]
	p.client = client
	p.reattach = reattach

	// NOTE: if this plugin supports multiple connections, reattach.Connections will be a list of all connections
	// provided by this plugin
	// add map entries for all other connections using this plugin (all pointing to same RunningPlugin)
	for _, c := range reattach.Connections {
		m.Plugins[c] = p
	}
	// mark as initialized
	close(p.initialized)
}

func (m *PluginManager) SetConnectionConfigMap(configMap map[string]*sdkproto.ConnectionConfig) {
	m.mut.Lock()
	defer m.mut.Unlock()

	names := make([]string, len(configMap))
	idx := 0
	for name := range configMap {
		names[idx] = name
		idx++
	}
	log.Printf("[TRACE] SetConnectionConfigMap: %s", strings.Join(names, ","))

	m.connectionConfig = configMap
}

// populate map of connection configs for each plugin
func (m *PluginManager) setPluginConnectionConfigs() {
	for _, config := range m.connectionConfig {
		m.pluginConnectionConfigs[config.Plugin] = append(m.pluginConnectionConfigs[config.Plugin], config)
	}
	log.Printf("[TRACE] setPluginConnectionConfigs: %v", m.pluginConnectionConfigs)
}

func (m *PluginManager) Shutdown(req *proto.ShutdownRequest) (resp *proto.ShutdownResponse, err error) {
	log.Printf("[TRACE] PluginManager Shutdown %v", m.Plugins)
	debug.PrintStack()

	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	for name, p := range m.Plugins {
		if p.client == nil {
			log.Printf("[WARN] plugin %s has no client - cannot kill", name)
			// shouldn't happen but has been observed in error situations
			continue
		}
		log.Printf("[TRACE] killing plugin %s (%v)", name, p.reattach.Pid)
		p.client.Kill()
	}
	return &proto.ShutdownResponse{}, nil
}

func (m *PluginManager) startPlugin(connection string) (*plugin.Client, *proto.ReattachConfig, error) {
	log.Printf("[TRACE] ************ start plugin %s ********************\n", connection)

	// get connection config
	connectionConfig, ok := m.connectionConfig[connection]
	if !ok {
		return nil, nil, fmt.Errorf("no config loaded for connection %s", connection)
	}

	pluginPath, err := GetPluginPath(connectionConfig.Plugin, connectionConfig.PluginShortName)
	if err != nil {
		return nil, nil, err
	}

	// create the plugin map
	pluginName := connectionConfig.Plugin
	pluginMap := map[string]plugin.Plugin{
		pluginName: &sdkshared.WrapperPlugin{},
	}

	cmd := exec.Command(pluginPath)

	// pass env to command
	cmd.Env = os.Environ()
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  sdkshared.Handshake,
		Plugins:          pluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		// pass our logger to the plugin client to ensure plugin logs end up in logfile
		Logger: m.logger,
	})

	if _, err := client.Start(); err != nil {
		return nil, nil, err
	}

	// get the supported operations
	pluginClient, err := sdkgrpc.NewPluginClient(client, pluginName)
	if err != nil {
		return nil, nil, err
	}

	supportedOperations, err := pluginClient.GetSupportedOperations()
	if err != nil {
		return nil, nil, err
	}
	log.Printf("[WARN] supportedOperations: %v", supportedOperations)
	var connections = []string{connection}

	if supportedOperations.MultipleConnections {
		// send the connection config for all connections for this plugin
		m.setAllConnectionConfigs(pluginClient, pluginName)
	} else {
		// send the connection config using legacy single connection function
		m.setSingleConnectionConfig(pluginClient, connection)
	}

	reattach := proto.NewReattachConfig(client.ReattachConfig(), proto.SupportedOperationsFromSdk(supportedOperations), connections)

	return client, reattach, nil
}

func (m *PluginManager) waitForPluginLoad(connection string, p *runningPlugin) error {
	pluginStartTimeoutSecs := 5

	select {
	case <-p.initialized:
		log.Printf("[TRACE] initialized: %d", p.reattach.Pid)
		return nil

	case <-time.After(time.Duration(pluginStartTimeoutSecs) * time.Second):
		return fmt.Errorf("timed out waiting for %s to startup after %d seconds", connection, pluginStartTimeoutSecs)
	}
}

func (m *PluginManager) getConnectionsForPlugin(pluginName string) []string {
	var res = make([]string, len(m.pluginConnectionConfigs[pluginName]))
	for i, c := range m.pluginConnectionConfigs[pluginName] {
		res[i] = c.Connection
	}
	return res
}

// set connection config for multiple connection, for compatible plugins)
func (m *PluginManager) setAllConnectionConfigs(pluginClient *sdkgrpc.PluginClient, pluginName string) error {
	configs, ok := m.pluginConnectionConfigs[pluginName]
	if !ok {
		// should never happen
		return fmt.Errorf("no config loaded for plugin '%s'", pluginName)
	}
	req := &sdkproto.SetAllConnectionConfigsRequest{
		Configs: configs,
	}

	return pluginClient.SetAllConnectionConfigs(req)
}

// set connection config for single connection, for legacy plugins)
func (m *PluginManager) setSingleConnectionConfig(pluginClient *sdkgrpc.PluginClient, connectionName string) error {
	connectionConfig, ok := m.connectionConfig[connectionName]
	if !ok {
		// should never happen
		return fmt.Errorf("no config loaded for connection '%s'", connectionName)
	}
	// set the connection config
	req := &sdkproto.SetConnectionConfigRequest{
		ConnectionName:   connectionName,
		ConnectionConfig: connectionConfig.Config,
	}

	return pluginClient.SetConnectionConfig(req)
}
