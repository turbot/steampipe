package db

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/turbot/steampipe/utils"

	"github.com/gertd/go-pluralize"
	"github.com/turbot/steampipe/connection_config"
)

func refreshConnections(client *Client) error {
	// first get a list of all existing schemas
	schemas := client.schemaMetadata.GetSchemas()

	// refresh the connection state file - the removes any connections which do not exist in the list of current schema
	log.Println("[TRACE] refreshConnections")
	updates, err := connection_config.GetConnectionsToUpdate(schemas)
	if err != nil {
		return err
	}
	log.Printf("[TRACE] updates: %+v\n", updates)

	missingCount := len(updates.MissingPlugins)
	if missingCount > 0 {
		// if any plugins are missing, error for now but we could prompt for an install
		p := pluralize.NewClient()
		return fmt.Errorf("%d %s referenced in the connection config not installed: \n  %v",
			missingCount,
			p.Pluralize("plugin", missingCount, false),
			strings.Join(updates.MissingPlugins, "\n  "))
	}

	var connectionQueries []string
	numUpdates := len(updates.Update)
	if numUpdates > 0 {
		s := utils.ShowSpinner("Refreshing connections...")
		defer utils.StopSpinner(s)
		connectionQueries = getSchemaQueries(updates.Update)

		if commentQueries, err := getCommentQueries(updates.Update); err != nil {
			return err
		} else {
			// add comments queries into the list
			connectionQueries = append(connectionQueries, commentQueries...)
		}
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

func getSchemaQueries(updates connection_config.ConnectionMap) []string {
	var schemaQueries []string
	for connectionName, plugin := range updates {
		remoteSchema := connection_config.PluginFQNToSchemaName(plugin.Plugin)
		log.Printf("[TRACE] update connection %s, plugin FQN %s, schema %s\n ", connectionName, plugin.Plugin, remoteSchema)
		schemaQueries = append(schemaQueries, updateConnectionQuery(connectionName, remoteSchema)...)
	}
	return schemaQueries
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

func getCommentQueries(updates connection_config.ConnectionMap) ([]string, error) {
	var commentQueries []string
	numUpdates := len(updates)
	var queryChan = make(chan []string, numUpdates)
	var errorChan = make(chan error, numUpdates)
	for connectionName, plugin := range updates {
		pluginFQN := plugin.Plugin

		// instantiate the connection plugin, and retrieve schema
		go getCommentsQueryAsync(pluginFQN, connectionName, plugin.ConnectionConfig, queryChan, errorChan)
	}

	for i := 0; i < numUpdates; i++ {
		select {
		case err := <-errorChan:
			log.Println("[TRACE] hydrate err chan select", "error", err)
			return nil, err
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("timed out retrieving schema from plugins")
		case a := <-queryChan:
			commentQueries = append(commentQueries, a...)
		}
	}

	return commentQueries, nil
}

func getCommentsQueryAsync(pluginFQN string, connectionName string, connectionConfig string, queryChan chan []string, errorChan chan error) {
	opts := &connection_config.ConnectionPluginOptions{
		PluginFQN:        pluginFQN,
		ConnectionName:   connectionName,
		ConnectionConfig: connectionConfig,
		DisableLogger:    true,
	}
	p, err := connection_config.CreateConnectionPlugin(opts)
	if err != nil {
		errorChan <- err
		return
	}
	queryChan <- commentsQuery(p)

	p.Plugin.Client.Kill()
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
