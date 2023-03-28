package db_local

import (
	"context"
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/schema"
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
		fmt.Sprintf(`create schema if not exists %s;`, constants.InternalSchema),
		fmt.Sprintf(`grant usage on schema %s to %s;`, constants.InternalSchema, constants.DatabaseUsersRole),
		// create connection state table
		fmt.Sprintf(`create table if not exists %s.connection_state (
    			name text primary key,
    			status text,
    			details text,
    			comments_set bool default false,
    			last_change timestamptz);`, constants.InternalSchema),
		// set all existing connections to pending
		fmt.Sprintf(`update %s.%s set status = '%s'`, constants.InternalSchema, constants.ConnectionStateTable, constants.ConnectionStatePending),
		fmt.Sprintf(`grant select on table %s.%s to %s;`, constants.InternalSchema, constants.ConnectionStateTable, constants.DatabaseUsersRole),
	}
	queries = append(queries, getFunctionAddStrings(constants.Functions)...)
	if _, err := executeSqlAsRoot(ctx, queries...); err != nil {
		return sperr.WrapWithMessage(err, "failed to initialise functions")
	}

	return nil
}

func getFunctionAddStrings(functions []schema.SQLFunc) []string {
	addStrings := []string{}
	for _, function := range functions {
		addStrings = append(addStrings, getFunctionAddString(function))
	}
	return addStrings
}

func getFunctionAddString(function schema.SQLFunc) string {
	if err := validateFunction(function); err != nil {
		// panic - this should never happen,
		// since the function definitions are
		// tightly bound to development
		panic(err)
	}

	inputParams := []string{}

	for argName, argType := range function.Params {
		inputParams = append(inputParams, fmt.Sprintf("%s %s", argName, argType))
	}

	return strings.TrimSpace(fmt.Sprintf(
		`
;create or replace function %s.%s (%s) returns %s language %s as
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

func validateFunction(f schema.SQLFunc) error {
	return nil
}
