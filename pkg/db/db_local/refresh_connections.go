package db_local

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

func RefreshConnectionAndSearchPaths(ctx context.Context, forceUpdateConnectionNames ...string) *steampipeconfig.RefreshConnectionResult {
	conn, err := CreateLocalDbConnection(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return steampipeconfig.NewErrorRefreshConnectionResult(err)
	}

	foreignSchemaNames, err := db_common.LoadForeignSchemaNames(ctx, conn)
	if err != nil {
		return steampipeconfig.NewErrorRefreshConnectionResult(err)
	}
	statushooks.SetStatus(ctx, "Refreshing connections")
	res := refreshConnections(ctx, conn, foreignSchemaNames, forceUpdateConnectionNames...)
	if res.Error != nil {
		return res
	}

	statushooks.SetStatus(ctx, "Loading steampipe connections")
	//set user search path first - client may fall back to using it
	statushooks.SetStatus(ctx, "Setting up search path")

	// we need to send a muted ctx here since this function selects from the database
	// which by default puts up a "Loading" spinner. We don't want that here
	mutedCtx := statushooks.DisableStatusHooks(ctx)
	err = setUserSearchPath(mutedCtx, conn, foreignSchemaNames)
	if err != nil {
		res.Error = err
		return res
	}

	return res
}

// RefreshConnections loads required connections from config
// and update the database schema and search path to reflect the required connections
// return whether any changes have been made
func refreshConnections(ctx context.Context, conn *pgx.Conn, foreignSchemaNames []string, forceUpdateConnectionNames ...string) (res *steampipeconfig.RefreshConnectionResult) {
	utils.LogTime("db.refreshConnections start")
	defer utils.LogTime("db.refreshConnections end")

	// determine any necessary connection updates
	var connectionUpdates *steampipeconfig.ConnectionUpdates
	connectionUpdates, res = steampipeconfig.NewConnectionUpdates(foreignSchemaNames, forceUpdateConnectionNames...)
	defer logRefreshConnectionResults(connectionUpdates, res)
	if res.Error != nil {
		return res
	}

	// before finishing - be sure to save connection state if
	// 	- it was modified in the loading process (indicating it contained non-existent connections)
	//  - connections have been updated
	defer func() {
		if res.Error == nil && connectionUpdates.ConnectionStateModified || res.UpdatedConnections {
			// now serialise the connection state

			if res.Error == nil && connectionUpdates.ConnectionStateModified || res.UpdatedConnections {
				serialiseConnectionState(res, connectionUpdates)
			}
		}
	}()

	var connectionNames, pluginNames []string
	// add warning if there are connections left over, from missing plugins
	if len(connectionUpdates.MissingPlugins) > 0 {
		// warning
		for a, conns := range connectionUpdates.MissingPlugins {
			for _, con := range conns {
				connectionNames = append(connectionNames, con.Name)
			}
			pluginNames = append(pluginNames, utils.GetPluginName(a))
		}
		res.AddWarning(fmt.Sprintf("%d %s required by %s %s missing. To install, please run %s",
			len(pluginNames),
			utils.Pluralize("plugin", len(pluginNames)),
			utils.Pluralize("connection", len(connectionNames)),
			utils.Pluralize("is", len(pluginNames)),
			constants.Bold(fmt.Sprintf("steampipe plugin install %s", strings.Join(pluginNames, " ")))))
	}

	if !connectionUpdates.HasUpdates() {
		log.Println("[TRACE] RefreshConnections: no updates required")
		return res
	}

	// now build list of necessary queries to perform the update
	queryRes := executeConnectionUpdateQueries(ctx, connectionUpdates, conn)
	// merge results into local results
	res.Merge(queryRes)
	if res.Error != nil {
		return res
	}

	res.UpdatedConnections = true

	return res
}

func serialiseConnectionState(res *steampipeconfig.RefreshConnectionResult, connectionUpdates *steampipeconfig.ConnectionUpdates) {
	// now serialise the connection state
	connectionState := make(steampipeconfig.ConnectionDataMap, len(connectionUpdates.RequiredConnectionState))
	for k, v := range connectionUpdates.RequiredConnectionState {
		connectionState[k] = v
	}
	// NOTE: add any connection which failed
	for c := range res.FailedConnections {
		connectionState[c].Loaded = false
		connectionState[c].Error = "plugin failed to start"
	}
	for pluginName, connections := range connectionUpdates.MissingPlugins {
		// add in missing connections
		for _, c := range connections {
			connectionData := steampipeconfig.NewConnectionData(pluginName, &c, time.Now())
			connectionData.Loaded = false
			connectionData.Error = "plugin not installed"
			connectionState[c.Name] = connectionData
		}
	}

	// update connection state and write the missing and failed plugin connections
	if err := connectionState.Save(); err != nil {
		res.Error = err
	}
}

func logRefreshConnectionResults(updates *steampipeconfig.ConnectionUpdates, res *steampipeconfig.RefreshConnectionResult) {
	var cmdName = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command).Name()
	if cmdName != "plugin-manager" {
		return
	}

	var op strings.Builder
	if updates != nil {
		op.WriteString(fmt.Sprintf("%s", updates.String()))
	}
	if res != nil {
		op.WriteString(fmt.Sprintf("%s\n", res.String()))
	}

	log.Printf("[INFO] refresh connections: \n%s\n", helpers.Tabify(op.String(), "    "))
}

func executeConnectionUpdateQueries(ctx context.Context, connectionUpdates *steampipeconfig.ConnectionUpdates, conn *pgx.Conn) *steampipeconfig.RefreshConnectionResult {
	utils.LogTime("db.executeConnectionUpdateQueries start")
	defer utils.LogTime("db.executeConnectionUpdateQueries start")

	res := &steampipeconfig.RefreshConnectionResult{}

	numUpdates := len(connectionUpdates.Update)
	log.Printf("[TRACE] executeConnectionUpdateQueries: num updates %d", numUpdates)

	if numUpdates > 0 {
		// find any plugins which use a newer sdk version than steampipe.
		validationFailures, validatedUpdates, validatedPlugins := steampipeconfig.ValidatePlugins(connectionUpdates.Update, connectionUpdates.ConnectionPlugins)
		if len(validationFailures) > 0 {
			res.Warnings = append(res.Warnings, steampipeconfig.BuildValidationWarningString(validationFailures))
		}

		// get schema queries - this updates schemas for validated plugins and drops schemas for unvalidated plugins
		err := executeUpdateQueries(ctx, conn, validationFailures, validatedUpdates, validatedPlugins)
		if err != nil {
			log.Printf("[TRACE] executeUpdateQueries returned error: %v", err)
			res.Error = err
			return res
		}
	}

	for c := range connectionUpdates.Delete {
		utils.LogTime("delete connection start")
		log.Printf("[TRACE] delete connection %s\n ", c)
		query := getDeleteConnectionQuery(c)
		_, err := conn.Exec(ctx, query)
		if err != nil {
			res.Error = err
			return res
		}
		utils.LogTime("delete connection end")
	}

	return res
}

func executeUpdateQueries(ctx context.Context, rootClient *pgx.Conn, failures []*steampipeconfig.ValidationFailure, updates steampipeconfig.ConnectionDataMap, validatedPlugins map[string]*steampipeconfig.ConnectionPlugin) error {
	utils.LogTime("db.executeUpdateQueries start")
	defer utils.LogTime("db.executeUpdateQueries end")
	idx := 0
	numUpdates := len(updates)

	var builder strings.Builder

	// TODO execute in single transaction???
	log.Printf("[TRACE] executing %d update %s", numUpdates, utils.Pluralize("query", numUpdates))
	for connectionName, connectionData := range updates {
		remoteSchema := utils.PluginFQNToSchemaName(connectionData.Plugin)
		statements := []string{
			"lock table pg_namespace;",
			getUpdateConnectionQuery(connectionName, remoteSchema),
		}
		_, err := executeSqlInTransaction(ctx, rootClient, statements...)
		if err != nil {
			return err
		}
		idx++
	}

	log.Printf("[TRACE] all update queries executed")

	for _, failure := range failures {
		log.Printf("[TRACE] remove schema for connection failing validation connection %s, plugin Name %s\n ", failure.ConnectionName, failure.Plugin)
		if failure.ShouldDropIfExists {
			statements := []string{
				"lock table pg_namespace;",
				getDeleteConnectionQuery(failure.ConnectionName),
			}
			_, err := executeSqlInTransaction(ctx, rootClient, statements...)
			if err != nil {
				return err
			}
		}
	}

	if viper.GetBool(constants.ArgSchemaComments) {
		idx = 0
		builder.Reset()
		numCommentsUpdates := len(validatedPlugins)
		log.Printf("[TRACE] executing %d comment %s", numCommentsUpdates, utils.Pluralize("query", numCommentsUpdates))

		statements := []string{"lock table pg_namespace;"}
		for connectionName, connectionPlugin := range validatedPlugins {
			statements = append(statements, getCommentsQueryForPlugin(connectionName, connectionPlugin))
		}
		_, err := executeSqlInTransaction(ctx, rootClient, statements...)
		if err != nil {
			return err
		}
	}

	log.Printf("[TRACE] executeUpdateQueries complete")
	return nil
}

func getCommentsQueryForPlugin(connectionName string, p *steampipeconfig.ConnectionPlugin) string {
	var statements strings.Builder
	for t, schema := range p.ConnectionMap[connectionName].Schema.Schema {
		table := db_common.PgEscapeName(t)
		schemaName := db_common.PgEscapeName(connectionName)
		if schema.Description != "" {
			tableDescription := db_common.PgEscapeString(schema.Description)
			statements.WriteString(fmt.Sprintf("COMMENT ON FOREIGN TABLE %s.%s is %s;\n", schemaName, table, tableDescription))
		}
		for _, c := range schema.Columns {
			if c.Description != "" {
				column := db_common.PgEscapeName(c.Name)
				columnDescription := db_common.PgEscapeString(c.Description)
				statements.WriteString(fmt.Sprintf("COMMENT ON COLUMN %s.%s.%s is %s;\n", schemaName, table, column, columnDescription))
			}
		}
	}
	return statements.String()
}

func getUpdateConnectionQuery(localSchema, remoteSchema string) string {
	// escape the name
	localSchema = db_common.PgEscapeName(localSchema)

	var statements strings.Builder

	// Each connection has a unique schema. The schema, and all objects inside it,
	// are owned by the root user.
	statements.WriteString(fmt.Sprintf("drop schema if exists %s cascade;\n", localSchema))
	statements.WriteString(fmt.Sprintf("create schema %s;\n", localSchema))
	statements.WriteString(fmt.Sprintf("comment on schema %s is 'steampipe plugin: %s';\n", localSchema, remoteSchema))

	// Steampipe users are allowed to use the new schema
	statements.WriteString(fmt.Sprintf("grant usage on schema %s to steampipe_users;\n", localSchema))

	// Permissions are limited to select only, and should be granted for all new
	// objects. Steampipe users cannot create tables or modify data in the
	// connection schema - they need to use the public schema for that.  These
	// commands alter the defaults for any objects created in the future.
	// See https://www.postgresql.org/docs/12/ddl-priv.html
	statements.WriteString(fmt.Sprintf("alter default privileges in schema %s grant select on tables to steampipe_users;\n", localSchema))

	// If there are any objects already then grant their permissions now. (This
	// should not actually do anything at this point.)
	statements.WriteString(fmt.Sprintf("grant select on all tables in schema %s to steampipe_users;\n", localSchema))

	// Import the foreign schema into this connection.
	statements.WriteString(fmt.Sprintf("import foreign schema \"%s\" from server steampipe into %s;\n", remoteSchema, localSchema))

	return statements.String()
}

func getDeleteConnectionQuery(name string) string {
	return fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;\n", db_common.PgEscapeName(name))
}
