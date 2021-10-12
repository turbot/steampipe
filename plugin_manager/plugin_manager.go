package plugin_manager

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/turbot/steampipe-plugin-sdk/logging"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	sdkpluginshared "github.com/turbot/steampipe-plugin-sdk/grpc/shared"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

type RunningPlugin struct {
	reattach *pb.ReattachConfig
	refCount int
}

// PluginManager is the real implementation of grpc.PluginManager
type PluginManager struct {
	pb.UnimplementedPluginManagerServer

	Plugins        map[string]*RunningPlugin
	InvalidPlugins map[string]*RunningPlugin

	configDir        string
	mut              sync.Mutex
	connectionConfig map[string]*pb.ConnectionConfig
}

func NewPluginManager(connectionConfig map[string]*pb.ConnectionConfig) *PluginManager {
	return &PluginManager{
		connectionConfig: connectionConfig,
		Plugins:          make(map[string]*RunningPlugin),
		InvalidPlugins:   make(map[string]*RunningPlugin),
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

	// is this plugin already running
	if plugin, ok := m.Plugins[req.Connection]; ok {
		// inc the ref count
		plugin.refCount++
		// return the reattach config
		return &pb.GetResponse{
			Reattach: plugin.reattach,
		}, nil
	}

	// so we need to start the plugin
	reattach, err := m.startPlugin(req)
	if err != nil {
		return nil, err
	}

	// store the reattach config in our map
	m.Plugins[req.Connection] = &RunningPlugin{
		reattach: reattach,
		refCount: 1,
	}

	// and return
	return &pb.GetResponse{
		Reattach: reattach,
	}, nil

}

func (m *PluginManager) Release(req *pb.ReleaseRequest) (resp *pb.ReleaseResponse, err error) {
	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	// first look for plugin in map of valid plugins
	foundPlugin, err := m.decrementPluginUsage(req, m.Plugins)
	if err != nil {
		return nil, err
	}
	if foundPlugin {
		return &pb.ReleaseResponse{}, nil
	}

	// now try invalid plugins
	foundPlugin, err = m.decrementPluginUsage(req, m.InvalidPlugins)
	if err != nil {
		return nil, err
	}
	if foundPlugin {
		return &pb.ReleaseResponse{}, nil
	}

	// we could not find the plugin
	return nil, fmt.Errorf("no plugin found for connection %s, pid %s")
}

func (m *PluginManager) Reload(req *pb.ReloadRequest) (resp *pb.ReloadResponse, err error) {
	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()
	return nil, nil
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
	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	var errs []error
	for _, p := range m.Plugins {
		err = m.killPlugin(p.reattach.Pid)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return &pb.ShutdownResponse{}, utils.CombineErrorsWithPrefix(fmt.Sprintf("failed to shutdown %d plugins", len(errs)), errs...)
}

func (m *PluginManager) startPlugin(req *pb.GetRequest) (*pb.ReattachConfig, error) {
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
		pluginName: &sdkpluginshared.WrapperPlugin{},
	}
	loggOpts := &hclog.LoggerOptions{Name: "plugin"}
	if req.DisableLogger {
		loggOpts.Exclude = func(hclog.Level, string, ...interface{}) bool { return true }
	}
	logger := logging.NewLogger(loggOpts)

	cmd := exec.Command(pluginPath)
	// pass env to command
	cmd.Env = os.Environ()
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  sdkpluginshared.Handshake,
		Plugins:          pluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})

	if _, err := client.Start(); err != nil {
		return nil, err
	}
	reattach := client.ReattachConfig()
	return pb.NewReattachConfig(reattach), nil
}

//func (m *PluginManager) SetConnectionConfigMap(connectionName string, connectionConfig string, client *sdkgrpc.PluginClient) error {
//	req := &proto.SetConnectionConfigMapRequest{
//		ConnectionName:   connectionName,
//		ConnectionConfig: connectionConfig,
//	}
//
//	_, err := client.Stub.SetConnectionConfigMap(req)
//	if err != nil {
//		// create a new cleaner error, ignoring Not Implemented errors for backwards compatibility
//		return utils.HandleGrpcError(err, connectionName, "SetConnectionConfigMap")
//	}
//	return nil
//
//}

func (m *PluginManager) decrementPluginUsage(req *pb.ReleaseRequest, mp map[string]*RunningPlugin) (bool, error) {
	if plugin, ok := mp[req.Connection]; ok {
		// we found a plugin for this connection name - check the pid
		if plugin.reattach.Pid == req.Pid {
			plugin.refCount--
			if plugin.refCount == 0 {
				// terminate the plugin
				err := m.killPlugin(plugin.reattach.Pid)
				if err != nil {
					return false, err
				}
				// remove from the map
				delete(m.Plugins, req.Connection)
			} else {
				m.Plugins[req.Connection] = plugin
			}
			// we found the plugin
			return true, nil
		}
	}
	// we did not find the plugin
	return false, nil
}

func (m *PluginManager) killPlugin(pid int64) error {
	process, err := os.FindProcess(int(pid))
	if err != nil {
		return err
	}
	return process.Kill()

}
