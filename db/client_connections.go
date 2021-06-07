package db

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

// RefreshConnections :: load required connections from config
// and update the database schema and search path to reflect the required connections
// return whether any changes have been made
func (c *Client) RefreshConnections() (bool, error) {
	utils.LogTime("db.RefreshConnections start")
	defer utils.LogTime("db.RefreshConnections end")
	// load required connection from global config
	requiredConnections := steampipeconfig.Config.Connections

	// first get a list of all existing schemas
	schemas := c.schemaMetadata.GetSchemas()

	// refresh the connection state file - the removes any connections which do not exist in the list of current schema
	updates, err := steampipeconfig.GetConnectionsToUpdate(schemas, requiredConnections)
	if err != nil {
		return false, err
	}
	log.Printf("[TRACE] RefreshConnections, updates: %+v\n", updates)

	missingCount := len(updates.MissingPlugins)
	if missingCount > 0 {
		// if any plugins are missing, error for now but we could prompt for an install
		return false, fmt.Errorf("%d %s referenced in the connection config not installed: \n  %v",
			missingCount,
			utils.Pluralize("plugin", missingCount),
			strings.Join(updates.MissingPlugins, "\n  "))
	}

	var connectionQueries []string
	var warningString string
	numUpdates := len(updates.Update)
	if numUpdates > 0 {
		defer func() {
			// if any warnings were returned, display them on stderr
			if len(warningString) > 0 {
				// println writes to stderr
				println(warningString)
			}
		}()
		// in query, this can only start when in interactive
		if cmdconfig.Viper().GetBool(constants.ConfigKeyShowInteractiveOutput) {
			spin := display.ShowSpinner("Refreshing connections...")
			defer display.StopSpinner(spin)
		}

		// first instantiate connection plugins for all updates
		connectionPlugins, err := getConnectionPlugins(updates.Update)
		if err != nil {
			return false, err
		}
		// find any plugins which use a newer sdk version than steampipe.
		validationFailures, validatedUpdates, validatedPlugins := steampipeconfig.ValidatePlugins(updates.Update, connectionPlugins)
		warningString = steampipeconfig.BuildValidationWarningString(validationFailures)

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
		log.Println("[TRACE] no connections to update")
		return false, nil
	}

	// execute the connection queries
	if err = executeConnectionQueries(connectionQueries, updates); err != nil {
		return false, err
	}

	// so there ARE connections to update

	// reload the database schemas, since they have changed - otherwise we wouldn't be here
	log.Println("[TRACE] reloading schema")
	c.loadSchema()

	// update the service and client search paths
	log.Println("[TRACE] setting search path")
	c.SetServiceSearchPath()
	c.SetClientSearchPath()

	// finally update the connection map
	if err = c.updateConnectionMap(); err != nil {
		return false, err
	}

	return true, nil

}

func (c *Client) updateConnectionMap() error {
	// load the connection state and cache it!
	log.Println("[TRACE]", "retrieving connection map")
	connectionMap, err := steampipeconfig.GetConnectionState(c.schemaMetadata.GetSchemas())
	if err != nil {
		return err
	}
	log.Println("[TRACE]", "setting connection map")
	c.connectionMap = &connectionMap

	return nil
}

func getConnectionPlugins(updates steampipeconfig.ConnectionMap) ([]*steampipeconfig.ConnectionPlugin, error) {
	var connectionPlugins []*steampipeconfig.ConnectionPlugin
	numUpdates := len(updates)
	var pluginChan = make(chan *steampipeconfig.ConnectionPlugin, numUpdates)
	var errorChan = make(chan error, numUpdates)
	for connectionName, connectionData := range updates {

		// instantiate the connection plugin, and retrieve schema
		go getConnectionPluginsAsync(connectionName, connectionData, pluginChan, errorChan)
	}

	for i := 0; i < numUpdates; i++ {
		select {
		case err := <-errorChan:
			log.Println("[TRACE] get connections err chan select", "error", err)
			return nil, err
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("timed out retrieving schema from plugins")
		case p := <-pluginChan:
			connectionPlugins = append(connectionPlugins, p)
		}
	}

	return connectionPlugins, nil
}

func getConnectionPluginsAsync(connectionName string, connectionData *steampipeconfig.ConnectionData, pluginChan chan *steampipeconfig.ConnectionPlugin, errorChan chan error) {
	opts := &steampipeconfig.ConnectionPluginInput{
		ConnectionName:    connectionName,
		PluginName:        connectionData.Plugin,
		ConnectionOptions: connectionData.ConnectionOptions,
		DisableLogger:     true}
	p, err := steampipeconfig.CreateConnectionPlugin(opts)
	if err != nil {
		errorChan <- err
		return
	}
	pluginChan <- p

	p.Plugin.Client.Kill()
}

func getSchemaQueries(updates steampipeconfig.ConnectionMap, failures []*steampipeconfig.ValidationFailure) []string {
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
	localSchema = PgEscapeName(localSchema)
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
		table := PgEscapeName(t)
		schemaName := PgEscapeName(p.ConnectionName)
		if schema.Description != "" {
			tableDescription := PgEscapeString(schema.Description)
			statements = append(statements, fmt.Sprintf("COMMENT ON FOREIGN TABLE %s.%s is %s;", schemaName, table, tableDescription))
		}
		for _, c := range schema.Columns {
			if c.Description != "" {
				column := PgEscapeName(c.Name)
				columnDescription := PgEscapeString(c.Description)
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
	log.Printf("[DEBUG] there are connections to update, queries: \n%s\n", schemaQueries)
	_, err := executeSqlAsRoot(schemaQueries)
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
