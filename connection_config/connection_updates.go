package connection_config

import (
	"log"
	"reflect"

	"github.com/turbot/go-kit/helpers"
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

type ConnectionData struct {
	// the fully qualified name of the plugin
	Plugin string `yaml:"plugin"`
	// the checksum of the plugin file
	CheckSum string `yaml:"checkSum"`
	// connection name
	ConnectionName string
	FdwOptions     *FdwOptions
	PluginOptions  *PluginOptions
	// connection data (unparsed)
	ConnectionConfig string
}

func (p ConnectionData) equals(other *ConnectionData) bool {
	return p.Plugin == other.Plugin &&
		p.CheckSum == other.CheckSum &&
		p.ConnectionName == other.ConnectionName &&
		(p.FdwOptions == nil && other.FdwOptions == nil) || p.FdwOptions.equals(other.FdwOptions) &&
		(p.PluginOptions == nil && other.PluginOptions == nil) || p.PluginOptions.equals(other.PluginOptions) &&
		reflect.DeepEqual(p.ConnectionConfig, other.ConnectionConfig)
}

type ConnectionMap map[string]*ConnectionData

// GetConnectionsToUpdate :: returns updates to be made to the database to sync with connection config
func GetConnectionsToUpdate(schemas []string, connectionConfig map[string]*Connection) (*ConnectionUpdates, error) {
	log.Println("[TRACE] GetConnectionsToUpdate")
	// load the connection state file and filter out any connections which are not in the list of schemas
	// this allows for the database being rebuilt,modified externally
	connectionState, err := GetConnectionState(schemas)

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
		if !ok || !current.equals(requiredPlugin) {
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
func getRequiredConnections(connectionConfig map[string]*Connection) (ConnectionMap, []string, error) {

	requiredConnections := ConnectionMap{}
	var missingPlugins []string

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
			FdwOptions:       config.FdwOptions,
			PluginOptions:    config.PluginOptions,
		}
	}

	return requiredConnections, missingPlugins, nil
}
