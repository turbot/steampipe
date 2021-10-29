package plugin_manager

import (
	"io/ioutil"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

const maxRetries = 2

// PluginManagerClient is the client used by steampipe to access the plugin manager
type PluginManagerClient struct {
	manager            pluginshared.PluginManager
	pluginManagerState *pluginManagerState
}

func NewPluginManagerClient(pluginManagerState *pluginManagerState) (*PluginManagerClient, error) {
	res := &PluginManagerClient{
		pluginManagerState: pluginManagerState,
	}
	err := res.attachToPluginManager()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *PluginManagerClient) attachToPluginManager() error {
	// discard logging from the plugin client (plugin logs will still flow through)
	loggOpts := &hclog.LoggerOptions{Name: "plugin", Output: ioutil.Discard}
	logger := logging.NewLogger(loggOpts)

	// construct a client using the plugin manager reaattach config
	newClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.PluginMap,
		Reattach:        c.pluginManagerState.reattachConfig(),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		Logger: logger,
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

func (c *PluginManagerClient) Get(req *pb.GetRequest) (res *pb.GetResponse, err error) {
	return c.manager.Get(req)
}

func (c *PluginManagerClient) SetConnectionConfigMap(req *pb.SetConnectionConfigMapRequest) (res *pb.SetConnectionConfigMapResponse, err error) {
	return c.manager.SetConnectionConfigMap(req)
}

func (c *PluginManagerClient) Shutdown(req *pb.ShutdownRequest) (res *pb.ShutdownResponse, err error) {
	log.Printf("[WARN] PluginManagerClient Shutdown")
	return c.manager.Shutdown(req)
}
