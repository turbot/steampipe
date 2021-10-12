package plugin_manager

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

// Start instantiates the plugin manager and saves the reattach config
// TODO handle if service is already running
func Start() error {
	// try to load the reattach config
	reattach, err := loadReattachConfig(true)
	if err != nil {

		return err
	}
	if reattach != nil {
		// TODO should we kill it?
		return nil
	}

	// We don't want to see the plugin logs.
	//log.SetOutput(ioutil.Discard)

	// launch the plugin manager the plugin process.
	// TODO pass config path or set connection config
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.PluginMap,
		Cmd:             exec.Command("sh", "-c", "steampipe plugin-manager"),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})
	if _, err := client.Start(); err != nil {
		return err
	}

	state := newPluginManagerState(client.ReattachConfig())

	// now save the reattach config
	return state.save()
}

// GetPluginManager connects to a running plugin manager
func GetPluginManager() (pluginshared.PluginManager, error) {
	// try to load the reattach config
	reattach, err := loadReattachConfig(true)
	if err != nil {
		log.Printf("[TRACE] failed to load plugin manager reattach config: %s", err.Error())
		return nil, err
	}
	// if we did not load it and there was no error, it means the plugin manager is not running
	if reattach == nil {
		return nil, fmt.Errorf("plugin manager is not running")
	}

	// construct a client using this reaattach config
	newClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.PluginMap,
		Reattach:        reattach,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})

	// connect via RPC
	rpcClient, err := newClient.Client()
	if err != nil {
		log.Printf("[TRACE] failed to connect to plugin manager: %s", err.Error())
		return nil, err
	}

	// request the plugin
	raw, err := rpcClient.Dispense(pluginshared.PluginName)
	if err != nil {
		log.Printf("[TRACE] failed to retreive to plugin manager from running plugin process: %s", err.Error())
		return nil, err
	}

	// cast to correct type
	pluginManager := raw.(pluginshared.PluginManager)

	return pluginManager, nil
}
