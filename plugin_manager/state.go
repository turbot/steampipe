package plugin_manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	"github.com/turbot/steampipe/utils"
)

type pluginManagerState struct {
	Protocol        plugin.Protocol
	ProtocolVersion int
	Addr            *pb.SimpleAddr
	Pid             int
	// path to the steampipe executable
	Executable string
	// is the plugin manager running
	Running bool
}

func NewPluginManagerState(executable string, reattach *plugin.ReattachConfig) *pluginManagerState {
	return &pluginManagerState{
		Executable:      executable,
		Protocol:        reattach.Protocol,
		ProtocolVersion: reattach.ProtocolVersion,
		Addr:            pb.NewSimpleAddr(reattach.Addr),
		Pid:             reattach.Pid,
	}
}

func (s *pluginManagerState) reattachConfig() *plugin.ReattachConfig {
	return &plugin.ReattachConfig{
		Protocol:        s.Protocol,
		ProtocolVersion: s.ProtocolVersion,
		Addr:            *s.Addr,
		Pid:             s.Pid,
	}
}

func (s *pluginManagerState) Save() error {
	content, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(constants.PluginManagerStateFilePath(), content, 0644)
}

// check whether the plugin manager is running
// it it is NOT, delete the state file
// if it is, set the 'running property of the statefile to true
func (s *pluginManagerState) verifyRunning() error {
	pidExists, err := utils.PidExists(s.Pid)

	// if we fail to determine if the plugin manager is running, assume it is NOT
	if err == nil && pidExists {
		s.Running = true
	} else if err = s.delete(); err != nil {
		// file is outdated - delete
		log.Printf("[WARN] plugin manager is not running but failed to delete state file: %s", err.Error())
		err = fmt.Errorf("plugin manager is not running but failed to delete state file: %s", err.Error())
	}
	// return error (which may be nil)
	return err
}

// kill the plugin manager process and delete the state
func (s *pluginManagerState) kill() error {
	// the state file contains the Pid of the daemon process - find and kill the process
	process, err := utils.FindProcess(s.Pid)
	if err != nil {
		return err
	}
	// kill the plugin manager process by sending a SIGTERM (to give it a chance to clean up its children)
	err = process.SendSignal(syscall.SIGTERM)
	if err != nil {
		return err
	}
	// delete the state file as we have shutdown the plugin manager
	return s.delete()
}

func (s *pluginManagerState) delete() error {
	return os.Remove(constants.PluginManagerStateFilePath())
}

func loadPluginManagerState() (*pluginManagerState, error) {
	if !helpers.FileExists(constants.PluginManagerStateFilePath()) {
		log.Printf("[TRACE] plugin manager state file not found")
		return nil, nil
	}

	fileContent, err := ioutil.ReadFile(constants.PluginManagerStateFilePath())
	if err != nil {
		return nil, err
	}
	var s = new(pluginManagerState)
	err = json.Unmarshal(fileContent, s)
	if err != nil {
		return nil, err
	}

	// check is the manager is running - this deletes that state file if it si not running,
	// and set the 'Running' property on the state if it is
	if err = s.verifyRunning(); err != nil {
		return nil, err
	}

	return s, nil
}
