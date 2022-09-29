package steampipeconfig

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/migrate"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pluginmanager"
)

type ConnectionDataMap map[string]*ConnectionData

// IsValid checks whether the struct was correctly deserialized,
// by checking if the ConnectionData StructVersion is populated
func (s *ConnectionDataMap) IsValid() bool {
	for _, v := range *s {
		if !v.IsValid() {
			return false
		}
	}
	return true
}

func (s *ConnectionDataMap) MigrateFrom() migrate.Migrateable {
	for _, v := range *s {
		v.MigrateLegacy()
	}
	return s
}

func (f *ConnectionDataMap) Save() error {
	connFilePath := filepaths.ConnectionStatePath()
	for _, v := range *f {
		v.MaintainLegacy()
	}
	connFileJSON, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		log.Println("[ERROR]", "Error while writing state file", err)
		return err
	}
	return os.WriteFile(connFilePath, connFileJSON, 0644)
}

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
func NewConnectionDataMap(connectionMap map[string]*modconfig.Connection) (ConnectionDataMap, map[string][]modconfig.Connection, error) {
	utils.LogTime("steampipeconfig.getRequiredConnections start")
	defer utils.LogTime("steampipeconfig.getRequiredConnections end")

	requiredConnections := ConnectionDataMap{}

	// cache plugin file creation times in a dictionary to avoid reloading the same plugin file multiple times
	modTimeMap := make(map[string]time.Time)

	// map ofd missing polugins, keyed by plugin, value is list of conections using missing plugin
	missingPluginMap := make(map[string][]modconfig.Connection)

	utils.LogTime("steampipeconfig.getRequiredConnections config-iteration start")
	// populate file mod time for each referenced plugin
	for name, connection := range connectionMap {
		remoteSchema := connection.Plugin
		pluginPath, _ := pluginmanager.GetPluginPath(connection.Plugin, connection.PluginShortName)
		// ignore error if plugin is not available
		// if plugin is not installed, the path will be returned as empty
		if pluginPath == "" {
			missingPluginMap[connection.Plugin] = append(missingPluginMap[connection.Plugin], *connection)

			// if !helpers.StringSliceContains(missingPlugins, connection.Plugin) {
			// 	missingPlugins = append(missingPlugins, connection.Plugin)
			// }
			continue
		}

		// get the plugin file mod time
		var modTime time.Time
		var ok bool
		if modTime, ok = modTimeMap[pluginPath]; !ok {
			modTime, err := utils.FileModTime(pluginPath)
			if err != nil {
				return nil, nil, err
			}
			modTimeMap[pluginPath] = modTime
		}

		requiredConnections[name] = NewConnectionData(remoteSchema, connection, modTime)
	}
	utils.LogTime("steampipeconfig.getRequiredConnections config-iteration end")

	return requiredConnections, missingPluginMap, nil
}
