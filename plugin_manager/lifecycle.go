package plugin_manager

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"

	"github.com/hashicorp/go-plugin"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

// Start instantiates the plugin manager and saves the state
func Start() error {
	// try to load the plugin manager state
	state, err := loadPluginManagerState(true)

	if err != nil {
		return err
	}
	if state != nil {
		log.Printf("[TRACE] plugin manager Start() found previous instance of plugin manager still running - stopping it")

		if err := stop(state); err != nil {
			log.Printf("[WARN] failed to stop previous instance of plugin manager: %s", err)
			return err
		}
	}

	// we want to see the plugin manaer log
	log.SetOutput(os.Stdout)

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

	state = newPluginManagerState(client.ReattachConfig())

	// now save the state
	return state.save()
}

func Stop() error {
	// try to load the plugin manager state
	state, err := loadPluginManagerState(true)
	if err != nil {
		return err
	}
	if state == nil {
		// nothing to do
		return nil
	}
	return stop(state)
}

func stop(state *pluginManagerState) error {
	pluginManager, err := attachToPluginManager(state)
	if err != nil {
		return err
	}

	// tell plugin manager to kill all plugins
	_, err = pluginManager.Shutdown(&pb.ShutdownRequest{})
	if err != nil {
		return err
	}
	// now kill the plugin manager
	return state.kill()

	return err

}

// GetPluginManager connects to a running plugin manager
func GetPluginManager() (pluginshared.PluginManager, error) {
	return getPluginManager(true)
}

func getPluginManager(startIfNeeded bool) (pluginshared.PluginManager, error) {
	// try to load the plugin manager state
	state, err := loadPluginManagerState(true)
	if err != nil {
		log.Printf("[TRACE] failed to load plugin manager state: %s", err.Error())
		return nil, err
	}
	// if we did not load it and there was no error, it means the plugin manager is not running
	if state == nil {
		log.Printf("[TRACE] GetPluginManager called but plugin manager not running - calling Start()")
		if startIfNeeded {
			// start the plugin manager
			if err := Start(); err != nil {
				return nil, err
			}
			// recurse in, setting startIfNeeded to false to avoid further recursion on failure
			return getPluginManager(false)
		}
		// not retrying - just fail
		return nil, fmt.Errorf("plugin manager is not running")
	}

	pluginManager, err := attachToPluginManager(state)
	if err != nil {
		return nil, err
	}

	return pluginManager, nil
}

func attachToPluginManager(state *pluginManagerState) (pluginshared.PluginManager, error) {
	// construct a client using this reaattach config
	newClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.PluginMap,
		Reattach:        state.reattachConfig(),
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
