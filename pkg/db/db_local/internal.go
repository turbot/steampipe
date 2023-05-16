package db_local

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/connection_state"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
)

// dropLegacyInternal looks for a schema named 'internal'
// which has a function called 'glob' and maybe a table named 'connection_state'
// and drops it
func dropLegacyInternal(ctx context.Context, conn *pgx.Conn) error {
	utils.LogTime("db_local.dropLegacyInternal start")
	defer utils.LogTime("db_local.dropLegacyInternal end")

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

	var fNames string
	var tNames string
	err := row.Scan(&fNames, &tNames)
	if err != nil {
		return sperr.WrapWithMessage(err, "could not query legacy 'internal' schema objects: '%s'", constants.LegacyInternalSchema)
	}

	if len(fNames) == 0 && len(tNames) == 0 {
		log.Println("[TRACE] could not find any objects in 'internal' - skipping drop")
		return nil
	}

	functionNames := strings.Split(fNames, ",")
	tableNames := strings.Split(tNames, ",")

	log.Println("[TRACE] dropLegacyInternal: available function names", functionNames)
	log.Println("[TRACE] dropLegacyInternal: available table names", tableNames)

	expectedFunctions := []string{"glob"}
	expectedTables := []string{"connection_state", constants.ConnectionStateTable}

	unexpectedFunctions := helpers.StringSliceDiff(functionNames, expectedFunctions)
	unexpectedTables := helpers.StringSliceDiff(tableNames, expectedTables)

	if len(unexpectedFunctions) > 0 {
		log.Println("[TRACE] dropLegacyInternal: unexpected functions", unexpectedFunctions)
		log.Println("[INFO] found 'internal' with unexpected functions - skipping drop")
		return nil
	}

	if len(unexpectedTables) > 0 {
		log.Println("[TRACE] dropLegacyInternal: unexpected tables", unexpectedTables)
		log.Println("[INFO] found 'internal' with unexpected tables - skipping drop")
		return nil
	}

	log.Println("[TRACE] dropping legacy 'internal' schema")
	if _, err := conn.Exec(ctx, fmt.Sprintf("DROP SCHEMA %s CASCADE", constants.LegacyInternalSchema)); err != nil {
		return sperr.WrapWithMessage(err, "could not drop legacy schema: '%s'", constants.LegacyInternalSchema)
	}
	log.Println("[TRACE] dropped legacy 'internal' schema")
	return nil
}

func setupInternal(ctx context.Context, conn *pgx.Conn) error {
	utils.LogTime("db_local.setupInternal start")
	defer utils.LogTime("db_local.setupInternal end")

	if err := dropLegacyInternal(ctx, conn); err != nil {
		log.Println("[INFO] failed to drop legacy 'internal' schema", err)
	}

	queries := []string{
		"lock table pg_namespace;",
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, constants.InternalSchema),
		fmt.Sprintf(`GRANT USAGE ON SCHEMA %s TO %s;`, constants.InternalSchema, constants.DatabaseUsersRole),
		fmt.Sprintf("IMPORT FOREIGN SCHEMA \"%s\" FROM SERVER steampipe INTO %s;\n", constants.InternalSchema, constants.InternalSchema),
		fmt.Sprintf("GRANT INSERT ON %s.%s TO %s;", constants.InternalSchema, constants.CommandTableSettings, constants.DatabaseUsersRole),
		fmt.Sprintf("GRANT SELECT ON %s.%s TO %s;", constants.InternalSchema, constants.CommandTableScanMetadata, constants.DatabaseUsersRole),
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
	// first create the table if necessary
	createQueries := []db_common.QueryWithArgs{
		connection_state.GetConnectionStateTableCreateSql(),
		{Query: fmt.Sprintf(`GRANT SELECT ON TABLE %s.%s to %s;`, constants.InternalSchema, constants.ConnectionStateTable, constants.DatabaseUsersRole)},
	}
	if _, err := ExecuteSqlWithArgsInTransaction(ctx, conn, createQueries...); err != nil {
		return err
	}

	// now load the state
	connectionStateMap, err := steampipeconfig.LoadConnectionState(ctx, conn)
	if err != nil {
		return err
	}

	// if any connections are not in a ready or error state, set them to pending_incpomplete
	incompleteErrorSql := connection_state.GetIncompleteConnectionStatePendingIncompleteSql()
	queries := []db_common.QueryWithArgs{
		incompleteErrorSql,
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
