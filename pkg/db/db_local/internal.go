package db_local

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/connection_state"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

//// dropLegacySchemas drops the legacy 'steampipe_command' schema if it exists
//// and the 'internal' schema if it contains only the 'glob' function
//// and maybe the 'connection_state' table
//func dropLegacySchemas(ctx context.Context, conn *pgx.Conn) error {
//	utils.LogTime("db_local.dropLegacySchema start")
//	defer utils.LogTime("db_local.dropLegacySchema end")
//
//	return error_helpers.CombineErrors(
//		dropLegacyInternalSchema(ctx, conn),
//		dropLegacySteampipeCommandSchema(ctx, conn),
//	)
//}
//
//// dropLegacySteampipeCommandSchema drops the 'steampipe_command' schema if it exists
//func dropLegacySteampipeCommandSchema(ctx context.Context, conn *pgx.Conn) error {
//	utils.LogTime("db_local.dropLegacySteampipeCommand start")
//	defer utils.LogTime("db_local.dropLegacySteampipeCommand end")
//
//	_, err := conn.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", constants.LegacyCommandSchema))
//	return err
//}

// dropLegacyInternalSchema looks for a schema named 'internal'
// which has a function called 'glob' and maybe a table named 'connection_state'
// and drops it
func dropLegacyInternalSchema(ctx context.Context, conn *pgx.Conn) error {
	utils.LogTime("db_local.dropLegacyInternal start")
	defer utils.LogTime("db_local.dropLegacyInternal end")

	if exists, err := legacyInternalExists(ctx, conn); err == nil && !exists {
		log.Println("[TRACE] could not find legacy 'internal' schema")
		return nil
	}

	log.Println("[TRACE] dropping legacy 'internal' schema")
	if _, err := conn.Exec(ctx, fmt.Sprintf("DROP SCHEMA %s CASCADE", constants.LegacyInternalSchema)); err != nil {
		return sperr.WrapWithMessage(err, "could not drop legacy schema: '%s'", constants.LegacyInternalSchema)
	}
	log.Println("[TRACE] dropped legacy 'internal' schema")

	return nil
}

// legacyInternalExists looks for a schema named 'internal'
// which has a function called 'glob' and maybe a table named 'connection_state'
func legacyInternalExists(ctx context.Context, conn *pgx.Conn) (bool, error) {
	utils.LogTime("db_local.isLegacyInternalExists start")
	defer utils.LogTime("db_local.isLegacyInternalExists end")

	log.Println("[TRACE] querying for legacy 'internal' schema")

	legacySchemaCountQuery := `
WITH 
internal_functions AS (
		SELECT
			COALESCE(STRING_AGG(DISTINCT(p.proname),','),'') as function_names
		FROM
			pg_proc p
			LEFT JOIN pg_namespace n ON p.pronamespace = n.oid
		WHERE
			n.nspname = $1
),
internal_tables AS (
		SELECT 
				COALESCE(STRING_AGG(DISTINCT(table_name),','),'') as table_names
		FROM 
				information_schema.tables 
		WHERE 
				table_schema = $1
)
SELECT 
		internal_functions.function_names, 
		internal_tables.table_names 
FROM
		internal_functions 
INNER JOIN
		internal_tables
		ON true;
	`

	row := conn.QueryRow(ctx, legacySchemaCountQuery, constants.LegacyInternalSchema)

	var functionNames string
	var tableNames string
	err := row.Scan(&functionNames, &tableNames)
	if err != nil {
		return false, sperr.WrapWithMessage(err, "could not query legacy 'internal' schema objects: '%s'", constants.LegacyInternalSchema)
	}

	if len(functionNames) == 0 && len(tableNames) == 0 {
		log.Println("[TRACE] could not find any objects in 'internal' - skipping drop")
		return false, nil
	}

	functions := strings.Split(functionNames, ",")
	tables := strings.Split(tableNames, ",")

	log.Println("[TRACE] isLegacyInternalExists: available function names", functions)
	log.Println("[TRACE] isLegacyInternalExists: available table names", tables)

	expectedFunctions := map[string]bool{
		"glob": true,
	}
	expectedTables := map[string]bool{
		"connection_state":             true, // legacy table name
		constants.ConnectionStateTable: true,
	}

	for _, f := range functions {
		if !expectedFunctions[f] {
			log.Println("[TRACE] isLegacyInternalExists: unexpected function", f)
			return false, nil
		}
	}

	for _, t := range tables {
		if !expectedTables[t] {
			log.Println("[TRACE] isLegacyInternalExists: unexpected table", t)
			return false, nil
		}
	}

	return true, nil
}

func setupInternal(ctx context.Context, conn *pgx.Conn) error {
	statushooks.SetStatus(ctx, "Dropping legacy schema")
	if err := dropLegacyInternalSchema(ctx, conn); err != nil {
		// do not fail
		// worst case scenario is that we have a couple of extra schema
		// these won't be in the search path anyway
		log.Println("[INFO] failed to drop legacy 'internal' schema", err)
	}

	// setup internal schema
	// this includes setting the state of all connections in the connection_state table to pending
	statushooks.SetStatus(ctx, "Setting up internal schema")

	utils.LogTime("db_local.setupInternal start")
	defer utils.LogTime("db_local.setupInternal end")

	queries := []string{
		"lock table pg_namespace;",
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, constants.InternalSchema),
		fmt.Sprintf(`GRANT USAGE ON SCHEMA %s TO %s;`, constants.InternalSchema, constants.DatabaseUsersRole),
		fmt.Sprintf("IMPORT FOREIGN SCHEMA \"%s\" FROM SERVER steampipe INTO %s;", constants.InternalSchema, constants.InternalSchema),
		fmt.Sprintf("GRANT INSERT ON %s.%s TO %s;", constants.InternalSchema, constants.ForeignTableSettings, constants.DatabaseUsersRole),
		fmt.Sprintf("GRANT SELECT ON %s.%s TO %s;", constants.InternalSchema, constants.ForeignTableScanMetadata, constants.DatabaseUsersRole),
		// legacy command schema support
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, constants.LegacyCommandSchema),
		fmt.Sprintf(`GRANT USAGE ON SCHEMA %s TO %s;`, constants.LegacyCommandSchema, constants.DatabaseUsersRole),
		fmt.Sprintf("IMPORT FOREIGN SCHEMA \"%s\" FROM SERVER steampipe INTO %s;", constants.LegacyCommandSchema, constants.LegacyCommandSchema),
		fmt.Sprintf("GRANT INSERT ON %s.%s TO %s;", constants.LegacyCommandSchema, constants.LegacyCommandTableCache, constants.DatabaseUsersRole),
		fmt.Sprintf("GRANT SELECT ON %s.%s TO %s;", constants.LegacyCommandSchema, constants.LegacyCommandTableScanMetadata, constants.DatabaseUsersRole),
	}
	queries = append(queries, getFunctionAddStrings(db_common.Functions)...)
	if _, err := ExecuteSqlInTransaction(ctx, conn, queries...); err != nil {
		return sperr.WrapWithMessage(err, "failed to initialise functions")
	}

	return nil
}

func getFunctionAddStrings(functions []db_common.SQLFunction) []string {
	var addStrings []string
	for _, function := range functions {
		addStrings = append(addStrings, getFunctionAddString(function))
	}
	return addStrings
}

func getFunctionAddString(function db_common.SQLFunction) string {
	if err := validateFunction(function); err != nil {
		// panic - this should never happen,
		// since the function definitions are
		// tightly bound to development
		panic(err)
	}

	var inputParams []string
	for argName, argType := range function.Params {
		inputParams = append(inputParams, fmt.Sprintf("%s %s", argName, argType))
	}

	return strings.TrimSpace(fmt.Sprintf(
		`
;CREATE OR REPLACE FUNCTION %s.%s (%s) RETURNS %s LANGUAGE %s AS
$$
%s
$$;
`,
		constants.InternalSchema,
		function.Name,
		strings.Join(inputParams, ","),
		function.Returns,
		function.Language,
		strings.TrimSpace(function.Body),
	))
}

func validateFunction(f db_common.SQLFunction) error {
	return nil
}

func initializeConnectionStateTable(ctx context.Context, conn *pgx.Conn) error {

	createQueries := []db_common.QueryWithArgs{
		connection_state.GetConnectionStateTableCreateSql(),
		connection_state.GetConnectionStateTableGrantSql(),
	}

	// apply any alterations
	// this is for changes which are applied to the table after v0.20.0
	createQueries = append(createQueries, connection_state.GetConnectionStateTableColumnAlterSql()...)

	if _, err := ExecuteSqlWithArgsInTransaction(ctx, conn, createQueries...); err != nil {
		return err
	}

	// now load the state
	connectionStateMap, err := steampipeconfig.LoadConnectionState(ctx, conn)
	if err != nil {
		return err
	}

	// if any connections are in a ready  state, set them to pending - we need to run refresh connections before we know this connection is still valid
	pendingSql := connection_state.GetReadConnectionStatePendingSql()
	// if any connections are not in a ready or error state, set them to pending_incomplete
	incompleteErrorSql := connection_state.GetIncompleteConnectionStatePendingIncompleteSql()
	queries := []db_common.QueryWithArgs{
		incompleteErrorSql,
		pendingSql,
	}

	// for any connection in the connection config but not in the connection state table,
	// add an entry with `pending_incomplete` state this is to work around the race condition
	// where we wait for connection state before RefreshConnections has added
	// any new connections into the state table
	for connection, connectionConfig := range steampipeconfig.GlobalConfig.Connections {
		if _, ok := connectionStateMap[connection]; !ok {
			queries = append(queries, connection_state.GetNewConnectionStateTableInsertSql(connectionConfig))
		}
	}

	_, err = ExecuteSqlWithArgsInTransaction(ctx, conn, queries...)
	return err
}
