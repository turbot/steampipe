package steampipeconfig

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
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

// NewConnectionDataMap tries to populate a map of connection data for all connections in connectionMap
func NewConnectionDataMap(connectionMap map[string]*modconfig.Connection) (ConnectionDataMap, []string, error) {
	utils.LogTime("steampipeconfig.getRequiredConnections start")
	defer utils.LogTime("steampipeconfig.getRequiredConnections end")

	requiredConnections := ConnectionDataMap{}
	var missingPlugins []string

	utils.LogTime("steampipeconfig.getRequiredConnections config-iteration start")
	// populate checksum for each referenced plugin
	for name, connection := range connectionMap {
		remoteSchema := connection.Plugin
		pluginPath, err := GetPluginPath(connection)
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

		checksum, err := utils.FileHash(pluginPath)
		if err != nil {
			return nil, nil, err
		}

		requiredConnections[name] = NewConnectionData(remoteSchema, checksum, connection)
	}
	utils.LogTime("steampipeconfig.getRequiredConnections config-iteration end")

	return requiredConnections, missingPlugins, nil
}
