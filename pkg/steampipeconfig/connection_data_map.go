package steampipeconfig

import (
	"encoding/json"
	"golang.org/x/exp/maps"
	"log"
	"os"
	"time"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

type ConnectionStateSummary map[string]int

type ConnectionDataMap map[string]*ConnectionData

// NewConnectionDataMap populates a map of connection data for all connections in connectionMap
func NewConnectionDataMap(connectionMap map[string]*modconfig.Connection, currentConnectionState ConnectionDataMap) (ConnectionDataMap, map[string][]modconfig.Connection, error) {
	utils.LogTime("steampipeconfig.getRequiredConnections start")
	defer utils.LogTime("steampipeconfig.getRequiredConnections end")

	res := ConnectionDataMap{}

	// cache plugin file creation times in a dictionary to avoid reloading the same plugin file multiple times
	pluginModTimeMap := make(map[string]time.Time)

	// map of missing plugins, keyed by plugin, value is list of conections using missing plugin
	missingPluginMap := make(map[string][]modconfig.Connection)

	utils.LogTime("steampipeconfig.getRequiredConnections config - iteration start")
	// populate file mod time for each referenced plugin
	for name, connection := range connectionMap {
		remoteSchema := connection.Plugin
		pluginPath, _ := filepaths.GetPluginPath(connection.Plugin, connection.PluginShortName)
		// ignore error if plugin is not available
		// if plugin is not installed, the path will be returned as empty
		if pluginPath == "" {
			missingPluginMap[connection.Plugin] = append(missingPluginMap[connection.Plugin], *connection)
			continue
		}

		// get the plugin file mod time
		var pluginModTime time.Time
		var ok bool
		if pluginModTime, ok = pluginModTimeMap[pluginPath]; !ok {
			var err error
			pluginModTime, err = utils.FileModTime(pluginPath)
			if err != nil {
				return nil, nil, err
			}
		}
		pluginModTimeMap[pluginPath] = pluginModTime
		res[name] = NewConnectionData(remoteSchema, connection, pluginModTime)

		// NOTE: if the connection exists in the current state, copy the connection mod time
		// (this will be updated to 'now' later if we are updating the connection)
		if currentState, ok := currentConnectionState[name]; ok {
			res[name].ConnectionModTime = currentState.ConnectionModTime
		}
	}
	utils.LogTime("steampipeconfig.getRequiredConnections config - iteration end")

	return res, missingPluginMap, nil
}

func (m ConnectionDataMap) GetSummary() ConnectionStateSummary {
	res := make(map[string]int, len(m))
	for _, c := range m {
		res[c.State]++
	}
	return res
}

// Pending returns whether there are any connections in the map which are pending
// this indicates that the db has just started and RefreshConnections has not been called yet
func (m ConnectionDataMap) Pending() bool {
	return m.ConnectionsInState(constants.ConnectionStatePending)
}

// Loaded returns whether loading is complete, i.e.  all connections are either ready or error
// (optionally, a list of connections may be passed, in which case just these connections are checked)
func (m ConnectionDataMap) Loaded(connections ...string) bool {
	// if no connections were passed, check them all
	if len(connections) == 0 {
		connections = maps.Keys(m)
	}

	for _, connectionName := range connections {
		connectionState, ok := m[connectionName]
		if !ok {
			// ignore if we have no state loaded for this conneciton name
			continue
		}
		if !connectionState.Loaded() {
			return false
		}
	}
	return true
}

// ConnectionsInState returns whether there are any connections one of the given states
func (m ConnectionDataMap) ConnectionsInState(states ...string) bool {
	for _, c := range m {
		for _, state := range states {
			if c.State == state {
				return true
			}
		}
	}
	return false
}

func (m ConnectionDataMap) Save() error {
	connFilePath := filepaths.ConnectionStatePath()
	connFileJSON, err := json.MarshalIndent(m, "", "  ")
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

// ConnectionModTime returns the latest connection mod time
func (m ConnectionDataMap) ConnectionModTime() time.Time {
	var res time.Time
	for _, c := range m {
		if c.ConnectionModTime.After(res) {
			res = c.ConnectionModTime
		}
	}
	return res
}
