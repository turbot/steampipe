package steampipeconfig

import (
	"fmt"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pluginmanager"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

type ConnectionDataMap map[string]*ConnectionData

func (m ConnectionDataMap) Equals(other ConnectionDataMap) bool {
	if m != nil && other == nil {
		return false
	}
	for k, lVal := range m {
		rVal, ok := other[k]
		if !ok || !lVal.Equals(rVal) {
			return false
		}
	}
	for k := range other {
		if _, ok := m[k]; !ok {
			return false
		}
	}
	return true
}

func (m ConnectionDataMap) Connections() []*modconfig.Connection {
	var res = make([]*modconfig.Connection, len(m))
	idx := 0
	for _, d := range m {
		res[idx] = d.Connection
		idx++
	}
	return res
}

// NewConnectionDataMap populates a map of connection data for all connections in connectionMap
func NewConnectionDataMap(connectionMap map[string]*modconfig.Connection) (ConnectionDataMap, []string, error) {
	utils.LogTime("steampipeconfig.getRequiredConnections start")
	defer utils.LogTime("steampipeconfig.getRequiredConnections end")

	requiredConnections := ConnectionDataMap{}
	var missingPlugins []string

	// cache plugin file creation times in a dictionary to avoid reloading the same plugin file multiple times
	modTimeMap := make(map[string]time.Time)

	utils.LogTime("steampipeconfig.getRequiredConnections config-iteration start")
	// populate checksum for each referenced plugin
	for name, connection := range connectionMap {
		remoteSchema := connection.Plugin
		pluginPath, err := pluginmanager.GetPluginPath(connection.Plugin, connection.PluginShortName)
		if err != nil {
			err := fmt.Errorf("failed to load connection '%s': %v\n%s", connection.Name, err, connection.DeclRange)
			return nil, nil, err
		}
		// if plugin is not installed, the path will be returned as empty
		if pluginPath == "" {
			if !helpers.StringSliceContains(missingPlugins, connection.Plugin) {
				missingPlugins = append(missingPlugins, connection.Plugin)
			}
			continue
		}

		// get the plugin file mod time
		var modTime time.Time
		var ok bool
		if modTime, ok = modTimeMap[pluginPath]; !ok {
			modTime, err = utils.FileModTime(pluginPath)
			if err != nil {
				return nil, nil, err
			}
			modTimeMap[pluginPath] = modTime
		}

		requiredConnections[name] = NewConnectionData(remoteSchema, connection, modTime)
	}
	utils.LogTime("steampipeconfig.getRequiredConnections config-iteration end")

	return requiredConnections, missingPlugins, nil
}
