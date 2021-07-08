package steampipeconfig

import (
	"log"
	"reflect"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/options"
	"github.com/turbot/steampipe/utils"
)

type ConnectionUpdates struct {
	Update         ConnectionMap
	Delete         ConnectionMap
	MissingPlugins []string
	// the connections which will exist after the update
	RequiredConnections ConnectionMap
}

func newConnectionUpdates() *ConnectionUpdates {
	return &ConnectionUpdates{
		Update:              ConnectionMap{},
		Delete:              ConnectionMap{},
		MissingPlugins:      []string{},
		RequiredConnections: ConnectionMap{},
	}
}

// struct containing all details for a connection - the plugin name and checksum, the connection config and options
type ConnectionData struct {
	// the fully qualified name of the plugin
	Plugin string `yaml:"plugin"`
	// the checksum of the plugin file
	CheckSum string `yaml:"checkSum"`
	// connection name
	ConnectionName string
	// connection data (unparsed)
	ConnectionConfig string
	// steampipe connection options
	ConnectionOptions *options.Connection
}

func (p ConnectionData) Equals(other *ConnectionData) bool {
	connectionOptionsEqual := (p.ConnectionOptions == nil) == (other.ConnectionOptions == nil)
	if p.ConnectionOptions != nil {
		connectionOptionsEqual = p.ConnectionOptions.Equals(other.ConnectionOptions)
	}

	return p.Plugin == other.Plugin &&
		p.CheckSum == other.CheckSum &&
		p.ConnectionName == other.ConnectionName &&
		connectionOptionsEqual &&
		reflect.DeepEqual(p.ConnectionConfig, other.ConnectionConfig)
}

type ConnectionMap map[string]*ConnectionData

func (m ConnectionMap) Equals(other ConnectionMap) bool {
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
	// assume we will end up with the required connections
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
func getRequiredConnections(connectionConfig map[string]*modconfig.Connection) (ConnectionMap, []string, error) {
	utils.LogTime("steampipeconfig.getRequiredConnections start")
	defer utils.LogTime("steampipeconfig.getRequiredConnections end")

	requiredConnections := ConnectionMap{}
	var missingPlugins []string

	utils.LogTime("steampipeconfig.getRequiredConnections config-iteration start")
	// populate checksum for each referenced plugin
	for name, config := range connectionConfig {
		remoteSchema := config.Plugin
		pluginPath, err := GetPluginPath(remoteSchema)
		if err != nil {
			return nil, nil, err
		}
		// if plugin is not installed, the path will be returned as empty
		if pluginPath == "" {
			if !helpers.StringSliceContains(missingPlugins, config.Plugin) {
				missingPlugins = append(missingPlugins, config.Plugin)
			}
			continue
		}

		checksum, err := utils.FileHash(pluginPath)
		if err != nil {
			return nil, nil, err
		}

		requiredConnections[name] = &ConnectionData{
			Plugin:           remoteSchema,
			CheckSum:         checksum,
			ConnectionConfig: config.Config,
			ConnectionName:   config.Name,
		}
	}
	utils.LogTime("steampipeconfig.getRequiredConnections config-iteration end")

	return requiredConnections, missingPlugins, nil
}
