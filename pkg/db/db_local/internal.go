package db_local

import (
	"context"
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
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

func setupInternal(ctx context.Context) error {
	utils.LogTime("db.setupInternal start")
	defer utils.LogTime("db.setupInternal end")

	queries := []string{
		"lock table pg_namespace;",
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, constants.InternalSchema),
		fmt.Sprintf(`GRANT USAGE ON SCHEMA %s TO %s;`, constants.InternalSchema, constants.DatabaseUsersRole),
		// create connection state table
		getConnectionStateTableCreateSql(),
		// set all existing connections to pending
		fmt.Sprintf(`UPDATE %s.%s SET STATE = '%s'`, constants.InternalSchema, constants.ConnectionStateTable, constants.ConnectionStatePending),
		fmt.Sprintf(`GRANT SELECT ON TABLE %s.%s to %s;`, constants.InternalSchema, constants.ConnectionStateTable, constants.DatabaseUsersRole),
	}
	queries = append(queries, getFunctionAddStrings(db_common.Functions)...)
	if _, err := executeSqlAsRoot(ctx, queries...); err != nil {
		return sperr.WrapWithMessage(err, "failed to initialise functions")
	}

	return nil
}

func getConnectionStateTableCreateSql() string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
    			name TEXT PRIMARY KEY,
-- 			    connection_type TEXT,
-- 			    child_connections TEXT[],
    			state TEXT NOT NULL,
    			error TEXT NULL,
    			plugin TEXT NOT NULL,
    			schema_mode TEXT NOT NULL,
    			schema_hash TEXT NULL,
    			comments_set BOOL DEFAULT FALSE,
    			connection_mod_time TIMESTAMPTZ NOT NULL,
    			plugin_mod_time TIMESTAMPTZ NOT NULL
    			);`, constants.InternalSchema, constants.ConnectionStateTable)
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
