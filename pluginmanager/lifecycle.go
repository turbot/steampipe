package pluginmanager

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/logging"
	"github.com/turbot/steampipe/pkg/filepaths"
	pb "github.com/turbot/steampipe/pluginmanager_service/grpc/proto"
	pluginshared "github.com/turbot/steampipe/pluginmanager_service/grpc/shared"
)

// StartNewInstance loads the plugin manager state, stops any previous instance and instantiates a new plugin manager
func StartNewInstance(steampipeExecutablePath string) error {
	// try to load the plugin manager state
	state, err := LoadPluginManagerState()
	if err != nil {
		log.Printf("[WARN] plugin manager StartNewInstance() - load state failed: %s", err)
		return err
	}

	if state.Running {
		log.Printf("[TRACE] plugin manager StartNewInstance() found previous instance of plugin manager still running - stopping it")
		// stop the current instance
		if err := stop(state); err != nil {
			log.Printf("[WARN] failed to stop previous instance of plugin manager: %s", err)
			return err
		}
	}
	return start(steampipeExecutablePath)
}

// start plugin manager, without checking it is already running
// we need to be provided with the exe path as we have no way of knowing where the steampipe exe it
// when ther plugin mananager is first started by steampipe, we derive the exe path from the running process and
// store it in the plugin manager state file - then if the fdw needs to start the plugin manager it knows how to
func start(steampipeExecutablePath string) error {
	// note: we assume the install dir has been assigned to file_paths.SteampipeDir
	// - this is done both by the FDW and Steampipe
	pluginManagerCmd := exec.Command(steampipeExecutablePath, "plugin-manager", "--install-dir", filepaths.SteampipeDir)
	// set attributes on the command to ensure the process is not shutdown when its parent terminates
	pluginManagerCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// discard logging from the plugin manager client (plugin manager logs will still flow through to the log file
	// as this is set up in the pluginb manager)
	logger := logging.NewLogger(&hclog.LoggerOptions{Name: "plugin", Output: io.Discard})

	// launch the plugin manager the plugin process
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  pluginshared.Handshake,
		Plugins:          pluginshared.PluginMap,
		Cmd:              pluginManagerCmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})
	if _, err := client.Start(); err != nil {
		log.Printf("[WARN] plugin manager start() failed to start GRPC client for plugin manager: %s", err)
		return err
	}

	// create a plugin manager state.
	state := NewPluginManagerState(steampipeExecutablePath, client.ReattachConfig())

	log.Printf("[TRACE] start: started plugin manager, pid %d", state.Pid)

	// now save the state
	return state.Save()
}

// Stop loads the plugin manager state and if a running instance is found, stop it
func Stop() error {
	// try to load the plugin manager state
	state, err := LoadPluginManagerState()
	if err != nil {
		return err
	}
	if state == nil || !state.Running {
		// nothing to do
		return nil
	}
	return stop(state)
}

// stop the running plugin manager instance
func stop(state *PluginManagerState) error {
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
}

// GetPluginManager connects to a running plugin manager
func GetPluginManager() (pluginshared.PluginManager, error) {
	return getPluginManager(true)
}

// getPluginManager determines whether the plugin manager is running
// if not,and if startIfNeeded is true, it starts the manager
// it then returns a plugin manager client
func getPluginManager(startIfNeeded bool) (pluginshared.PluginManager, error) {
	// try to load the plugin manager state
	state, err := LoadPluginManagerState()
	if err != nil {
		log.Printf("[WARN] failed to load plugin manager state: %s", err.Error())
		return nil, err
	}
	// if we did not load it and there was no error, it means the plugin manager is not running
	// we cannot start it as we do not know the correct steampipe exe path - which is stored in the state
	// this is not expected - we would expect the plugin manager to have been started with the datatbase
	if state == nil {
		return nil, fmt.Errorf("plugin manager is not running and there is no state file")
	}
	// if the plugin manager is not running, it must have crashed/terminated
	if !state.Running {
		log.Printf("[TRACE] GetPluginManager called but plugin manager not running")
		// is we are not already recursing, start the plugin manager then recurse back into this function
		if startIfNeeded {
			log.Printf("[TRACE] calling StartNewInstance()")
			// start the plugin manager
			if err := start(state.Executable); err != nil {
				return nil, err
			}
			// recurse in, setting startIfNeeded to false to avoid further recursion on failure
			return getPluginManager(false)
		}
		// not retrying - just fail
		return nil, fmt.Errorf("plugin manager is not running")
	}
	log.Printf("[TRACE] plugin manager is running - returning client")
	return NewPluginManagerClient(state)
}
