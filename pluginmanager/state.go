package pluginmanager

import (
	"encoding/json"
	"log"
	"os"
	"syscall"

	"github.com/hashicorp/go-plugin"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
	pb "github.com/turbot/steampipe/pluginmanager_service/grpc/proto"
)

const PluginManagerStructVersion = 20220411

type PluginManagerState struct {
	Protocol        plugin.Protocol `json:"protocol"`
	ProtocolVersion int             `json:"protocol_version"`
	Addr            *pb.SimpleAddr  `json:"addr"`
	Pid             int             `json:"pid"`
	// path to the steampipe executable
	Executable string `json:"executable"`
	// is the plugin manager running
	Running       bool  `json:"-"`
	StructVersion int64 `json:"struct_version"`
}

func NewPluginManagerState(executable string, reattach *plugin.ReattachConfig) *PluginManagerState {
	return &PluginManagerState{
		Executable:      executable,
		Protocol:        reattach.Protocol,
		ProtocolVersion: reattach.ProtocolVersion,
		Addr:            pb.NewSimpleAddr(reattach.Addr),
		Pid:             reattach.Pid,
		StructVersion:   PluginManagerStructVersion,
	}
}

func LoadPluginManagerState() (*PluginManagerState, error) {
	// always return empty state
	s := new(PluginManagerState)
	if !filehelpers.FileExists(filepaths.PluginManagerStateFilePath()) {
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

	// check is the manager is running - this deletes that state file if it is not running,
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

func (s *PluginManagerState) Save() error {
	// set struct version
	s.StructVersion = PluginManagerStructVersion

	content, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepaths.PluginManagerStateFilePath(), content, 0644)
}

func (s *PluginManagerState) reattachConfig() *plugin.ReattachConfig {
	return &plugin.ReattachConfig{
		Protocol:        s.Protocol,
		ProtocolVersion: s.ProtocolVersion,
		Addr:            *s.Addr,
		Pid:             s.Pid,
	}
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
