package db_local

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/pluginmanager"
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

	if !connectionUpdates.HasUpdates() {
		log.Println("[TRACE] RefreshConnections: no updates required")
		return res
	}

	// now build list of necessary queries to perform the update
	otherRes := c.executeConnectionUpdateQueries(ctx, connectionUpdates)
	// merge results into local results
	res.Merge(otherRes)
	if res.Error != nil {
		return res
	}

	// now serialise the connection state
	// update required connections with the schema mode from the connection state and schema hash from the hash map
	if err := steampipeconfig.SaveConnectionState(connectionUpdates.RequiredConnectionState); err != nil {
		res.Error = err
		return res
	}
	// reload the database foreign schema names, since they have changed
	// this is to ensure search paths are correctly updated
	log.Println("[TRACE] RefreshConnections: reloading foreign schema names")
	c.LoadForeignSchemaNames(ctx)

	res.UpdatedConnections = true
	return res
}

func (c *LocalDbClient) executeConnectionUpdateQueries(ctx context.Context, connectionUpdates *steampipeconfig.ConnectionUpdates) *steampipeconfig.RefreshConnectionResult {

	res := &steampipeconfig.RefreshConnectionResult{}
	rootClient, err := createLocalDbClient(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		res.Error = err
		return res
	}

	numUpdates := len(connectionUpdates.Update)
	log.Printf("[TRACE] executeConnectionUpdateQueries: num updates %d", numUpdates)

	if numUpdates > 0 {
		// find any plugins which use a newer sdk version than steampipe.
		validationFailures, validatedUpdates, validatedPlugins := steampipeconfig.ValidatePlugins(connectionUpdates.Update, connectionUpdates.ConnectionPlugins)
		if len(validationFailures) > 0 {
			res.Warnings = append(res.Warnings, steampipeconfig.BuildValidationWarningString(validationFailures))
		}

		// get schema queries - this updates schemas for validated plugins and drops schemas for unvalidated plugins
		err := executeUpdateQueries(rootClient, validationFailures, validatedUpdates, validatedPlugins)
		if err != nil {
			res.Error = err
			return res
		}

	}

	for c := range connectionUpdates.Delete {
		log.Printf("[TRACE] delete connection %s\n ", c)
		query := getDeleteConnectionQuery(c)
		_, err := rootClient.Exec(query)
		if err != nil {
			res.Error = err
			return res
		}
	}

	return res
}

func executeUpdateQueries(rootClient *sql.DB, failures []*steampipeconfig.ValidationFailure, updates steampipeconfig.ConnectionDataMap, validatedPlugins []*steampipeconfig.ConnectionPlugin) error {
	defer func() {
		log.Println("[WARN] DONE!!!")
		f, _ := os.Create("/tmp/steampipe.pb.gz")
		runtime.GC()
		pprof.WriteHeapProfile(f)
		f.Close()
	}()

	idx := 0
	for connectionName, connectionData := range updates {
		remoteSchema := pluginmanager.PluginFQNToSchemaName(connectionData.Plugin)
		query := getUpdateConnectionQuery(connectionName, remoteSchema)
		idx++
		//log.Println(query)
		log.Println("[WARN]", idx)
		_, err := rootClient.Exec(query)
		if err != nil {
			return err
		}
	}
	for _, failure := range failures {
		log.Printf("[TRACE] remove schema for connection failing validation connection %s, plugin Name %s\n ", failure.ConnectionName, failure.Plugin)
		if failure.ShouldDropIfExists {
			query := getDeleteConnectionQuery(failure.ConnectionName)
			_, err := rootClient.Exec(query)
			if err != nil {
				return err
			}

		}
	}
	//if viper.GetBool(constants.ArgSchemaComments) {
	//	// add comments queries for validated connections
	//	query := getCommentsQuery(validatedPlugins)
	//	_, err := rootClient.Exec(query)
	//	if err != nil {
	//		return err
	//	}
	//}
	return nil
}

func getCommentsQuery(plugins []*steampipeconfig.ConnectionPlugin) string {
	var commentQueries []string
	for _, plugin := range plugins {
		commentQueries = append(commentQueries, getCommentsQueryForPlugin(plugin)...)
	}
	return strings.Join(commentQueries, ";\n")
}

func getCommentsQueryForPlugin(p *steampipeconfig.ConnectionPlugin) []string {
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

func getUpdateConnectionQuery(localSchema, remoteSchema string) string {
	// escape the name
	localSchema = db_common.PgEscapeName(localSchema)
	queries := []string{

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
	return strings.Join(queries, ";\n")
}

func getDeleteConnectionQuery(name string) string {
	queries := []string{
		fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE;`, db_common.PgEscapeName(name)),
	}
	return strings.Join(queries, ";\n")
}
