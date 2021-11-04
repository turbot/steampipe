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
	return &PluginManager{
		logger:           logger,
		connectionConfig: connectionConfig,
		Plugins:          make(map[string]runningPlugin),
	}
}

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

// plugin interface functions

func (m *PluginManager) Get(req *pb.GetRequest) (resp *pb.GetResponse, err error) {
	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	log.Printf("[TRACE] PluginManager Get connection '%s', plugins %+v\n", req.Connection, m.Plugins)

	// reason for starting the plugin (if we need to
	var reason string

	// is this plugin already running
	if p, ok := m.Plugins[req.Connection]; !ok {
		reason = fmt.Sprintf("PluginManager %p '%s' NOT found in map %v - starting", m, req.Connection, m.Plugins)
	} else {
		// so we have the plugin in our map - does it exist

		reattach := p.reattach
		// check the pid exists
		exists, _ := utils.PidExists(int(reattach.Pid))
		if exists {
			// so the plugin id good
			log.Printf("[TRACE] PluginManager found '%s' in map %v", req.Connection, m.Plugins)

			// return the reattach config
			return &pb.GetResponse{
				Reattach: reattach,
			}, nil
		}

		//  either the pid does not exist or the plugin has exited
		// remove from map
		delete(m.Plugins, req.Connection)
		// update reason
		reason = fmt.Sprintf("PluginManager found pid %d for connection '%s' in plugin map but plugin process does not exist - killing client and removing from map", reattach.Pid, req.Connection)
	}

	// fall through to plugin startup
	// log the startup reason
	log.Printf("[TRACE] %s", reason)
	// so we need to start the plugin
	client, err := m.startPlugin(req)
	if err != nil {
		return nil, err
	}

	// TODO ADD PLUGIN TO OUR STATE FILE - JUST SERIALISE THE Plugins map?

	// store the client to our map
	reattach := pb.NewReattachConfig(client.ReattachConfig())
	m.Plugins[req.Connection] = runningPlugin{client: client, reattach: reattach}

	log.Printf("[TRACE] PluginManager Get complete")

	// and return
	return &pb.GetResponse{Reattach: reattach}, nil

}

func (m *PluginManager) SetConnectionConfigMap(req *pb.SetConnectionConfigMapRequest) (resp *pb.SetConnectionConfigMapResponse, err error) {
	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()
	m.connectionConfig = req.ConfigMap
	return &pb.SetConnectionConfigMapResponse{}, nil
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

func (m *PluginManager) startPlugin(req *pb.GetRequest) (*plugin.Client, error) {

	log.Printf("[TRACE] ************ start plugin %s ********************\n", req.Connection)

	// get connection config
	connectionConfig, ok := m.connectionConfig[req.Connection]
	if !ok {
		return nil, fmt.Errorf("no config loaded for connection %s", req.Connection)
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
