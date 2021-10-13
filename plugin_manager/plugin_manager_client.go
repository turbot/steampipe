package plugin_manager

import (
	"log"

	"github.com/hashicorp/go-plugin"
	"github.com/turbot/steampipe/constants"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

const maxRetries = 2

// PluginManagerClientWithRetries is the client used by steampipe to access the plugin manager
// it implements retries on the grpc calls
type PluginManagerClientWithRetries struct {
	manager            pluginshared.PluginManager
	pluginManagerState *pluginManagerState
}

func NewPluginManagerClientWithRetries(pluginManagerState *pluginManagerState) (*PluginManagerClientWithRetries, error) {
	res := &PluginManagerClientWithRetries{
		pluginManagerState: pluginManagerState,
	}
	err := res.attachToPluginManager()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *PluginManagerClientWithRetries) attachToPluginManager() error {
	// construct a client using the plugin manager reaattach config
	newClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.PluginMap,
		Reattach:        c.pluginManagerState.reattachConfig(),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})

	// connect via RPC
	rpcClient, err := newClient.Client()
	if err != nil {
		log.Printf("[TRACE] failed to connect to plugin manager: %s", err.Error())
		return err
	}

	// request the plugin
	raw, err := rpcClient.Dispense(pluginshared.PluginName)
	if err != nil {
		log.Printf("[TRACE] failed to retreive to plugin manager from running plugin process: %s", err.Error())
		return err
	}

	// cast to correct type
	pluginManager := raw.(pluginshared.PluginManager)
	c.manager = pluginManager
	return nil
}

func (c *PluginManagerClientWithRetries) Get(req *pb.GetRequest) (res *pb.GetResponse, err error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		res, err = c.manager.Get(req)
		if !c.ShouldRetry(err) {
			break
		}
		// reattach to the plugin manager
		err = c.attachToPluginManager()
		if err != nil {
			return nil, err
		}
	}
	return res, err
}

func (c *PluginManagerClientWithRetries) SetConnectionConfigMap(req *pb.SetConnectionConfigMapRequest) (res *pb.SetConnectionConfigMapResponse, err error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		res, err = c.manager.SetConnectionConfigMap(req)
		if !c.ShouldRetry(err) {
			break
		}
	}
	return res, err
}

func (c *PluginManagerClientWithRetries) Shutdown(req *pb.ShutdownRequest) (res *pb.ShutdownResponse, err error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		res, err = c.manager.Shutdown(req)
		if !c.ShouldRetry(err) {
			break
		}
	}
	return res, err
}

func (c *PluginManagerClientWithRetries) ShouldRetry(err error) bool {
	return constants.IsGRPCConnectivityError(err)
}
