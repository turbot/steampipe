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

	log.Printf("[TRACE] PluginManager %p Get connection '%s', plugins %+v\n", m, req.Connection, m.Plugins)

	// is this plugin already running
	if p, ok := m.Plugins[req.Connection]; ok {
		reattach := p.reattach
		// check the pid exists and the clien thas not exited (i.e. th eplugin has crashed)
		exists, _ := utils.PidExists(int(reattach.Pid))
		exited := p.client.Exited()
		if exists && !exited {
			// so the plugin id good
			log.Printf("[WARN] PluginManager %p found '%s' in map %v", m, req.Connection, m.Plugins)

			// return the reattach config
			return &pb.GetResponse{
				Reattach: reattach,
			}, nil
		}

		//  either the pid does not exist or the plugin has exited

		// kill the client
		p.client.Kill()
		// remove from map
		delete(m.Plugins, req.Connection)

		// build log string
		var reason string
		if exited {
			reason = "client has exited"
			// TODO do we need to kill the PID?
		} else {
			reason = "pid does not exist"
		}
		log.Printf("[WARN] PluginManager %p plugin pid %d for connection '%s' found in plugin map but %s - killing client and removing from map", m, reattach.Pid, req.Connection, reason)

	} else {
		log.Printf("[WARN] PluginManager %p '%s' NOT found in map %v - starting", m, req.Connection, m.Plugins)
	}

	// fall through to plugin startup
	// so we need to start the plugin
	client, err := m.startPlugin(req)
	if err != nil {
		return nil, err
	}

	// TODO ADD PLUGIN TO OUR STATE FILE - JUST SERIALISE THE Plugins map?

	// store the client to our map
	reattach := pb.NewReattachConfig(client.ReattachConfig())
	m.Plugins[req.Connection] = runningPlugin{client: client, reattach: reattach}

	log.Printf("[TRACE] PluginManager %p Get complete", m)

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

func (m *PluginManager) Shutdown(*pb.ShutdownRequest) (resp *pb.ShutdownResponse, err error) {
	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	var errs []error
	for _, p := range m.Plugins {
		log.Printf("[WARN] kill %v", p)
		err = m.killPlugin(p.reattach.Pid)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return &pb.ShutdownResponse{}, utils.CombineErrorsWithPrefix(fmt.Sprintf("failed to shutdown %d plugins", len(errs)), errs...)
}

func (m *PluginManager) startPlugin(req *pb.GetRequest) (*plugin.Client, error) {

	log.Printf("[WARN] startPlugin ********************\n")

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
		Logger:           m.logger,
	})

	if _, err := client.Start(); err != nil {
		return nil, err
	}
	return client, nil
	/* hub did this

	// loop as we may need to retry if the plugin exists in the map but has actually exited
	const maxAttempts = 3
	for attempt := 1; attempt < maxAttempts; attempt++ {
		// ask connection map to get or create this connection
		c, err := h.connections.get(pluginFQN, connectionName)
		if err != nil {
			return nil, err
		}

		// make sure that the plugin is running
		// (i.e. it has not crashed)
		if !c.Plugin.Client.Exited() {
			// it is running, return it
			return c, nil
		}

		// remove connection from the connection map and kill the GRPC client
		h.connections.removeAndKill(pluginFQN, connectionName)
	}
	*/
}

func (m *PluginManager) killPlugin(pid int64) error {
	process, err := utils.FindProcess(int(pid))
	if err != nil {
		log.Printf("[WARN] error finding process %d", pid)
		return err
	}
	if process == nil {
		return nil
	}
	return process.Kill()
}
