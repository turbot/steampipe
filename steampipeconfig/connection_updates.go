package steampipeconfig

import (
	"fmt"
	"log"

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

// ConnectionData is a struct containing all details for a connection
// - the plugin name and checksum, the connection config and options
// json tags needed as this is stored in the connection state file
type ConnectionData struct {
	// the fully qualified name of the plugin
	Plugin string
	// the checksum of the plugin file
	CheckSum string
	// the underlying connection object
	Connection *modconfig.Connection
	// the hash of the connection schema
	SchemaHash string `json:"SchemaHash,omitempty"`
}

func (p *ConnectionData) Equals(other *ConnectionData) bool {
	if p.Connection == nil || other.Connection == nil {
		// if either object has a nil Connection, then it may be data from an old connection state file
		// return false, so that connections get refreshed and this file gets written in the new format in the process
		return false
	}

	return p.Plugin == other.Plugin &&
		p.CheckSum == other.CheckSum &&
		p.Connection.Equals(other.Connection)
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

func (m ConnectionDataMap) ConnectionsWithDynamicSchema() ConnectionDataMap {
	var res = make(ConnectionDataMap)
	for name, c := range m {
		if c.Connection.Options != nil && c.Connection.Options.DynamicSchema != nil && *c.Connection.Options.DynamicSchema {
			res[name] = c
		}
	}
	return res
}

// GetConnectionsToUpdate returns updates to be made to the database to sync with connection config
func GetConnectionsToUpdate(schemas []string, requiredConnections ConnectionDataMap, missingPlugins []string, dynamicSchemaHashMap map[string]string) (*ConnectionUpdates, error) {
	utils.LogTime("steampipeconfig.GetConnectionsToUpdate start")
	defer utils.LogTime("steampipeconfig.GetConnectionsToUpdate end")

	// load the connection state file and filter out any connections which are not in the list of schemas
	// this allows for the database being rebuilt,modified externally
	connectionState, err := GetConnectionState(schemas)
	if err != nil {
		return nil, err
	}

	result := newConnectionUpdates()
	result.MissingPlugins = missingPlugins
	result.RequiredConnections = requiredConnections

	// connections to create/update
	for name, requiredConnectionData := range requiredConnections {
		// check whether this connection exists in the state
		currentConnectionData, ok := connectionState[name]
		// if it does not exist, or is not equal, add to updates
		if !ok || !currentConnectionData.Equals(requiredConnectionData) {
			log.Printf("[TRACE] connection %s is out of date or missing\n", name)
			result.Update[name] = requiredConnectionData
		}
	}

	// connections to delete - any connection which is in connection state but NOT required connections
	for connection, requiredPlugin := range connectionState {
		if _, ok := requiredConnections[connection]; !ok {
			log.Printf("[TRACE] connection %s is no longer required\n", connection)
			result.Delete[connection] = requiredPlugin
		}
	}

	// now for every connection with dynamic schema, check whether the schema we have just fetched
	// matches the existing db schema
	for name, requiredHash := range dynamicSchemaHashMap {
		// get the connection data from the loaded connection state
		connectionData, ok := connectionState[name]
		// if the connection exists in the state, does the schemas hash match?
		if ok && connectionData.SchemaHash != requiredHash {
			result.Update[name] = connectionData
		}
	}

	return result, nil
}

// GetRequiredConnectionData loads and parses the connection config
func GetRequiredConnectionData(connectionConfig map[string]*modconfig.Connection) (ConnectionDataMap, []string, error) {
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
