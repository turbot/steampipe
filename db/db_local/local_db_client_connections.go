package db_local

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

// RefreshConnections loads required connections from config
// and update the database schema and search path to reflect the required connections
// return whether any changes have been made
func (c *LocalDbClient) refreshConnections() *db_common.RefreshConnectionResult {
	res := &db_common.RefreshConnectionResult{}
	utils.LogTime("db.refreshConnections start")
	defer utils.LogTime("db.refreshConnections end")

	// load required connection from global config
	requiredConnections := steampipeconfig.GlobalConfig.Connections

	// first get a list of all existing schemas
	schemas := c.client.SchemaMetadata().GetSchemas()

	// build connection data for all required connections
	// (this is the conneciton with the required plugin and it's checkum
	requiredConnectionData, missingPlugins, err := steampipeconfig.GetRequiredConnectionData(requiredConnections)
	if err != nil {
		res.Error = err
		return res
	}

	// for any connections with dynamic schema, we need to reload their schema
	// instantiate connection plugins for all connections with dynamic schema - this will retrieve their current schema
	var connectionsWithDynamicSchema = requiredConnectionData.ConnectionsWithDynamicSchema()
	connectionsPluginsWithDynamicSchema, res := getConnectionPlugins(connectionsWithDynamicSchema, nil)
	if res.Error != nil {
		return res
	}

	updates, err := steampipeconfig.GetConnectionsToUpdate(schemas, requiredConnectionData, missingPlugins, connectionsPluginsWithDynamicSchema)
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
		// first instantiate connection plugins for all updates
		var connectionPlugins map[string]*steampipeconfig.ConnectionPlugin
		// NOTE - we may have already created some connection plugins (if they have dynamic schema)
		// - pass in the list of connection plugins we have already loaded
		// NOTE - reuse 'res' defined above
		connectionPlugins, res = getConnectionPlugins(updates.Update, connectionsPluginsWithDynamicSchema)
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

func getConnectionPlugins(requiredConnections steampipeconfig.ConnectionDataMap, alreadyLoaded map[string]*steampipeconfig.ConnectionPlugin) (map[string]*steampipeconfig.ConnectionPlugin, *db_common.RefreshConnectionResult) {
	res := &db_common.RefreshConnectionResult{}
	// initialise result to the plugins we have already loaded
	var connectionPlugins = alreadyLoaded

	// create channels buffered to hold all updates
	numPluginsToCreate := len(requiredConnections) - len(alreadyLoaded)
	var pluginChan = make(chan *steampipeconfig.ConnectionPlugin, numPluginsToCreate)
	var errorChan = make(chan error, numPluginsToCreate)

	for connectionName, connectionData := range requiredConnections {
		// if we have NOT already loaded this plugin, do so
		if _, ok := alreadyLoaded[connectionName]; !ok {
			// instantiate the connection plugin, and retrieve schema
			go getConnectionPluginAsync(connectionName, connectionData, pluginChan, errorChan)
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

func getConnectionPluginAsync(connectionName string, connectionData *steampipeconfig.ConnectionData, pluginChan chan *steampipeconfig.ConnectionPlugin, errorChan chan error) {
	p, err := steampipeconfig.CreateConnectionPlugin(connectionData.Connection, true)
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
