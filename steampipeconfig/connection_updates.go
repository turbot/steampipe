package steampipeconfig

import (
	"fmt"
	"log"
	"reflect"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

type ConnectionUpdates struct {
	Update         ConnectionDataMap
	Delete         ConnectionDataMap
	MissingPlugins []string
	// the connections which will exist after the update
	RequiredConnections ConnectionDataMap
}

func newConnectionUpdates() *ConnectionUpdates {
	return &ConnectionUpdates{
		Update:              ConnectionDataMap{},
		Delete:              ConnectionDataMap{},
		MissingPlugins:      []string{},
		RequiredConnections: ConnectionDataMap{},
	}
}

// ConnectionData is a struct containing all details for a connection - the plugin name and checksum, the connection config and options
type ConnectionData struct {
	// the fully qualified name of the plugin
	Plugin string
	// the checksum of the plugin file
	CheckSum string
	// the underlying connection object
	Connection *modconfig.Connection
}

func (p *ConnectionData) Equals(other *ConnectionData) bool {
	if p.Connection == nil || other.Connection == nil {
		// this is data from an old connection file.
		// return false, so that connections get refreshed
		// and this file gets written in the new format in the process
		return false
	}
	connectionOptionsEqual := (p.Connection.Options == nil) == (other.Connection.Options == nil)
	if p.Connection.Options != nil {
		connectionOptionsEqual = p.Connection.Options.Equals(other.Connection.Options)
	}

	return p.Plugin == other.Plugin &&
		p.CheckSum == other.CheckSum &&
		p.Connection.Name == other.Connection.Name &&
		connectionOptionsEqual &&
		reflect.DeepEqual(p.Connection.Config, other.Connection.Config)
}

type ConnectionDataMap map[string]*ConnectionData

func (m ConnectionDataMap) Equals(other ConnectionDataMap) bool {
	if m != nil && other == nil {
		return false
	}
	for k, lVal := range m {
		rVal, ok := other[k]
		if !ok {
			return false
		}
		if !lVal.Equals(rVal) {
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

// GetConnectionsToUpdate :: returns updates to be made to the database to sync with connection config
func GetConnectionsToUpdate(schemas []string, connectionConfig map[string]*modconfig.Connection) (*ConnectionUpdates, error) {
	utils.LogTime("steampipeconfig.GetConnectionsToUpdate start")
	defer utils.LogTime("steampipeconfig.GetConnectionsToUpdate end")

	// load the connection state file and filter out any connections which are not in the list of schemas
	// this allows for the database being rebuilt,modified externally
	connectionState, err := GetConnectionState(schemas)
	if err != nil {
		return nil, err
	}

	requiredConnections, missingPlugins, err := getRequiredConnections(connectionConfig)
	if err != nil {
		return nil, err
	}

	result := newConnectionUpdates()
	result.MissingPlugins = missingPlugins
	result.RequiredConnections = requiredConnections

	// connections to create/update
	for connection, requiredPlugin := range requiredConnections {
		current, ok := connectionState[connection]
		if !ok || !current.Equals(requiredPlugin) {
			log.Printf("[TRACE] connection %s is out of date or missing\n", connection)
			result.Update[connection] = requiredPlugin
		}
	}

	// connections to delete
	for connection, requiredPlugin := range connectionState {
		if _, ok := requiredConnections[connection]; !ok {
			log.Printf("[TRACE] connection %s is no longer required\n", connection)
			result.Delete[connection] = requiredPlugin
		}
	}
	return result, nil
}

// load and parse the connection config
func getRequiredConnections(connectionConfig map[string]*modconfig.Connection) (ConnectionDataMap, []string, error) {
	utils.LogTime("steampipeconfig.getRequiredConnections start")
	defer utils.LogTime("steampipeconfig.getRequiredConnections end")

	requiredConnections := ConnectionDataMap{}
	var missingPlugins []string

	utils.LogTime("steampipeconfig.getRequiredConnections config-iteration start")
	// populate checksum for each referenced plugin
	for name, connection := range connectionConfig {
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

		requiredConnections[name] = &ConnectionData{
			Plugin:     remoteSchema,
			CheckSum:   checksum,
			Connection: connection,
		}
	}
	utils.LogTime("steampipeconfig.getRequiredConnections config-iteration end")

	return requiredConnections, missingPlugins, nil
}
