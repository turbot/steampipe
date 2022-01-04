package pluginmanager

import (
	"encoding/json"
	"log"
	"os"
	"syscall"

	"github.com/turbot/steampipe/filepaths"

	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	pb "github.com/turbot/steampipe/pluginmanager/grpc/proto"
	"github.com/turbot/steampipe/utils"
)

type PluginManagerState struct {
	Protocol        plugin.Protocol
	ProtocolVersion int
	Addr            *pb.SimpleAddr
	Pid             int
	// path to the steampipe executable
	Executable string
	// is the plugin manager running
	Running bool `json:"-"`
}

func NewPluginManagerState(executable string, reattach *plugin.ReattachConfig) *PluginManagerState {
	return &PluginManagerState{
		Executable:      executable,
		Protocol:        reattach.Protocol,
		ProtocolVersion: reattach.ProtocolVersion,
		Addr:            pb.NewSimpleAddr(reattach.Addr),
		Pid:             reattach.Pid,
	}
}

func (s *PluginManagerState) reattachConfig() *plugin.ReattachConfig {
	return &plugin.ReattachConfig{
		Protocol:        s.Protocol,
		ProtocolVersion: s.ProtocolVersion,
		Addr:            *s.Addr,
		Pid:             s.Pid,
	}
}

func (s *PluginManagerState) Save() error {
	content, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepaths.PluginManagerStateFilePath(), content, 0644)
}

// check whether the plugin manager is running
func (s *PluginManagerState) verifyRunning() (bool, error) {
	pidExists, err := utils.PidExists(s.Pid)
	if err != nil {
		return false, err
	}
	return pidExists, nil
}

// kill the plugin manager process and delete the state
func (s *PluginManagerState) kill() error {
	// the state file contains the Pid of the daemon process - find and kill the process
	process, err := utils.FindProcess(s.Pid)
	if err != nil {
		return err
	}
	if process == nil {
		log.Println("[TRACE] tried to kill plugin_manager, but couldn't find process")
		return nil
	}
	// kill the plugin manager process by sending a SIGTERM (to give it a chance to clean up its children)
	err = process.SendSignal(syscall.SIGTERM)
	if err != nil {
		log.Println("[TRACE] tried to kill plugin_manager, but couldn't send signal to process", err)
		return err
	}
	// delete the state file as we have shutdown the plugin manager
	s.delete()
	return nil
}

func (s *PluginManagerState) delete() {
	os.Remove(filepaths.PluginManagerStateFilePath())
}

func LoadPluginManagerState() (*PluginManagerState, error) {
	// always return empty state
	s := new(PluginManagerState)
	if !helpers.FileExists(filepaths.PluginManagerStateFilePath()) {
		log.Printf("[TRACE] plugin manager state file not found")
		return s, nil
	}

	fileContent, err := os.ReadFile(filepaths.PluginManagerStateFilePath())
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(fileContent, s)
	if err != nil {
		log.Printf("[TRACE] failed to unmarshall plugin manager state file at %s with error %s\n", filepaths.PluginManagerStateFilePath(), err.Error())
		log.Printf("[TRACE] deleting invalid plugin manager state file\n")
		s.delete()
		return s, nil
	}

	// check is the manager is running - this deletes that state file if it si not running,
	// and set the 'Running' property on the state if it is
	pluginManagerRunning, err := s.verifyRunning()
	if err != nil {
		return s, err
	}
	// save the running status on the state struct
	s.Running = pluginManagerRunning

	// return error (which may be nil)
	return s, err
}
