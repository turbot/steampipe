package plugin_manager

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/constants"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

// Start loads the plugin manager state, stops any previous instance and instantiates a new the plugin manager
func Start() error {
	// try to load the plugin manager state
	state, err := loadPluginManagerState(true)
	if err != nil {
		log.Printf("[WARN] Startload state failed: %s", err)
		return err
	}

	if state != nil {
		log.Printf("[WARN] plugin manager Start() found previous instance of plugin manager still running - stopping it")
		// stop the current instance
		if err := stop(state); err != nil {
			log.Printf("[WARN] failed to stop previous instance of plugin manager: %s", err)
			return err
		}
	}
	return start()
}

// start plugin manager, without checking it is already running
func start() error {
	// create command which will start plugin-manager
	// we have to spawn a separate process to do this so the plugin process itself is not an orphan

	// get the location of the currently running steampipe process
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	log.Printf("[WARN] start got steampipe exe path: %s", executable)

	pluginManagerCmd := exec.Command(executable, "daemon", "--install-dir", viper.GetString(constants.ArgInstallDir))
	// set attributes on the command to ensure the process is not shutdown when its parent terminates
	pluginManagerCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// discard logging from the plugin manager client (plugin manager logs will still flow through to the log file
	// as this is set up in the pluginb manager)
	logger := logging.NewLogger(&hclog.LoggerOptions{Name: "plugin", Output: ioutil.Discard})

	// launch the plugin manager the plugin process
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.PluginMap,
		Cmd:             pluginManagerCmd,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		Logger: logger,
	})
	if _, err := client.Start(); err != nil {
		log.Printf("[WARN] failed to start GRPC client for plugin manager: %s", err)
		return err
	}

	log.Printf("[WARN] start: started plugin manager")
	// create a plugin manager state
	state := NewPluginManagerState(client.ReattachConfig())

	// now save the state
	return state.Save()
}

// Stop loads the plugin manager state and if a running instance is found, stop it
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
	log.Printf("[TRACE] plugin manager stop")
	pluginManager, err := NewPluginManagerClient(state)
	if err != nil {
		return err
	}

	log.Printf("[TRACE] pluginManager.Shutdown")
	// tell plugin manager to kill all plugins
	_, err = pluginManager.Shutdown(&pb.ShutdownRequest{})
	if err != nil {
		return err
	}

	log.Printf("[TRACE] pluginManager state.kill")
	// now kill the plugin manager
	return state.kill()

	return err
}

// GetPluginManager connects to a running plugin manager
func GetPluginManager() (pluginshared.PluginManager, error) {
	return getPluginManager(true)
}

// getPluginManager determines whether the plugin manager is running
// if not,and if startIfNeeded is true, it starts the manager
// it then returns a plugin manager client
func getPluginManager(startIfNeeded bool) (pluginshared.PluginManager, error) {
	log.Printf("[WARN] getPluginManager startIfNeeded %v", startIfNeeded)
	// try to load the plugin manager state
	state, err := loadPluginManagerState(true)
	if err != nil {
		log.Printf("[WARN] failed to load plugin manager state: %s", err.Error())
		return nil, err
	}
	// if we did not load it and there was no error, it means the plugin manager is not running
	if state == nil {
		log.Printf("[WARN] GetPluginManager called, plugin manager not running")
		if startIfNeeded {
			log.Printf("[WARN] calling Start()")
			// start the plugin manager
			if err := start(); err != nil {
				log.Printf("[WARN] start returned error %s", err.Error())
				return nil, err
			}
			// recurse in, setting startIfNeeded to false to avoid further recursion on failure
			return getPluginManager(false)
		}
		// not retrying - just fail
		return nil, fmt.Errorf("plugin manager is not running")
	}
	log.Printf("[WARN] plugin manager is running - returning client")
	return NewPluginManagerClient(state)
}
