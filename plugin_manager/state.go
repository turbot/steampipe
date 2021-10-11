package plugin_manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
)

type pluginManagerState struct {
	Protocol        plugin.Protocol
	ProtocolVersion int
	Addr            *SimpleAddr
	Pid             int
}

func newPluginManagerState(reattach *plugin.ReattachConfig) *pluginManagerState {
	return &pluginManagerState{
		Protocol:        reattach.Protocol,
		ProtocolVersion: reattach.ProtocolVersion,
		Addr:            NewSimpleAddr(reattach.Addr),
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

func (s *pluginManagerState) save() error {
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

func (s *pluginManagerState) delete() error {
	return os.Remove(constants.PluginManagerStateFilePath())
}

func loadReattachConfig(verify bool) (*plugin.ReattachConfig, error) {
	if !helpers.FileExists(constants.PluginManagerStateFilePath()) {
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
			return nil, err
		} else if !running {
			return nil, nil
		}

	}
	return s.reattachConfig(), nil
}
