package db

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/connection_config"
	"github.com/turbot/steampipe/utils"
)

func RefreshConnections(client *Client) error {
	// first get a list of all existing schemas
	schemas := client.schemaMetadata.GetSchemas()

	// refresh the connection state file - the removes any connections which do not exist in the list of current schema
	log.Println("[TRACE] RefreshConnections")
	updates, err := connection_config.GetConnectionsToUpdate(schemas)
	if err != nil {
		return err
	}
	log.Printf("[TRACE] updates: %+v\n", updates)

	missingCount := len(updates.MissingPlugins)
	if missingCount > 0 {
		// if any plugins are missing, error for now but we could prompt for an install
		return fmt.Errorf("%d %s referenced in the connection config not installed: \n  %v",
			missingCount,
			utils.Pluralize("plugin", missingCount),
			strings.Join(updates.MissingPlugins, "\n  "))
	}

	var connectionQueries []string
	var warningString string
	numUpdates := len(updates.Update)
	if numUpdates > 0 {
		// in query, this can only start when in interactive
		if cmdconfig.Viper().GetBool(constants.ShowInteractiveOutputConfigKey) {
			spin := utils.ShowSpinner("Refreshing connections...")
			defer utils.StopSpinner(spin)
		}
		defer func() {
			// if any warnings were returned, display them on stderr
			if len(warningString) > 0 {
				// println writes to stderr
				println(warningString)
			}
		}()

		// first instantiate connection plugins for all updates
		connectionPlugins, err := getConnectionPlugins(updates.Update)
		if err != nil {
			return err
		}
		// find any plugins which use a newer sdk version than steampipe.
		validationFailures, validatedUpdates, validatedPlugins := connection_config.ValidatePlugins(updates.Update, connectionPlugins)
		warningString = connection_config.BuildValidationWarningString(validationFailures)

		// get schema queries - this updates schemas for validated plugins and drops schemas for unvalidated plugins
		connectionQueries = getSchemaQueries(validatedUpdates, validationFailures)
		// add comments queries for validated connections
		connectionQueries = append(connectionQueries, getCommentQueries(validatedPlugins)...)
	}

	for c := range updates.Delete {
		log.Printf("[TRACE] delete %s\n ", c)
		connectionQueries = append(connectionQueries, deleteConnectionQuery(c)...)
	}
	if len(connectionQueries) > 0 {
		if err = executeConnectionQueries(connectionQueries, updates); err != nil {
			return err
		}
	} else {
		log.Println("[DEBUG] no connections to update")
	}

	return updateConnectionMapAndSchema(client)
}

func getConnectionPlugins(updates connection_config.ConnectionMap) ([]*connection_config.ConnectionPlugin, error) {
	var connectionPlugins []*connection_config.ConnectionPlugin
	numUpdates := len(updates)
	var pluginChan = make(chan *connection_config.ConnectionPlugin, numUpdates)
	var errorChan = make(chan error, numUpdates)
	for connectionName, plugin := range updates {
		pluginFQN := plugin.Plugin
		connectionConfig := plugin.ConnectionConfig

		// instantiate the connection plugin, and retrieve schema
		go getConnectionPluginsAsync(pluginFQN, connectionName, connectionConfig, pluginChan, errorChan)
	}

	for i := 0; i < numUpdates; i++ {
		select {
		case err := <-errorChan:
			log.Println("[TRACE] hydrate err chan select", "error", err)
			return nil, err
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("timed out retrieving schema from plugins")
		case p := <-pluginChan:
			connectionPlugins = append(connectionPlugins, p)
		}
	}

	return connectionPlugins, nil
}

func getConnectionPluginsAsync(pluginFQN string, connectionName string, connectionConfig string, pluginChan chan *connection_config.ConnectionPlugin, errorChan chan error) {
	opts := &connection_config.ConnectionPluginOptions{
		PluginFQN:        pluginFQN,
		ConnectionName:   connectionName,
		ConnectionConfig: connectionConfig,
		DisableLogger:    true}
	p, err := connection_config.CreateConnectionPlugin(opts)
	if err != nil {
		errorChan <- err
		return
	}
	pluginChan <- p

	p.Plugin.Client.Kill()
}

func getSchemaQueries(updates connection_config.ConnectionMap, failures []*connection_config.ValidationFailure) []string {
	var schemaQueries []string
	for connectionName, plugin := range updates {
		remoteSchema := connection_config.PluginFQNToSchemaName(plugin.Plugin)
		log.Printf("[TRACE] update connection %s, plugin FQN %s, schema %s\n ", connectionName, plugin.Plugin, remoteSchema)
		schemaQueries = append(schemaQueries, updateConnectionQuery(connectionName, remoteSchema)...)
	}
	for _, failure := range failures {
		log.Printf("[TRACE] remove schema for conneciton failing validation connection %s, plugin FQN %s\n ", failure.ConnectionName, failure.Plugin)
		schemaQueries = append(schemaQueries, deleteConnectionQuery(failure.ConnectionName)...)
	}

	return schemaQueries
}

func getCommentQueries(plugins []*connection_config.ConnectionPlugin) []string {
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

func commentsQuery(p *connection_config.ConnectionPlugin) []string {
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

func executeConnectionQueries(schemaQueries []string, updates *connection_config.ConnectionUpdates) error {
	client, err := createSteampipeRootDbClient()
	if err != nil {
		return err
	}
	defer func() {
		client.Close()
	}()

	// combine queries
	schemaQueryString := strings.Join(schemaQueries, "\n")

	log.Printf("[DEBUG] there are connections to update, query: \n%s\n", schemaQueryString)
	_, err = client.Exec(schemaQueryString)
	if err != nil {
		return err
	}
	/* TODO - Log results
	log.Println("[TRACE] refresh connection results")
	for row := range schemaResult {
		log.Printf("[TRACE] %v\n", row)
	}
	*/

	// now update the state file
	err = connection_config.SaveConnectionState(updates.RequiredConnections)
	if err != nil {
		return err
	}
	return nil
}

func updateConnectionMapAndSchema(client *Client) error {
	// reload the database schemas, since they have changed
	// otherwise we wouldn't be here
	log.Println("[TRACE] reloading schema")
	client.loadSchema()

	// set the search path with the updates
	log.Println("[TRACE] setting search path")
	client.setSearchPath()

	// load the connection state and cache it!
	log.Println("[TRACE]", "retrieving connection map")
	connectionMap, err := connection_config.GetConnectionState(clientSingleton.schemaMetadata.GetSchemas())
	if err != nil {
		return err
	}
	log.Println("[TRACE]", "setting connection map")
	client.connectionMap = &connectionMap

	return nil
}
