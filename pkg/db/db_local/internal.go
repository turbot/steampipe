package db_local

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/connection_state"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
)

/**

Query to get functions:
SELECT
    p.proname AS function_name
FROM
    pg_proc p
    LEFT JOIN pg_namespace n ON p.pronamespace = n.oid
WHERE
    n.nspname = 'functionSchema'
ORDER BY
    function_name;

**/

func setupInternal(ctx context.Context, conn *pgx.Conn) error {
	utils.LogTime("db.setupInternal start")
	defer utils.LogTime("db.setupInternal end")

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
