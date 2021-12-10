package db_local

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/plugin_manager"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

// RefreshConnections loads required connections from config
// and update the database schema and search path to reflect the required connections
// return whether any changes have been made
func (c *LocalDbClient) refreshConnections(ctx context.Context) *steampipeconfig.RefreshConnectionResult {
	res := &steampipeconfig.RefreshConnectionResult{}
	utils.LogTime("db.refreshConnections start")
	defer utils.LogTime("db.refreshConnections end")

	// get a list of all existing schema names
	schemaNames := c.client.ForeignSchemas()

	// determine any necessary connection updates
	connectionUpdates, res := steampipeconfig.NewConnectionUpdates(schemaNames)
	if res.Error != nil {
		return res
	}

	// if any plugins are missing, error for now but we could prompt for an install
	missingCount := len(connectionUpdates.MissingPlugins)
	if missingCount > 0 {
		res.Error = fmt.Errorf("%d %s referenced in the connection config not installed: \n  %v",
			missingCount,
			utils.Pluralize("plugin", missingCount),
			strings.Join(connectionUpdates.MissingPlugins, "\n  "))
		return res
	}

	// now build list of necessary queries to perform the update
	connectionQueries, otherRes := c.buildConnectionUpdateQueries(connectionUpdates)
	// merge results into local results
	res.Merge(otherRes)
	if res.Error != nil {
		return res
	}

	log.Printf("[TRACE] refreshConnections, %d connection update %s\n", len(connectionQueries), utils.Pluralize("query", len(connectionQueries)))

	// if there are no connection queries, we are done
	if len(connectionQueries) == 0 {
		return res
	}

	// so there ARE connections to update
	// execute the connection queries
	if err := executeConnectionQueries(ctx, connectionQueries); err != nil {
		res.Error = err
		return res
	}

	// now serialise the connection state
	// update required connections with the schema mode from the connection state and schema hash from the hash map
	if err := steampipeconfig.SaveConnectionState(connectionUpdates.RequiredConnectionState); err != nil {
		res.Error = err
		return res
	}
	// reload the database foreign schema names, since they have changed
	// this is to ensuire search paths are correctly updated
	log.Println("[TRACE] RefreshConnections: reloading foreign schema names")
	c.LoadForeignSchemaNames(ctx)

	res.UpdatedConnections = true
	return res

}

func (c *LocalDbClient) buildConnectionUpdateQueries(connectionUpdates *steampipeconfig.ConnectionUpdates) ([]string, *steampipeconfig.RefreshConnectionResult) {
	var connectionQueries []string
	var res *steampipeconfig.RefreshConnectionResult
	numUpdates := len(connectionUpdates.Update)

	log.Printf("[TRACE] buildConnectionUpdateQueries: num updates %d", numUpdates)

	if numUpdates > 0 {
		// find any plugins which use a newer sdk version than steampipe.
		validationFailures, validatedUpdates, validatedPlugins := steampipeconfig.ValidatePlugins(connectionUpdates.Update, connectionUpdates.ConnectionPlugins)
		if len(validationFailures) > 0 {
			res.Warnings = append(res.Warnings, steampipeconfig.BuildValidationWarningString(validationFailures))
		}

		// get schema queries - this updates schemas for validated plugins and drops schemas for unvalidated plugins
		connectionQueries = getSchemaQueries(validatedUpdates, validationFailures)
		if viper.GetBool(constants.ArgSchemaComments) {
			// add comments queries for validated connections
			connectionQueries = append(connectionQueries, getCommentQueries(validatedPlugins)...)
		}
	}

	for c := range connectionUpdates.Delete {
		log.Printf("[TRACE] delete connection %s\n ", c)
		connectionQueries = append(connectionQueries, deleteConnectionQuery(c)...)
	}
	return connectionQueries, res
}

func (c *LocalDbClient) updateConnectionMap() error {
	// load the connection state and cache it!
	log.Println("[TRACE]", "retrieving connection map")
	connectionMap, err := steampipeconfig.GetConnectionState(c.client.ForeignSchemas())
	if err != nil {
		return err
	}
	log.Println("[TRACE]", "setting connection map")
	c.connectionMap = &connectionMap

	return nil
}

func getSchemaQueries(updates steampipeconfig.ConnectionDataMap, failures []*steampipeconfig.ValidationFailure) []string {
	var schemaQueries []string
	for connectionName, connectionData := range updates {
		remoteSchema := plugin_manager.PluginFQNToSchemaName(connectionData.Plugin)
		log.Printf("[TRACE] update connection %s, plugin Name %s, schema %s, schemaQueries %v\n ", connectionName, connectionData.Plugin, remoteSchema, schemaQueries)
		queries := updateConnectionQuery(connectionName, remoteSchema)
		schemaQueries = append(schemaQueries, queries...)

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
		fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE;`, db_common.PgEscapeName(name)),
	}
}

func executeConnectionQueries(ctx context.Context, schemaQueries []string) error {
	log.Printf("[TRACE] there are connections to update\n")
	_, err := executeSqlAsRoot(ctx, schemaQueries...)
	if err != nil {
		return err
	}

	return nil
}
