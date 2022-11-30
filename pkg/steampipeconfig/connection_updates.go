package steampipeconfig

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

type ConnectionUpdates struct {
	Update         ConnectionDataMap
	Delete         ConnectionDataMap
	MissingPlugins map[string][]modconfig.Connection
	// the connections which will exist after the update
	RequiredConnectionState ConnectionDataMap
	// connection plugins required to perform the updates
	ConnectionPlugins       map[string]*ConnectionPlugin
	ConnectionStateModified bool
	currentConnectionState  ConnectionDataMap
}

// NewConnectionUpdates returns updates to be made to the database to sync with connection config
func NewConnectionUpdates(schemaNames []string, forceUpdateConnectionNames ...string) (*ConnectionUpdates, *RefreshConnectionResult) {
	utils.LogTime("NewConnectionUpdates start")
	defer utils.LogTime("NewConnectionUpdates end")

	res := &RefreshConnectionResult{}
	// build connection data for all required connections
	// NOTE: this will NOT populate SchemaMode for the connections, as we need to load the schema for that
	// this will be updated below on the call to updateRequiredStateWithSchemaProperties
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
	currentConnectionState, stateModified, err := GetConnectionState(schemaNames)
	if err != nil {
		res.Error = err
		return nil, res
	}
	updates.currentConnectionState = currentConnectionState
	updates.ConnectionStateModified = stateModified

	// for any connections with dynamic schema, we need to reload their schema
	// instantiate connection plugins for all connections with dynamic schema - this will retrieve their current schema
	dynamicSchemaHashMap, connectionsPluginsWithDynamicSchema, err := getSchemaHashesForDynamicSchemas(requiredConnectionState, currentConnectionState)
	if err != nil {
		res.Error = err
		return nil, res
	}

	// connections to create/update
	for name, requiredConnectionData := range requiredConnectionState {
		// check whether this connection exists in the state
		currentConnectionData, schemaExistsInState := currentConnectionState[name]
		// if it does not exist, or is not equal, add to updates
		if helpers.StringSliceContains(forceUpdateConnectionNames, name) ||
			!schemaExistsInState ||
			!currentConnectionData.Equals(requiredConnectionData) {
			log.Printf("[TRACE] connection %s is out of date or missing\n", name)
			updates.Update[name] = requiredConnectionData
		}
	}

	// connections to delete - any connection which is in connection state but NOT required connections
	for connection, requiredPlugin := range currentConnectionState {
		if _, ok := requiredConnectionState[connection]; !ok {
			log.Printf("[TRACE] connection %s is no longer required\n", connection)
			updates.Delete[connection] = requiredPlugin
		}
	}

	// now for every connection with dynamic schema,
	// check whether the schema we have just fetched matches the existing db schema
	// if not, add to updates
	for name, requiredHash := range dynamicSchemaHashMap {
		// get the connection data from the loaded connection state
		connectionData, ok := currentConnectionState[name]
		// if the connection exists in the state, does the schemas hash match?
		if ok && connectionData.SchemaHash != requiredHash {
			updates.Update[name] = connectionData
		}
	}

	//  instantiate connection plugins for all updates
	otherRes := updates.populateConnectionPlugins(connectionsPluginsWithDynamicSchema)
	res.Merge(otherRes)
	if res.Error != nil {
		return nil, res
	}

	// set the schema mode and hash on the connection data in required state
	// this uses data from the ConnectionPlugins which we have now loaded
	updates.updateRequiredStateWithSchemaProperties(dynamicSchemaHashMap)
	res.Updates = updates
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
			v.SchemaMode = connectionPlugin.ConnectionMap[k].Schema.Mode
			// if the schema mode is dynamic and the hash is not set yet, calculate the value from the connection plugin schema
			// this will happen the first time we load a plugin - as schemaHashMap will NOT include the has
			// because we do not know yet that the plugin is dynamic
			if v.SchemaMode == plugin.SchemaModeDynamic && v.SchemaHash == "" {
				v.SchemaHash = pluginSchemaHash(connectionPlugin.ConnectionMap[k].Schema)
			}
		}

	}
}

func (u *ConnectionUpdates) populateConnectionPlugins(alreadyCreatedConnectionPlugins map[string]*ConnectionPlugin) *RefreshConnectionResult {
	// get list of connections to update:
	// - exclude connections already created
	// - for any aggregator connections, instantiate the first child connection instead
	connectionsToCreate := u.getConnectionsToCreate(alreadyCreatedConnectionPlugins)
	// now create them
	connectionPlugins, res := CreateConnectionPlugins(connectionsToCreate)
	if res.Error != nil {
		return res
	}
	// add back in the already created plugins
	for name, connectionPlugin := range alreadyCreatedConnectionPlugins {
		connectionPlugins[name] = connectionPlugin
	}
	// and set our ConnectionPlugins property
	u.ConnectionPlugins = connectionPlugins
	return res
}

func (u *ConnectionUpdates) getConnectionsToCreate(alreadyCreatedConnectionPlugins map[string]*ConnectionPlugin) []*modconfig.Connection {
	updateConnections := u.Update.Connections()
	// put connections into a map to avoid dupes
	var connectionMap = make(map[string]*modconfig.Connection, len(updateConnections))
	for _, connection := range updateConnections {
		// if this connection is an aggregator, instantiate its first child connection instead
		if connection.Type == modconfig.ConnectionTypeAggregator {
			connection = connection.FirstChild()
		}

		connectionMap[connection.Name] = connection
	}
	// NOTE - we may have already created some connection plugins (if they have dynamic schema)
	// - remove these from list of plugins to create
	for name := range alreadyCreatedConnectionPlugins {
		delete(connectionMap, name)
	}

	// now copy result into array
	res := make([]*modconfig.Connection, len(connectionMap))
	idx := 0
	for _, c := range connectionMap {
		res[idx] = c
		idx++
	}

	return res
}

func (u *ConnectionUpdates) HasUpdates() bool {
	return len(u.Update)+len(u.Delete) > 0
}

func (u *ConnectionUpdates) String() string {
	var op strings.Builder
	update := utils.SortedMapKeys(u.Update)
	delete := utils.SortedMapKeys(u.Delete)
	stateConnections := utils.SortedMapKeys(u.RequiredConnectionState)
	if len(update) > 0 {
		op.WriteString(fmt.Sprintf("Update: %s\n", strings.Join(update, ",")))
	}
	if len(delete) > 0 {
		op.WriteString(fmt.Sprintf("Delete: %s\n", strings.Join(delete, ",")))
	}
	if len(stateConnections) > 0 {
		op.WriteString(fmt.Sprintf("Connection state: %s\n", strings.Join(stateConnections, ",")))
	} else {
		op.WriteString(fmt.Sprintf("Connection state EMPTY\n"))
	}
	return op.String()
}

func getSchemaHashesForDynamicSchemas(requiredConnectionData ConnectionDataMap, connectionState ConnectionDataMap) (map[string]string, map[string]*ConnectionPlugin, error) {
	log.Printf("[TRACE] getSchemaHashesForDynamicSchemas")
	// for every required connection, check the connection state to determine whether the schema mode is 'dynamic'
	// if we have never loaded the connection, there will be no state, so we cannot retrieve this information
	// however in this case we will load the connection anyway
	// - at which point the state will be updated with the schema mode for the next time round

	var connectionsWithDynamicSchema = make(ConnectionDataMap)
	for requiredConnectionName, requiredConnection := range requiredConnectionData {
		if existingConnection, ok := connectionState[requiredConnectionName]; ok {
			// SchemaMode will be unpopulated for plugins using an older version of the sdk
			// that is fine, we treat that as SchemaModeDynamic
			if existingConnection.SchemaMode == plugin.SchemaModeDynamic {
				log.Printf("[TRACE] fetching schema for connection %s using dynamic plugin %s", requiredConnectionName, requiredConnection.Plugin)
				connectionsWithDynamicSchema[requiredConnectionName] = requiredConnection
			}
		}
	}
	connectionsPluginsWithDynamicSchema, res := CreateConnectionPlugins(connectionsWithDynamicSchema.Connections())
	if res.Error != nil {
		return nil, nil, res.Error
	}

	log.Printf("[TRACE] fetched schema for %d dynamic %s", len(connectionsPluginsWithDynamicSchema), utils.Pluralize("plugin", len(connectionsPluginsWithDynamicSchema)))

	hashMap := make(map[string]string)
	for name, c := range connectionsPluginsWithDynamicSchema {
		// update schema hash stored in required connections so it is persisted in the state if updates are made
		schemaHash := pluginSchemaHash(c.ConnectionMap[name].Schema)
		hashMap[name] = schemaHash
	}
	return hashMap, connectionsPluginsWithDynamicSchema, nil
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
	return helpers.GetMD5Hash(str)
}
