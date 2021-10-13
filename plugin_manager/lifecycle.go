package plugin_manager

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"

	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"

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
	pluginManagerCmd := exec.Command("steampipe", "start-plugin-manager")
	pluginManagerCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := pluginManagerCmd.Start()
	time.Sleep(5 * time.Second)
	return err

	//// we want to see the plugin manager log
	//log.SetOutput(os.Stdout)
	//
	//// create command which will run steampipe in plugin-manager mode
	//pluginManagerCmd := exec.Command("steampipe", "plugin-manager", "--spawn")
	//// set attributes on the command to ensure the process is not shutdown when its parent terminates
	//pluginManagerCmd.SysProcAttr = &syscall.SysProcAttr{
	//	Setpgid:    true,
	//	Foreground: false,
	//}
	//// launch the plugin manager the plugin process.
	//client := plugin.NewClient(&plugin.ClientConfig{
	//	HandshakeConfig: pluginshared.Handshake,
	//	Plugins:         pluginshared.PluginMap,
	//	Cmd:             pluginManagerCmd,
	//	AllowedProtocols: []plugin.Protocol{
	//		plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	//})
	//if _, err := client.Start(); err != nil {
	//	return err
	//}
	//
	//// create a plugin manager state
	//state := NewPluginManagerState(client.ReattachConfig())
	//
	//// now Save the state
	//return state.Save()
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

func getPluginManager(startIfNeeded bool) (pluginshared.PluginManager, error) {
	// try to load the plugin manager state
	state, err := loadPluginManagerState(true)
	if err != nil {
		log.Printf("[TRACE] failed to load plugin manager state: %s", err.Error())
		return nil, err
	}
	// if we did not load it and there was no error, it means the plugin manager is not running
	if state == nil {
		log.Printf("[WARN] GetPluginManager called but plugin manager not running - calling Start()")
		if startIfNeeded {
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
