package plugin_manager

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	sdkshared "github.com/turbot/steampipe-plugin-sdk/grpc/shared"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
	"github.com/turbot/steampipe/utils"
)

type runningPlugin struct {
	client   *plugin.Client
	reattach *pb.ReattachConfig
}

// PluginManager is the real implementation of grpc.PluginManager
type PluginManager struct {
	pb.UnimplementedPluginManagerServer

	Plugins          map[string]runningPlugin
	configDir        string
	mut              sync.Mutex
	connectionConfig map[string]*pb.ConnectionConfig
	logger           hclog.Logger
}

func NewPluginManager(connectionConfig map[string]*pb.ConnectionConfig, logger hclog.Logger) *PluginManager {
	pluginManager := &PluginManager{
		logger:           logger,
		connectionConfig: connectionConfig,
		Plugins:          make(map[string]runningPlugin),
	}
	return pluginManager
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
	resp := &pb.GetResponse{ReattachMap: make(map[string]*pb.ReattachConfig)}
	var errors []error
	var resultLock sync.Mutex
	var resultWg sync.WaitGroup

	for _, c := range req.Connections {
		resultWg.Add(1)
		go func(connectionName string) {
			reattach, err := m.getPlugin(c)

			resultLock.Lock()
			if err != nil {
				errors = append(errors, err)
				resultLock.Unlock()
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

	return resp, nil
}

func (m *PluginManager) getPlugin(connection string) (_ *pb.ReattachConfig, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	log.Printf("[TRACE] PluginManager Get connection '%s'\n", connection)

	// reason for starting the plugin (if we need to
	var reason string

	// is this plugin already running
	m.mut.Lock()
	p, ok := m.Plugins[connection]
	m.mut.Unlock()

	if !ok {
		reason = fmt.Sprintf("PluginManager %p '%s' NOT found in map  - starting", m, connection)
	} else {
		// so we have the plugin in our map - does it exist

		reattach := p.reattach
		// check the pid exists
		exists, _ := utils.PidExists(int(reattach.Pid))
		if exists {
			// so the plugin id good
			log.Printf("[TRACE] PluginManager found '%s' in map", connection)

			// return the reattach config
			return reattach, nil
		}

		//  either the pid does not exist or the plugin has exited
		// remove from map
		m.mut.Lock()
		delete(m.Plugins, connection)
		m.mut.Unlock()
		// update reason
		reason = fmt.Sprintf("PluginManager found pid %d for connection '%s' in plugin map but plugin process does not exist - killing client and removing from map", reattach.Pid, connection)
	}

	// fall through to plugin startup
	// log the startup reason
	log.Printf("[TRACE] %s", reason)
	// so we need to start the plugin
	client, err := m.startPlugin(connection)
	if err != nil {
		return nil, err
	}

	// store the client to our map
	reattach := pb.NewReattachConfig(client.ReattachConfig())
	m.mut.Lock()
	m.Plugins[connection] = runningPlugin{client: client, reattach: reattach}
	m.mut.Unlock()
	log.Printf("[TRACE] PluginManager Get complete, returning reattach config with PID: %d", reattach.Pid)

	// and return
	return reattach, nil
}

func (m *PluginManager) SetConnectionConfigMap(configMap map[string]*pb.ConnectionConfig) {
	m.mut.Lock()
	defer m.mut.Unlock()

	m.connectionConfig = configMap
}

func (m *PluginManager) Shutdown(req *pb.ShutdownRequest) (resp *pb.ShutdownResponse, err error) {
	log.Printf("[TRACE] PluginManager Shutdown")

	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	for _, p := range m.Plugins {
		log.Printf("[TRACE] killing plugin %v", p)
		p.client.Kill()
	}
	return &pb.ShutdownResponse{}, nil
}

func (m *PluginManager) startPlugin(connection string) (*plugin.Client, error) {

	log.Printf("[TRACE] ************ start plugin %s ********************\n", connection)

	// get connection config
	connectionConfig, ok := m.connectionConfig[connection]
	if !ok {
		return nil, fmt.Errorf("no config loaded for connection %s", connection)
	}

	pluginPath, err := GetPluginPath(connectionConfig.Plugin, connectionConfig.PluginShortName)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return client, nil
}
