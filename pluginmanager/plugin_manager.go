package pluginmanager

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	sdkshared "github.com/turbot/steampipe-plugin-sdk/grpc/shared"
	pb "github.com/turbot/steampipe/pluginmanager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/pluginmanager/grpc/shared"
	"github.com/turbot/steampipe/utils"
)

type runningPlugin struct {
	client      *plugin.Client
	reattach    *pb.ReattachConfig
	initialized chan (bool)
}

// PluginManager is the real implementation of grpc.PluginManager
type PluginManager struct {
	pb.UnimplementedPluginManagerServer

	Plugins          map[string]*runningPlugin
	mut              sync.Mutex
	connectionConfig map[string]*pb.ConnectionConfig
	logger           hclog.Logger
}

func NewPluginManager(connectionConfig map[string]*pb.ConnectionConfig, logger hclog.Logger) *PluginManager {
	pluginManager := &PluginManager{
		logger:           logger,
		connectionConfig: connectionConfig,
		Plugins:          make(map[string]*runningPlugin),
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

	log.Printf("[TRACE] PluginManager Get, connections: '%s'\n", req.Connections)
	for _, c := range req.Connections {
		resultWg.Add(1)
		go func(connectionName string) {
			reattach, err := m.getPlugin(connectionName)

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
	log.Printf("[TRACE] PluginManager get returning %+v", resp)
	return resp, nil
}

func (m *PluginManager) getPlugin(connection string) (_ *pb.ReattachConfig, err error) {
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
	client, err := m.startPlugin(connection)
	if err != nil {
		return nil, err
	}

	// store the client to our map
	reattach := m.storeClientToMap(connection, client)
	log.Printf("[TRACE] PluginManager Get complete, returning reattach config with PID: %d", reattach.Pid)

	// and return
	return reattach, nil
}

// create reattach config for plugin, store to map and close initialized channel
func (m *PluginManager) storeClientToMap(connection string, client *plugin.Client) *pb.ReattachConfig {
	// lock access to map
	m.mut.Lock()
	defer m.mut.Unlock()

	reattach := pb.NewReattachConfig(client.ReattachConfig())
	p := m.Plugins[connection]
	p.client = client
	p.reattach = reattach
	m.Plugins[connection] = p
	// mark as initialized
	close(p.initialized)
	return reattach
}

func (m *PluginManager) SetConnectionConfigMap(configMap map[string]*pb.ConnectionConfig) {
	m.mut.Lock()
	defer m.mut.Unlock()

	m.connectionConfig = configMap
}

func (m *PluginManager) Shutdown(req *pb.ShutdownRequest) (resp *pb.ShutdownResponse, err error) {
	log.Printf("[TRACE] PluginManager Shutdown %v", m.Plugins)

	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	for _, p := range m.Plugins {
		log.Printf("[TRACE] killing plugin %v", p.reattach.Pid)
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
