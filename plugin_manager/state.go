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
}

func NewPluginManagerState(reattach *plugin.ReattachConfig) *pluginManagerState {
	return &pluginManagerState{
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

func (s *pluginManagerState) verifyServiceRunning() (bool, error) {
	pidExists, err := utils.PidExists(s.Pid)
	if err != nil {
		return false, fmt.Errorf("failed to verify plugin manager is running: %s", err.Error())
	}
	if !pidExists {
		// file is outdated - delete
		if err := s.delete(); err != nil {
			return false, err
		}
		// plugin manager is NOT running
		return false, nil
	}
	// plugin manager IS running
	return true, nil
}

// kill the plugin manager process and delete the state
func (s *pluginManagerState) kill() error {
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

func loadPluginManagerState(verify bool) (*pluginManagerState, error) {
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

	if verify {
		if running, err := s.verifyServiceRunning(); err != nil {
			log.Printf("[WARN] plugin manager is running, pid %d", s.Pid)
			return nil, err
		} else if !running {
			log.Printf("[WARN] plugin manager state file exists but pid %d is not running - deleting file", s.Pid)
			return nil, nil
		}

	}
	return s, nil
}
