package db_local

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_client"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

// LocalDbClient wraps over DbClient
type LocalDbClient struct {
	client        *db_client.DbClient
	invoker       constants.Invoker
	connectionMap *steampipeconfig.ConnectionDataMap
}

// GetLocalClient starts service if needed and creates a new LocalDbClient
func GetLocalClient(invoker constants.Invoker) (db_common.Client, error) {
	// start db if necessary
	err := EnsureDbAndStartService(invoker)
	if err != nil {
		return nil, err
	}

	client, err := NewLocalClient(invoker)
	if err != nil {
		ShutdownService(invoker)
	}
	// NOTE:  client shutdown will shutdown service (if invoker matches)
	return client, nil
}

// NewLocalClient ensures that the database instance is running
// and returns a `Client` to interact with it
func NewLocalClient(invoker constants.Invoker) (*LocalDbClient, error) {
	utils.LogTime("db.NewLocalClient start")
	defer utils.LogTime("db.NewLocalClient end")

	connString, err := getLocalSteampipeConnectionString()
	if err != nil {
		return nil, err
	}
	dbClient, err := db_client.NewDbClient(connString)
	if err != nil {
		return nil, err
	}

	c := &LocalDbClient{client: dbClient, invoker: invoker}
	log.Printf("[TRACE] created local client %p", c)
	return c, nil
}

// Close implements Client
// close the connection to the database and shuts down the backend
func (c *LocalDbClient) Close() error {
	log.Printf("[TRACE] close local client %p", c)
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return err
		}
	}
	ShutdownService(c.invoker)
	return nil
}

// EnsureSessionState implements Client
func (c *LocalDbClient) SetEnsureSessionStateFunc(f db_common.EnsureSessionStateCallback) {
	c.client.SetEnsureSessionStateFunc(f)
}

// SchemaMetadata implements Client
func (c *LocalDbClient) SchemaMetadata() *schema.Metadata {
	return c.client.SchemaMetadata()
}

func (c *LocalDbClient) ConnectionMap() *steampipeconfig.ConnectionDataMap {
	return c.connectionMap
}

// LoadSchema  implements Client
func (c *LocalDbClient) LoadSchema() {
	c.client.LoadSchema()
}

// ExecuteSync implements Client
func (c *LocalDbClient) ExecuteSync(ctx context.Context, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error) {
	return c.client.ExecuteSync(ctx, query, disableSpinner)
}

// Execute implements Client
func (c *LocalDbClient) Execute(ctx context.Context, query string, disableSpinner bool) (res *queryresult.Result, err error) {
	return c.client.Execute(ctx, query, disableSpinner)
}

// CacheOn implements Client
func (c *LocalDbClient) CacheOn() error {
	return c.client.CacheOn()
}

// CacheOff implements Client
func (c *LocalDbClient) CacheOff() error {
	return c.client.CacheOff()
}

// CacheClear implements Client
func (c *LocalDbClient) CacheClear() error {
	return c.client.CacheClear()
}

// GetCurrentSearchPath implements Client
func (c *LocalDbClient) GetCurrentSearchPath() ([]string, error) {
	// NOTE: create a new client to do this, so we respond to any recent changes in user search path
	// (as the user search path may have changed  after creating client 'c', e.g. if connections have changed)
	newClient, err := NewLocalClient(constants.InvokerService)
	if err != nil {
		return nil, err
	}
	defer newClient.Close()
	return newClient.client.GetCurrentSearchPath()
}

// SetSessionSearchPath implements Client
func (c *LocalDbClient) SetSessionSearchPath(currentSearchPath ...string) error {
	return c.client.SetSessionSearchPath(currentSearchPath...)
}

// local only functions

func (c *LocalDbClient) RefreshConnectionAndSearchPaths() *db_common.RefreshConnectionResult {
	res := c.refreshConnections()
	if res.Error != nil {
		return res
	}
	if err := refreshFunctions(); err != nil {
		res.Error = err
		return res
	}

	// load the connection state and cache it!
	connectionMap, err := steampipeconfig.GetConnectionState(c.SchemaMetadata().GetSchemas())
	if err != nil {
		res.Error = err
		return res
	}
	c.connectionMap = &connectionMap
	// set user search path first - client may fall back to using it
	if err := c.setUserSearchPath(); err != nil {
		res.Error = err
		return res
	}

	// get current search path, creating a new client to ensure we pick up recent changes
	currentSearchPath, err := c.GetCurrentSearchPath()
	if err != nil {
		res.Error = err
		return res
	}
	if err := c.SetSessionSearchPath(currentSearchPath...); err != nil {
		res.Error = err
		return res
	}

	return res
}

// SetUserSearchPath sets the search path for the all steampipe users of the db service
// do this wy finding all users assigned to the role steampipe_users and set their search path
func (c *LocalDbClient) setUserSearchPath() error {
	log.Println("[Trace] SetUserSearchPath")
	var searchPath []string

	// is there a user search path in the config?
	// check ConfigKeyDatabaseSearchPath config (this is the value specified in the database config)
	if viper.IsSet(constants.ConfigKeyDatabaseSearchPath) {
		searchPath = viper.GetStringSlice(constants.ConfigKeyDatabaseSearchPath)
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.FunctionSchema)
	} else {
		// no config set - set user search path to default
		searchPath = c.getDefaultSearchPath()
	}

	// escape the schema names
	searchPath = db_common.PgEscapeSearchPath(searchPath)

	log.Println("[TRACE] setting user search path to", searchPath)

	// get all roles which are a member of steampipe_users
	query := fmt.Sprintf(`select usename from pg_user where pg_has_role(usename, '%s', 'member')`, constants.DatabaseUsersRole)
	res, err := c.ExecuteSync(context.Background(), query, true)
	if err != nil {
		return err
	}

	// set the search path for all these roles
	var queries []string
	for _, row := range res.Rows {
		rowResult := row.(*queryresult.RowResult)
		user := string(rowResult.Data[0].([]uint8))
		if user == "root" {
			continue
		}
		queries = append(queries, fmt.Sprintf(
			"alter user %s set search_path to %s;",
			user,
			strings.Join(searchPath, ","),
		))
	}
	query = strings.Join(queries, "\n")
	log.Printf("[TRACE] user search path sql: %s", query)
	_, err = executeSqlAsRoot(query)
	if err != nil {
		return err
	}
	return nil
}

// build default search path from the connection schemas, bookended with public and internal
func (c *LocalDbClient) getDefaultSearchPath() []string {
	searchPath := c.SchemaMetadata().GetSchemas()
	sort.Strings(searchPath)
	// add the 'public' schema as the first schema in the search_path. This makes it
	// easier for users to build and work with their own tables, and since it's normally
	// empty, doesn't make using steampipe tables any more difficult.
	searchPath = append([]string{"public"}, searchPath...)
	// add 'internal' schema as last schema in the search path
	searchPath = append(searchPath, constants.FunctionSchema)

	return searchPath
}

// RefreshConnections loads required connections from config
// and update the database schema and search path to reflect the required connections
// return whether any changes have been made
func (c *LocalDbClient) refreshConnections() *db_common.RefreshConnectionResult {

	res := &db_common.RefreshConnectionResult{}
	utils.LogTime("db.refreshConnections start")
	defer utils.LogTime("db.refreshConnections end")

	// load required connection from global config
	requiredConnections := steampipeconfig.Config.Connections

	// first get a list of all existing schemas
	schemas := c.client.SchemaMetadata().GetSchemas()

	// refresh the connection state file - the removes any connections which do not exist in the list of current schema
	updates, err := steampipeconfig.GetConnectionsToUpdate(schemas, requiredConnections)
	if err != nil {
		res.Error = err
		return res
	}
	log.Printf("[TRACE] refreshConnections, updates: %+v\n", updates)

	missingCount := len(updates.MissingPlugins)
	if missingCount > 0 {
		// if any plugins are missing, error for now but we could prompt for an install
		res.Error = fmt.Errorf("%d %s referenced in the connection config not installed: \n  %v",
			missingCount,
			utils.Pluralize("plugin", missingCount),
			strings.Join(updates.MissingPlugins, "\n  "))
		return res
	}

	var connectionQueries []string
	numUpdates := len(updates.Update)
	log.Printf("[TRACE] RefreshConnections: num updates %d", numUpdates)
	if numUpdates > 0 {
		// first instantiate connection plugins for all updates (reuse 'res' defined above)
		var connectionPlugins []*steampipeconfig.ConnectionPlugin
		connectionPlugins, res = getConnectionPlugins(updates.Update)
		if res.Error != nil {
			return res
		}

		// find any plugins which use a newer sdk version than steampipe.
		validationFailures, validatedUpdates, validatedPlugins := steampipeconfig.ValidatePlugins(updates.Update, connectionPlugins)
		if len(validationFailures) > 0 {
			res.Warnings = append(res.Warnings, steampipeconfig.BuildValidationWarningString(validationFailures))
		}

		// get schema queries - this updates schemas for validated plugins and drops schemas for unvalidated plugins
		connectionQueries = getSchemaQueries(validatedUpdates, validationFailures)
		// add comments queries for validated connections
		connectionQueries = append(connectionQueries, getCommentQueries(validatedPlugins)...)
	}

	for c := range updates.Delete {
		log.Printf("[TRACE] delete connection %s\n ", c)
		connectionQueries = append(connectionQueries, deleteConnectionQuery(c)...)
	}

	if len(connectionQueries) == 0 {
		return res
	}

	// execute the connection queries
	if err = executeConnectionQueries(connectionQueries, updates); err != nil {
		res.Error = err
		return res
	}

	// so there ARE connections to update

	// reload the database schemas, since they have changed - otherwise we wouldn't be here
	log.Println("[TRACE] RefreshConnections: reloading schema")
	c.LoadSchema()

	res.UpdatedConnections = true
	return res

}

func (c *LocalDbClient) updateConnectionMap() error {
	// load the connection state and cache it!
	log.Println("[TRACE]", "retrieving connection map")
	connectionMap, err := steampipeconfig.GetConnectionState(c.client.SchemaMetadata().GetSchemas())
	if err != nil {
		return err
	}
	log.Println("[TRACE]", "setting connection map")
	c.connectionMap = &connectionMap

	return nil
}

func getConnectionPlugins(updates steampipeconfig.ConnectionDataMap) ([]*steampipeconfig.ConnectionPlugin, *db_common.RefreshConnectionResult) {
	res := &db_common.RefreshConnectionResult{}
	var connectionPlugins []*steampipeconfig.ConnectionPlugin

	// create channels buffered to hold all updates
	numUpdates := len(updates)
	var pluginChan = make(chan *steampipeconfig.ConnectionPlugin, numUpdates)
	var errorChan = make(chan error, numUpdates)

	for connectionName, connectionData := range updates {
		// instantiate the connection plugin, and retrieve schema
		go getConnectionPluginAsync(connectionName, connectionData, pluginChan, errorChan)
	}

	for i := 0; i < numUpdates; i++ {
		select {
		case err := <-errorChan:
			log.Println("[TRACE] get connections err chan select - adding warning", "error", err)
			res.Warnings = append(res.Warnings, err.Error())
		case p := <-pluginChan:
			connectionPlugins = append(connectionPlugins, p)
		case <-time.After(10 * time.Second):
			res.Error = fmt.Errorf("timed out retrieving schema from plugins")
			return nil, res
		}
	}
	return connectionPlugins, res
}

func getConnectionPluginAsync(connectionName string, connectionData *steampipeconfig.ConnectionData, pluginChan chan *steampipeconfig.ConnectionPlugin, errorChan chan error) {
	opts := &steampipeconfig.ConnectionPluginInput{
		ConnectionName:    connectionName,
		PluginName:        connectionData.Plugin,
		ConnectionOptions: connectionData.ConnectionOptions,
		ConnectionConfig:  connectionData.ConnectionConfig,
		DisableLogger:     true}
	p, err := steampipeconfig.CreateConnectionPlugin(opts)
	if err != nil {
		errorChan <- err
		return
	}
	pluginChan <- p

	p.Plugin.Client.Kill()
}

func getSchemaQueries(updates steampipeconfig.ConnectionDataMap, failures []*steampipeconfig.ValidationFailure) []string {
	var schemaQueries []string
	for connectionName, plugin := range updates {
		remoteSchema := steampipeconfig.PluginFQNToSchemaName(plugin.Plugin)
		log.Printf("[TRACE] update connection %s, plugin Name %s, schema %s\n ", connectionName, plugin.Plugin, remoteSchema)
		schemaQueries = append(schemaQueries, updateConnectionQuery(connectionName, remoteSchema)...)
	}
	for _, failure := range failures {
		log.Printf("[TRACE] remove schema for conneciton failing validation connection %s, plugin Name %s\n ", failure.ConnectionName, failure.Plugin)
		if failure.ShouldDropIfExists {
			schemaQueries = append(schemaQueries, deleteConnectionQuery(failure.ConnectionName)...)
		}
	}

	return schemaQueries
}

func getCommentQueries(plugins []*steampipeconfig.ConnectionPlugin) []string {
	var commentQueries []string
	for _, plugin := range plugins {
		commentQueries = append(commentQueries, commentsQuery(plugin)...)
	}
	return commentQueries
}

func updateConnectionQuery(localSchema, remoteSchema string) []string {
	// escape the name
	localSchema = db_common.PgEscapeName(localSchema)
	return []string{

		// Each connection has a unique schema. The schema, and all objects inside it,
		// are owned by the root user.
		fmt.Sprintf(`drop schema if exists %s cascade;`, localSchema),
		fmt.Sprintf(`create schema %s;`, localSchema),
		fmt.Sprintf(`comment on schema %s is 'steampipe plugin: %s';`, localSchema, remoteSchema),

		// Steampipe users are allowed to use the new schema
		fmt.Sprintf(`grant usage on schema %s to steampipe_users;`, localSchema),

		// Permissions are limited to select only, and should be granted for all new
		// objects. Steampipe users cannot create tables or modify data in the
		// connection schema - they need to use the public schema for that.  These
		// commands alter the defaults for any objects created in the future.
		// See https://www.postgresql.org/docs/12/ddl-priv.html
		fmt.Sprintf(`alter default privileges in schema %s grant select on tables to steampipe_users;`, localSchema),

		// If there are any objects already then grant their permissions now. (This
		// should not actually do anything at this point.)
		fmt.Sprintf(`grant select on all tables in schema %s to steampipe_users;`, localSchema),

		// Import the foreign schema into this connection.
		fmt.Sprintf(`import foreign schema "%s" from server steampipe into %s;`, remoteSchema, localSchema),
	}
}

func commentsQuery(p *steampipeconfig.ConnectionPlugin) []string {
	var statements []string
	for t, schema := range p.Schema.Schema {
		table := db_common.PgEscapeName(t)
		schemaName := db_common.PgEscapeName(p.ConnectionName)
		if schema.Description != "" {
			tableDescription := db_common.PgEscapeString(schema.Description)
			statements = append(statements, fmt.Sprintf("COMMENT ON FOREIGN TABLE %s.%s is %s;", schemaName, table, tableDescription))
		}
		for _, c := range schema.Columns {
			if c.Description != "" {
				column := db_common.PgEscapeName(c.Name)
				columnDescription := db_common.PgEscapeString(c.Description)
				statements = append(statements, fmt.Sprintf("COMMENT ON COLUMN %s.%s.%s is %s;", schemaName, table, column, columnDescription))
			}
		}
	}
	return statements
}

func deleteConnectionQuery(name string) []string {
	return []string{
		fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE;`, name),
	}
}

func executeConnectionQueries(schemaQueries []string, updates *steampipeconfig.ConnectionUpdates) error {
	log.Printf("[TRACE] there are connections to update\n")
	_, err := executeSqlAsRoot(schemaQueries...)
	if err != nil {
		return err
	}

	// now update the state file
	err = steampipeconfig.SaveConnectionState(updates.RequiredConnections)
	if err != nil {
		return err
	}
	return nil
}
