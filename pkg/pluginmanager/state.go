package pluginmanager

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/hashicorp/go-plugin"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

const PluginManagerStructVersion = 20220411

// stateMutex protects concurrent writes to the state file
var stateMutex sync.Mutex

type State struct {
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

func NewState(executable string, reattach *plugin.ReattachConfig) *State {
	return &State{
		Executable:      executable,
		Protocol:        reattach.Protocol,
		ProtocolVersion: reattach.ProtocolVersion,
		Addr:            pb.NewSimpleAddr(reattach.Addr),
		Pid:             reattach.Pid,
		StructVersion:   PluginManagerStructVersion,
	}
}

func LoadState() (*State, error) {
	// always return empty state
	s := new(State)
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
		log.Printf("[TRACE] error verifying plugin manager running: %s", err)
		return s, err
	}

	// save the running status on the state struct
	s.Running = pluginManagerRunning

	// return error (which may be nil)
	return s, err
}

func (s *State) Save() error {
	// Protect concurrent writes with a mutex
	stateMutex.Lock()
	defer stateMutex.Unlock()

	// set struct version
	s.StructVersion = PluginManagerStructVersion

	content, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Use atomic write to prevent file corruption from concurrent writes
	// Write to a temporary file first, then atomically rename it
	stateFilePath := filepaths.PluginManagerStateFilePath()

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(stateFilePath), 0755); err != nil {
		return err
	}

	tempFile := stateFilePath + ".tmp"

	// Write to temporary file
	if err := os.WriteFile(tempFile, content, 0644); err != nil {
		return err
	}

	// Atomically rename the temp file to the final location
	// This ensures that the state file is never partially written
	return os.Rename(tempFile, stateFilePath)
}

func (s *State) reattachConfig() *plugin.ReattachConfig {
	// if Addr is nil, we cannot create a valid reattach config
	if s.Addr == nil {
		return nil
	}
	return &plugin.ReattachConfig{
		Protocol:        s.Protocol,
		ProtocolVersion: s.ProtocolVersion,
		Addr:            *s.Addr,
		Pid:             s.Pid,
	}
}

// check whether the plugin manager is running
func (s *State) verifyRunning() (bool, error) {
	log.Printf("[TRACE] verify plugin manager running, pid: %d", s.Pid)
	p, err := utils.FindProcess(s.Pid)
	if err != nil {
		log.Printf("[WARN] error finding process %d: %s", s.Pid, err)
		return false, err
	}
	if p == nil {
		log.Printf("[TRACE] process %d not found", s.Pid)
		return false, nil
	}

	// verify this is the correct process (and not a reused pid for a different process)
	exe, _ := p.Exe()
	cmd, _ := p.Cmdline()
	log.Printf("[TRACE] found process %d, checking if it is the plugin manager, exe: %s, cmd: %s, expected exe: %s", s.Pid, exe, cmd, s.Executable)
	// verify this is a plugin manager process by comparing the executable name and the command line
	return exe == s.Executable && strings.Contains(cmd, "plugin-manager"), nil
}

// kill the plugin manager process and delete the state
func (s *State) kill() (err error) {
	log.Printf("[TRACE] kill plugin manager, pid: %d", s.Pid)

	defer func() {
		// no error means the process is no longer running - delete the state file
		if err == nil {
			log.Printf("[TRACE] plugin manager process %d killed, deleting state file", s.Pid)
			s.delete()
		}
	}()
	// the state file contains the Pid of the daemon process - find and kill the process
	process, err := utils.FindProcess(s.Pid)
	if err != nil {
		return err
	}

	if process == nil {
		log.Printf("[TRACE] tried to kill plugin_manager, but couldn't find process (%d)", s.Pid)
		return nil
	}
	// kill the plugin manager process by sending a SIGTERM (to give it a chance to clean up its children)
	err = process.SendSignal(syscall.SIGTERM)
	if err != nil {
		log.Println("[WARN] tried to kill plugin_manager, but couldn't send signal to process", err)
		return err
	}

	return nil
}

func (s *State) delete() {
	_ = os.Remove(filepaths.PluginManagerStateFilePath())
}
