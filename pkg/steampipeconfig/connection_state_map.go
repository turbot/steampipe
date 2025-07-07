package steampipeconfig

import (
	"encoding/json"
	"log"
	"os"
	"time"

	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/utils"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	"golang.org/x/exp/maps"
)

type ConnectionStateSummary map[string]int

type ConnectionStateMap map[string]*ConnectionState

// GetRequiredConnectionStateMap populates a map of connection data for all connections in connectionMap
func GetRequiredConnectionStateMap(connectionMap map[string]*modconfig.SteampipeConnection, currentConnectionState ConnectionStateMap) (ConnectionStateMap, map[string][]modconfig.SteampipeConnection, error_helpers.ErrorAndWarnings) {
	utils.LogTime("steampipeconfig.GetRequiredConnectionStateMap start")
	defer utils.LogTime("steampipeconfig.GetRequiredConnectionStateMap end")

	var res = error_helpers.ErrorAndWarnings{}
	requiredState := ConnectionStateMap{}

	// cache plugin file creation times in a dictionary to avoid reloading the same plugin file multiple times
	pluginModTimeMap := make(map[string]time.Time)

	// map of missing plugins, keyed by plugin alias, value is list of connections using missing plugin
	missingPluginMap := make(map[string][]modconfig.SteampipeConnection)

	utils.LogTime("steampipeconfig.getRequiredConnections config - iteration start")
	// populate file mod time for each referenced plugin
	for name, connection := range connectionMap {
		// if the connection is in error, create an error connection state
		// this may have been set by the loading code
		if connection.Error != nil {
			// add error connection state
			requiredState[connection.Name] = newErrorConnectionState(connection)
			// if error is a missing plugin, add to missingPluginMap
			// this will be used to build missing plugin warnings
			if connection.Error.Error() == pconstants.ConnectionErrorPluginNotInstalled {
				missingPluginMap[connection.PluginAlias] = append(missingPluginMap[connection.PluginAlias], *connection)
			} else {
				// otherwise add error to result as warning, so we display it
				res.AddWarning(connection.Error.Error())
			}
			continue
		}

		// to get here, PluginPath must be set
		pluginPath := *connection.PluginPath

		// get the plugin file mod time
		var pluginModTime time.Time
		var ok bool
		if pluginModTime, ok = pluginModTimeMap[pluginPath]; !ok {
			var err error
			pluginModTime, err = utils.FileModTime(pluginPath)
			if err != nil {
				res.Error = err
				return nil, nil, res
			}
		}
		pluginModTimeMap[pluginPath] = pluginModTime
		requiredState[name] = NewConnectionState(connection, pluginModTime)
		// the comments _will_ eventually be set
		requiredState[name].CommentsSet = true
		// if schema import is disabled, set desired state as disabled
		if connection.ImportSchema == modconfig.ImportSchemaDisabled {
			requiredState[name].State = constants.ConnectionStateDisabled
		}
		// NOTE: if the connection exists in the current state, copy the connection mod time
		// (this will be updated to 'now' later if we are updating the connection)
		if currentState, ok := currentConnectionState[name]; ok {
			requiredState[name].ConnectionModTime = currentState.ConnectionModTime
		}
	}

	return requiredState, missingPluginMap, res
}

func newErrorConnectionState(connection *modconfig.SteampipeConnection) *ConnectionState {
	res := NewConnectionState(connection, time.Now())
	res.SetError(connection.Error.Error())
	return res
}

func (m ConnectionStateMap) GetSummary() ConnectionStateSummary {
	res := make(map[string]int, len(m))
	for _, c := range m {
		res[c.State]++
	}
	return res
}

// Pending returns whether there are any connections in the map which are pending
// this indicates that the db has just started and RefreshConnections has not been called yet
func (m ConnectionStateMap) Pending() bool {
	return m.ConnectionsInState(constants.ConnectionStatePending, constants.ConnectionStatePendingIncomplete)
}

// Loaded returns whether loading is complete, i.e.  all connections are either ready or error
// (optionally, a list of connections may be passed, in which case just these connections are checked)
func (m ConnectionStateMap) Loaded(connections ...string) bool {
	// if no connections were passed, check them all
	if len(connections) == 0 {
		connections = maps.Keys(m)
	}

	for _, connectionName := range connections {
		connectionState, ok := m[connectionName]
		if !ok {
			// ignore if we have no state loaded for this connection name
			continue
		}
		log.Println("[TRACE] Checking state for", connectionName)
		if !connectionState.Loaded() {
			return false
		}
	}
	return true
}

// ConnectionsInState returns whether there are any connections one of the given states
func (m ConnectionStateMap) ConnectionsInState(states ...string) bool {
	for _, c := range m {
		for _, state := range states {
			if c.State == state {
				return true
			}
		}
	}
	return false
}

func (m ConnectionStateMap) Save() error {
	connFilePath := filepaths.ConnectionStatePath()
	connFileJSON, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Println("[ERROR]", "Error while writing state file", err)
		return err
	}
	return os.WriteFile(connFilePath, connFileJSON, 0644)
}

func (m ConnectionStateMap) Equals(other ConnectionStateMap) bool {
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

// ConnectionModTime returns the latest connection mod time
func (m ConnectionStateMap) ConnectionModTime() time.Time {
	var res time.Time
	for _, c := range m {
		if c.ConnectionModTime.After(res) {
			res = c.ConnectionModTime
		}
	}
	return res
}

func (m ConnectionStateMap) GetFirstSearchPathConnectionForPlugins(searchPath []string) []string {
	// build map of the connections which we must wait for:
	// for static plugins, just the first connection in the search path
	// for dynamic schemas all schemas in the search paths (as we do not know which schema may provide a given table)
	requiredSchemasMap := m.getFirstSearchPathConnectionMapForPlugins(searchPath)
	// convert this into a list
	var requiredSchemas []string
	for _, connections := range requiredSchemasMap {
		requiredSchemas = append(requiredSchemas, connections...)
	}
	return requiredSchemas
}

func (m ConnectionStateMap) GetPluginToConnectionMap() map[string][]string {
	res := make(map[string][]string)
	for connectionName, connectionState := range m {
		res[connectionState.Plugin] = append(res[connectionState.Plugin], connectionName)
	}
	return res
}

// getFirstSearchPathConnectionMapForPlugins builds map of plugin to the connections which must be loaded to ensure we can resolve unqualified queries
// for static plugins, just the first connection in the search path is included
// for dynamic schemas all search paths are included
func (m ConnectionStateMap) getFirstSearchPathConnectionMapForPlugins(searchPath []string) map[string][]string {
	res := make(map[string][]string)
	for _, connectionName := range searchPath {
		// is this in the connection state map
		connectionState, ok := m[connectionName]
		if !ok {
			continue
		}
		// if this connection is disabled, skip it
		if connectionState.Disabled() {
			continue
		}

		// get the plugin
		plugin := connectionState.Plugin
		// if this is the first connection for this plugin, or this is a dynamic plugin, add to the result map
		if len(res[plugin]) == 0 || connectionState.SchemaMode == sdkplugin.SchemaModeDynamic {
			res[plugin] = append(res[plugin], connectionName)
		}
	}
	return res
}

func (m ConnectionStateMap) SetConnectionsToPendingOrIncomplete() {
	for _, state := range m {
		if state.State == constants.ConnectionStateReady {
			state.State = constants.ConnectionStatePending
			state.ConnectionModTime = time.Now()
		} else if state.State != constants.ConnectionStateDisabled {
			state.State = constants.ConnectionStatePendingIncomplete
			state.ConnectionModTime = time.Now()
		}
	}
}

// PopulateFilename sets the Filename, StartLineNumber and EndLineNumber properties
// this is required as these fields were added to the table after release
func (m ConnectionStateMap) PopulateFilename() {
	// get the connection from config
	connections := GlobalConfig.Connections
	for name, state := range m {
		// do we have config for this connection (
		if connection := connections[name]; connection != nil {
			state.setFilename(connection)
		}
	}
}
