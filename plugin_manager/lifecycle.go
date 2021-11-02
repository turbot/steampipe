package plugin_manager

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"syscall"

	"github.com/hashicorp/go-plugin"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"

	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

// Start loads the plugin manager state, stops any previous instance and instantiates a new the plugin manager
func Start() error {
	// try to load the plugin manager state
	state, err := loadPluginManagerState(true)

	if err != nil {
		return err
	}
	if state != nil {
		log.Printf("[WARN] ******************** plugin manager Start() found previous instance of plugin manager still running - stopping it")
		// stop the current instance
		if err := stop(state); err != nil {
			log.Printf("[WARN] ******************** failed to stop previous instance of plugin manager: %s", err)
			return err
		}
	}
	return start()
}

// start plugin manager, without checking it is already running
func start() error {
	// We don't want to see the plugin logs.
	log.SetOutput(ioutil.Discard)

	// create command which will start plugin-manager
	// we have to spawn a separate process to do this so the plugin process itself is not an orphan
	// TODO more detail about this
	pluginManagerCmd := exec.Command("steampipe", "plugin-manager", "--spawn")
	// set attributes on the command to ensure the process is not shutdown when its parent terminates
	pluginManagerCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	// launch the plugin manager the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.PluginMap,
		Cmd:             pluginManagerCmd,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})
	if _, err := client.Start(); err != nil {
		return err
	}

	// create a plugin manager state
	state := NewPluginManagerState(client.ReattachConfig())

	// now Save the state
	return state.Save()
}

//Stop loads the plugin manager state and if a running instance is found, stop it
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

// stop the running plugin manager instance
func stop(state *pluginManagerState) error {
	pluginManager, err := NewPluginManagerClientWithRetries(state)
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
	log.Printf("[WARN] ******************** GetPluginManager")
	return getPluginManager(true)
}

// getPluginManager determines whether the plugin manager is running
// if not,and if startIfNeeded is true, it starts the manager
// it then returns a plugin manager client
func getPluginManager(startIfNeeded bool) (pluginshared.PluginManager, error) {
	// try to load the plugin manager state
	state, err := loadPluginManagerState(true)
	if err != nil {
		log.Printf("[TRACE] failed to load plugin manager state: %s", err.Error())
		return nil, err
	}
	// if we did not load it and there was no error, it means the plugin manager is not running
	if state == nil {
		log.Printf("[WARN] GetPluginManager called but plugin manager not running")
		if startIfNeeded {
			log.Printf("[WARN] calling Start()")
			// start the plugin manager
			if err := start(); err != nil {
				return nil, err
			}
			// recurse in, setting startIfNeeded to false to avoid further recursion on failure
			return getPluginManager(false)
		}
		// not retrying - just fail
		return nil, fmt.Errorf("plugin manager is not running")
	}
	return NewPluginManagerClientWithRetries(state)
}
