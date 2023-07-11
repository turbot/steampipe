package steampipeconfig

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"golang.org/x/exp/maps"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

type ConnectionUpdates struct {
	Update          ConnectionStateMap
	Delete          map[string]struct{}
	Disabled        map[string]struct{}
	MissingComments ConnectionStateMap

	MissingPlugins map[string][]modconfig.Connection
	// the connections which will exist after the update
	FinalConnectionState ConnectionStateMap
	// connection plugins required to perform the updates, keyed by connection name
	ConnectionPlugins      map[string]*ConnectionPlugin
	CurrentConnectionState ConnectionStateMap
	InvalidConnections     map[string]*ValidationFailure
}

// NewConnectionUpdates returns updates to be made to the database to sync with connection config
func NewConnectionUpdates(ctx context.Context, pool *pgxpool.Pool, forceUpdateConnectionNames ...string) (*ConnectionUpdates, *RefreshConnectionResult) {
	updates, res := populateConnectionUpdates(ctx, pool, forceUpdateConnectionNames...)
	if res.Error != nil {
		return updates, res
	}
	// validate the updates
	// this will validate all plugins and connection names  and remove any updates which use invalid connections
	updates.validate()

	return updates, res
}

func populateConnectionUpdates(ctx context.Context, pool *pgxpool.Pool, forceUpdateConnectionNames ...string) (*ConnectionUpdates, *RefreshConnectionResult) {
	utils.LogTime("NewConnectionUpdates start")
	defer utils.LogTime("NewConnectionUpdates end")
	log.Printf("[TRACE] NewConnectionUpdates")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("[WARN] failed to acquire conneciton from pool: %s", err.Error())
		return nil, NewErrorRefreshConnectionResult(err)
	}
	defer conn.Release()

	log.Printf("[INFO] Loading connection state")
	// load the connection state file and filter out any connections which are not in the list of schemas
	// this allows for the database being rebuilt,modified externally
	currentConnectionStateMap, err := LoadConnectionState(ctx, conn.Conn())
	if err != nil {
		log.Printf("[WARN] failed to load connection state: %s", err.Error())
		return nil, NewErrorRefreshConnectionResult(err)
	}

	// build connection data for all required connections
	// NOTE: this will NOT populate SchemaMode for the connections, as we need to load the schema for that
	// this will be updated below on the call to updateRequiredStateWithSchemaProperties
	requiredConnectionStateMap, missingPlugins, err := GetRequiredConnectionStateMap(GlobalConfig.Connections, currentConnectionStateMap)
	if err != nil {
		log.Printf("[WARN] failed to build required connection state: %s", err.Error())
		return nil, NewErrorRefreshConnectionResult(err)
	}
	log.Printf("[INFO] built required connection state")

	// build lookup of disabled connections
	disabled := make(map[string]struct{})
	for _, c := range requiredConnectionStateMap {
		if c.Disabled() {
			disabled[c.ConnectionName] = struct{}{}
		}
	}

	updates := &ConnectionUpdates{
		Delete:               make(map[string]struct{}),
		Disabled:             disabled,
		Update:               ConnectionStateMap{},
		MissingComments:      ConnectionStateMap{},
		MissingPlugins:       missingPlugins,
		FinalConnectionState: requiredConnectionStateMap,
		InvalidConnections:   make(map[string]*ValidationFailure),
	}

	log.Printf("[INFO] loaded connection state")
	updates.CurrentConnectionState = currentConnectionStateMap

	log.Printf("[INFO] loading dynamic schema hashes")

	// for any connections with dynamic schema, we need to reload their schema
	// instantiate connection plugins for all connections with dynamic schema - this will retrieve their current schema
	dynamicSchemaHashMap, connectionsPluginsWithDynamicSchema, err := getSchemaHashesForDynamicSchemas(requiredConnectionStateMap, currentConnectionStateMap)
	if err != nil {
		log.Printf("[WARN] getSchemaHashesForDynamicSchemas failed: %s", err.Error())
		return nil, NewErrorRefreshConnectionResult(err)
	}
	log.Printf("[INFO] connectionsPluginsWithDynamicSchema: %s", strings.Join(maps.Keys(connectionsPluginsWithDynamicSchema), "'"))

	log.Printf("[INFO] dynamicSchemaHashMap")
	for k, v := range dynamicSchemaHashMap {
		log.Printf("[INFO] %s: %s", k, v)
	}
	log.Printf("[INFO] identify connections to update")

	modTime := time.Now()
	// connections to create/update
	for name, requiredConnectionState := range requiredConnectionStateMap {
		// if the connection requires update, add to list
		if connectionRequiresUpdate(forceUpdateConnectionNames, name, currentConnectionStateMap, requiredConnectionState) {
			updates.Update[name] = requiredConnectionState
			log.Printf("[INFO] connection %s is out of date or missing. updates: %v", name, maps.Keys(updates.Update))

			// set the connection mod time of required connection data to now
			requiredConnectionState.ConnectionModTime = modTime
		}
	}

	log.Printf("[INFO] Identify connections to delete")
	// connections to delete - any connection which is in connection state but NOT required connections
	for name, currentState := range currentConnectionStateMap {
		if _, connectionRequired := requiredConnectionStateMap[name]; !connectionRequired {
			log.Printf("[TRACE] connection %s in current state but not in required state - marking for deletion\n", name)
			updates.Delete[name] = struct{}{}
		} else if updates.FinalConnectionState[name].Disabled() && !currentState.Disabled() {
			// if required connection state is disabled and it is not currently disabled, mark for deletion
			log.Printf("[TRACE] connection %s is disabled - marking for deletion\n", name)
			updates.Delete[name] = struct{}{}
		}
	}

	// if there are any foreign schemas which do not exist in currentConnectionState OR requiredConnectionState,
	// add them into deletions
	// (if they exist in required current state but not required state, they will already be marked for deletion)
	// load foreign schema names
	foreignSchemaNames, err := db_common.LoadForeignSchemaNames(ctx, conn.Conn())
	if err != nil {
		log.Printf("[WARN] failed to load foreign schema names: %s", err.Error())
		return nil, NewErrorRefreshConnectionResult(err)
	}
	for _, name := range foreignSchemaNames {
		_, existsInCurrentState := currentConnectionStateMap[name]
		_, existsInRequiredState := requiredConnectionStateMap[name]
		if !existsInCurrentState && !existsInRequiredState {
			log.Printf("[TRACE] connection %s exists in db foreign schemas state but not current or required state - marking for deletion\n", name)
			updates.Delete[name] = struct{}{}
		}
	}

	// now for every connection with dynamic schema,
	// check whether the schema we have just fetched matches the existing db schema
	// if not, add to updates
	for name, requiredHash := range dynamicSchemaHashMap {
		// get the connection data from the loaded connection state
		connectionData, ok := currentConnectionStateMap[name]
		// if the connection exists in the state, does the schemas hash match?
		if ok && connectionData.SchemaHash != requiredHash {
			updates.Update[name] = connectionData
		}
	}

	log.Printf("[TRACE] Connecting to plugins")
	// now identify any connections which are not being updated/deleted but which have not got comments set
	updates.IdentifyMissingComments()

	//  instantiate connection plugins for all updates (including comment updates)
	res := updates.populateConnectionPlugins(connectionsPluginsWithDynamicSchema)
	if res.Error != nil {
		return nil, res
	}

	// set the schema mode and hash on the connection data in required state
	// this uses data from the ConnectionPlugins which we have now loaded
	updates.updateRequiredStateWithSchemaProperties(dynamicSchemaHashMap)

	// for all updates/deletes, if there any aggregators of the same plugin type, update those as well
	updates.populateAggregators()
	res.Updates = updates
	return updates, res
}

func connectionRequiresUpdate(forceUpdateConnectionNames []string, name string, currentConnectionStateMap ConnectionStateMap, requiredConnectionState *ConnectionState) bool {
	// check whether this connection exists in the state
	currentConnectionState, schemaExistsInState := currentConnectionStateMap[name]

	// if the required plugin is not installed, return false
	if typehelpers.SafeString(requiredConnectionState.ConnectionError) == constants.ConnectionErrorPluginNotInstalled {
		return false
	}
	// if the connection has been disabled, return false
	if requiredConnectionState.Disabled() {
		return false
	}
	// is this is a new connection
	if !schemaExistsInState {
		return true
	}
	// if the connection has been enabled (i.e. if it was previously DISABLED) , return true
	if currentConnectionState.Disabled() {
		return true
	}

	// are we are forcing an update of this connection,
	if helpers.StringSliceContains(forceUpdateConnectionNames, name) {
		return true
	}

	// has this connection previously not fully loaded
	if currentConnectionState.State == constants.ConnectionStatePendingIncomplete {
		return true
	}
	// update if the connection state is different
	return !currentConnectionState.Equals(requiredConnectionState)
}

// update requiredConnections - set the schema hash and schema mode for all elements of FinalConnectionState
// default to the existing state, but if an update is required, get the updated value
func (u *ConnectionUpdates) updateRequiredStateWithSchemaProperties(dynamicSchemaHashMap map[string]string) {
	// we only need to update connections which are being updated
	for k, v := range u.FinalConnectionState {
		if currentConectionState, ok := u.CurrentConnectionState[k]; ok {
			v.SchemaHash = currentConectionState.SchemaHash
			v.SchemaMode = currentConectionState.SchemaMode
		}
		// if the schemaHashMap contains this connection, use that value
		if schemaHash, ok := dynamicSchemaHashMap[k]; ok {
			v.SchemaHash = schemaHash
		}
		// have we loaded a connection plugin for this connection
		// - if so us the schema mode from the schema  it has loaded
		if connectionPlugin, ok := u.ConnectionPlugins[k]; ok {
			v.SchemaMode = connectionPlugin.ConnectionMap[k].Schema.Mode
			// if the schema mode is dynamic and the hash is not set yet, calculate the value from the connection plugin schema
			// this will happen the first time we load a plugin - as schemaHashMap will NOT include the hash
			// because we do not know yet that the plugin is dynamic
			if v.SchemaMode == plugin.SchemaModeDynamic && v.SchemaHash == "" {
				v.SchemaHash = pluginSchemaHash(connectionPlugin.ConnectionMap[k].Schema)
			}
		}

	}
}

func (u *ConnectionUpdates) populateConnectionPlugins(alreadyCreatedConnectionPlugins map[string]*ConnectionPlugin) *RefreshConnectionResult {
	utils.LogTime("populateConnectionPlugins start")
	defer utils.LogTime("populateConnectionPlugins end")

	// get list of connections to update:
	// - add connections which will be updated or have the comments updated
	// - exclude connections already created
	// - for any aggregator connections, instantiate the first child connection instead
	connectionsToCreate := u.getConnectionsToCreate(alreadyCreatedConnectionPlugins)
	// now create them
	connectionPlugins, res := CreateConnectionPlugins(connectionsToCreate)
	// if any plugins failed to load, set those connections to error
	for c, reason := range res.FailedConnections {
		u.setError(c, reason)
	}

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

func (u *ConnectionUpdates) getConnectionsToCreate(alreadyCreatedConnectionPlugins map[string]*ConnectionPlugin) []string {
	// ensure we instantiate all plugins required for schema AND comment updates
	connections := append(maps.Keys(u.Update), maps.Keys(u.MissingComments)...)
	// put connections into a map to avoid dupes
	var connectionMap = make(map[string]*modconfig.Connection, len(connections))
	for _, connectionName := range connections {
		connection := GlobalConfig.Connections[connectionName]
		connectionMap[connectionName] = connection
		// if this connection is an aggregator, add all its children
		for _, child := range connection.Connections {
			connectionMap[child.Name] = child
		}
	}
	// NOTE - we may have already created some connection plugins (if they have dynamic schema)
	// - remove these from list of plugins to create
	for name := range alreadyCreatedConnectionPlugins {
		delete(connectionMap, name)
	}

	return maps.Keys(connectionMap)
}

func (u *ConnectionUpdates) HasUpdates() bool {
	return len(u.Update)+len(u.Delete)+len(u.MissingComments) > 0
}

func (u *ConnectionUpdates) String() string {
	var op strings.Builder
	update := utils.SortedMapKeys(u.Update)
	toDelete := maps.Keys(u.Delete)
	sort.Strings(toDelete)
	stateConnections := utils.SortedMapKeys(u.FinalConnectionState)
	if len(update) > 0 {
		op.WriteString(fmt.Sprintf("Update: %s\n", strings.Join(update, ",")))
	}
	if len(toDelete) > 0 {
		op.WriteString(fmt.Sprintf("Delete: %s\n", strings.Join(toDelete, ",")))
	}
	if len(stateConnections) > 0 {
		op.WriteString(fmt.Sprintf("Connection state: %s\n", strings.Join(stateConnections, ",")))
	} else {
		op.WriteString(fmt.Sprintf("Connection state EMPTY\n"))
	}
	return op.String()
}

func (u *ConnectionUpdates) setError(connectionName string, error string) {
	failedConnection, ok := u.FinalConnectionState[connectionName]
	if !ok {
		return
	}
	failedConnection.State = constants.ConnectionStateError
	failedConnection.SetError(error)
	// remove from updating (in case it is there)
	delete(u.Update, connectionName)
}

// IdentifyMissingComments identifies any connections which are not being updated/deleted but which have not got comments set
// NOTE: this mutates FinalConnectionState to set comment_set (if needed)
func (u *ConnectionUpdates) IdentifyMissingComments() {
	for name, state := range u.FinalConnectionState {
		// if plugin is not installed, skip
		if typehelpers.SafeString(state.ConnectionError) == constants.ConnectionErrorPluginNotInstalled {
			continue
		}
		if currentState, existsInCurrentState := u.CurrentConnectionState[name]; existsInCurrentState {
			if !currentState.CommentsSet {
				_, updating := u.Update[name]
				_, deleting := u.Delete[name]
				if !updating || deleting {
					u.MissingComments[name] = state
				}
			}
		}
	}
}

// DynamicUpdates returns the names of all dynamic plugins which are being updated
func (u *ConnectionUpdates) DynamicUpdates() []string {
	var dynamicUpdates []string
	for _, c := range u.Update {
		if c.SchemaMode == plugin.SchemaModeDynamic {
			dynamicUpdates = append(dynamicUpdates, c.ConnectionName)
		}
	}
	return dynamicUpdates
}

func (u *ConnectionUpdates) populateAggregators() {
	log.Printf("[INFO] populateAggregators")
	// build map of aggregator connections keyed by plugin
	pluginAggregatorMap := make(map[string][]string)

	for connectionName, state := range u.FinalConnectionState {
		if state.GetType() == modconfig.ConnectionTypeAggregator {
			pluginAggregatorMap[state.Plugin] = append(pluginAggregatorMap[state.Plugin], connectionName)
		}
	}

	log.Printf("[INFO] found %d %s with aggregators", len(pluginAggregatorMap), utils.Pluralize("plugin", len(pluginAggregatorMap)))

	// for all updates/deletes, if there any aggregators of the same plugin type, update those as well
	// build a map of all plugins with connecitons being updated/deleted
	modifiedPluginLookup := make(map[string]struct{})
	for _, c := range u.Update {
		modifiedPluginLookup[c.Plugin] = struct{}{}
	}
	for c := range u.Delete {
		plugin := u.CurrentConnectionState[c].Plugin
		modifiedPluginLookup[plugin] = struct{}{}
	}
	for plugin := range modifiedPluginLookup {
		aggregatorsForPlugin := pluginAggregatorMap[plugin]
		numAggregatorsForPlugin := len(aggregatorsForPlugin)
		if numAggregatorsForPlugin > 0 {
			log.Printf("[INFO] plugin %s has modified connections - marking  %d %s as requiring update", plugin, numAggregatorsForPlugin, utils.Pluralize("aggregator", numAggregatorsForPlugin))
			for _, aggregatorConnection := range aggregatorsForPlugin {
				u.Update[aggregatorConnection] = u.FinalConnectionState[aggregatorConnection]
			}
		}
	}

}

func getSchemaHashesForDynamicSchemas(requiredConnectionData ConnectionStateMap, connectionState ConnectionStateMap) (map[string]string, map[string]*ConnectionPlugin, error) {
	log.Printf("[TRACE] getSchemaHashesForDynamicSchemas")
	// for every required connection, check the connection state to determine whether the schema mode is 'dynamic'
	// if we have never loaded the connection, there will be no state, so we cannot retrieve this information
	// however in this case we will load the connection anyway
	// - at which point the state will be updated with the schema mode for the next time round

	var connectionsWithDynamicSchema = make(ConnectionStateMap)
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
	connectionsPluginsWithDynamicSchema, res := CreateConnectionPlugins(maps.Keys(connectionsWithDynamicSchema))
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
