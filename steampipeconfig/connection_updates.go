package steampipeconfig

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/utils"
)

type ConnectionUpdates struct {
	Update         ConnectionDataMap
	Delete         ConnectionDataMap
	MissingPlugins []string
	// the connections which will exist after the update
	RequiredConnectionState ConnectionDataMap
	// connection plugins required to perform the updates
	ConnectionPlugins      map[string]*ConnectionPlugin
	currentConnectionState ConnectionDataMap
}

// NewConnectionUpdates returns updates to be made to the database to sync with connection config
func NewConnectionUpdates(schemaNames []string) (*ConnectionUpdates, *RefreshConnectionResult) {
	utils.LogTime("NewConnectionUpdates start")
	defer utils.LogTime("NewConnectionUpdates end")

	res := &RefreshConnectionResult{}
	// build connection data for all required connections
	requiredConnectionState, missingPlugins, err := NewConnectionDataMap(GlobalConfig.Connections)
	if err != nil {
		res.Error = err
		return nil, res
	}

	updates := &ConnectionUpdates{
		Update:                  ConnectionDataMap{},
		Delete:                  ConnectionDataMap{},
		MissingPlugins:          missingPlugins,
		RequiredConnectionState: requiredConnectionState,
	}

	// load the connection state file and filter out any connections which are not in the list of schemas
	// this allows for the database being rebuilt,modified externally
	connectionState, err := GetConnectionState(schemaNames)
	if err != nil {
		res.Error = err
		return nil, res
	}
	updates.currentConnectionState = connectionState

	// for any connections with dynamic schema, we need to reload their schema
	// instantiate connection plugins for all connections with dynamic schema - this will retrieve their current schema
	dynamicSchemaHashMap, connectionsPluginsWithDynamicSchema, res := getSchemaHashesForDynamicSchemas(requiredConnectionState, connectionState)
	if res.Error != nil {
		return nil, res
	}

	// connections to create/update
	for name, requiredConnectionData := range requiredConnectionState {
		// check whether this connection exists in the state
		currentConnectionData, ok := connectionState[name]
		// if it does not exist, or is not equal, add to updates
		if !ok || !currentConnectionData.Equals(requiredConnectionData) {
			log.Printf("[TRACE] connection %s is out of date or missing\n", name)
			updates.Update[name] = requiredConnectionData
		}
	}

	// connections to delete - any connection which is in connection state but NOT required connections
	for connection, requiredPlugin := range connectionState {
		if _, ok := requiredConnectionState[connection]; !ok {
			log.Printf("[TRACE] connection %s is no longer required\n", connection)
			updates.Delete[connection] = requiredPlugin
		}
	}

	// now for every connection with dynamic schema, check whether the schema we have just fetched
	// matches the existing db schema
	for name, requiredHash := range dynamicSchemaHashMap {
		// get the connection data from the loaded connection state
		connectionData, ok := connectionState[name]
		// if the connection exists in the state, does the schemas hash match?
		if ok && connectionData.SchemaHash != requiredHash {
			updates.Update[name] = connectionData
		}
	}

	//  instantiate connection plugins for all updates
	// NOTE - we may have already created some connection plugins (if they have dynamic schema)
	// - pass in the list of connection plugins we have already loaded

	connectionPlugins, otherRes := createConnectionPlugins(updates.Update, connectionsPluginsWithDynamicSchema)
	// merge results into local results
	res.Merge(otherRes)
	if res.Error != nil {
		return nil, res
	}

	updates.ConnectionPlugins = connectionPlugins
	// set the schema mode and hash on the connection data in required state
	updates.updateRequiredStateWithSchemaProperties(dynamicSchemaHashMap)

	return updates, res
}

// update requiredConnections - set the schema hash and schema mode for all elements of RequiredConnectionState
// default to the existing state, but if anm update is required, get the updated value
func (u *ConnectionUpdates) updateRequiredStateWithSchemaProperties(schemaHashMap map[string]string) {
	// we only need to update connections which are being updated
	for k, v := range u.RequiredConnectionState {
		if currentConectionState, ok := u.currentConnectionState[k]; ok {
			v.SchemaHash = currentConectionState.SchemaHash
			v.SchemaMode = currentConectionState.SchemaMode
		}
		// if the schemaHashMap contains this connection, use that value
		if schemaHash, ok := schemaHashMap[k]; ok {
			v.SchemaHash = schemaHash
		}
		// have we loaded a connection plugin for this connection
		// - if so us the schema mode from the schema  it has loaded
		if connectionPlugin, ok := u.ConnectionPlugins[k]; ok {
			v.SchemaMode = connectionPlugin.Schema.Mode
			// if the schema hash is still not set, calculate the value from the conneciton plugin schema
			// this will happen the first time we load a plugin
			if v.SchemaHash == "" {
				v.SchemaHash = pluginSchemaHash(connectionPlugin.Schema)
			}
		}

	}
}

func createConnectionPlugins(requiredConnections ConnectionDataMap, alreadyLoaded map[string]*ConnectionPlugin) (map[string]*ConnectionPlugin, *RefreshConnectionResult) {
	if alreadyLoaded == nil {
		alreadyLoaded = make(map[string]*ConnectionPlugin)
	}
	res := &RefreshConnectionResult{}

	// initialise result to the plugins we have already loaded
	var connectionPlugins = alreadyLoaded

	// do we have anything to do?
	if len(requiredConnections) <= len(alreadyLoaded) {
		return connectionPlugins, res
	}

	// create channels buffered to hold all updates
	numPluginsToCreate := len(requiredConnections) - len(alreadyLoaded)

	var pluginChan = make(chan *ConnectionPlugin, numPluginsToCreate)
	var errorChan = make(chan error, numPluginsToCreate)

	for connectionName, connectionData := range requiredConnections {
		// if we have NOT already loaded this plugin, do so
		if _, ok := alreadyLoaded[connectionName]; !ok {
			// instantiate the connection plugin, and retrieve schema
			go getConnectionPluginAsync(connectionData, pluginChan, errorChan)
		}
	}

	for i := 0; i < numPluginsToCreate; i++ {
		select {
		case err := <-errorChan:
			log.Println("[TRACE] get connections err chan select - adding warning", "error", err)
			res.Warnings = append(res.Warnings, err.Error())
		case p := <-pluginChan:
			connectionPlugins[p.ConnectionName] = p
		case <-time.After(10 * time.Second):
			res.Error = fmt.Errorf("timed out retrieving schema from plugins")
			return nil, res
		}
	}

	return connectionPlugins, res
}

func getConnectionPluginAsync(connectionData *ConnectionData, pluginChan chan *ConnectionPlugin, errorChan chan error) {
	p, err := CreateConnectionPlugin(connectionData.Connection, true)
	if err != nil {
		errorChan <- err
		return
	}
	pluginChan <- p

	p.Plugin.Client.Kill()
}

func getSchemaHashesForDynamicSchemas(requiredConnectionData ConnectionDataMap, connectionState ConnectionDataMap) (map[string]string, map[string]*ConnectionPlugin, *RefreshConnectionResult) {
	// for every required connection, check the connection state to determine whether it has a dynamic schema
	// if we have never loaded the conneciton, there wil be no stste so we cannot retrieve this informaiton
	// however in this case we will load the connection anyway

	var connectionsWithDynamicSchema = make(ConnectionDataMap)
	for requiredConnectionName, requiredConnection := range requiredConnectionData {
		if existingConnection, ok := connectionState[requiredConnectionName]; ok {
			// SchemaMode will be unpopulated for plugins using an older version of the sdk
			// that is fine, we treat that as SchemaModeDynamic
			if existingConnection.SchemaMode == plugin.SchemaModeDynamic {
				connectionsWithDynamicSchema[requiredConnectionName] = requiredConnection
			}
		}
	}

	connectionsPluginsWithDynamicSchema, res := createConnectionPlugins(connectionsWithDynamicSchema, nil)
	hashMap := make(map[string]string)
	for name, c := range connectionsPluginsWithDynamicSchema {
		// update schema hash stored in required connections so it is persisted in the state ius updates are made
		schemaHash := pluginSchemaHash(c.Schema)
		hashMap[name] = schemaHash
	}
	return hashMap, connectionsPluginsWithDynamicSchema, res
}

func pluginSchemaHash(s *proto.Schema) string {
	var sb strings.Builder

	// build ordered list of tables
	var tables = make([]string, len(s.Schema))
	idx := 0
	for tableName := range s.Schema {
		tables[idx] = tableName
		idx++
	}
	sort.Strings(tables)

	// now build  a string from the ordered table schemas
	for _, tableName := range tables {
		sb.WriteString(tableName)
		tableSchema := s.Schema[tableName]
		for _, c := range tableSchema.Columns {
			sb.WriteString(c.Name)
			sb.WriteString(fmt.Sprintf("%d", c.Type))
		}
	}
	str := sb.String()
	return utils.GetMD5Hash(str)
}
