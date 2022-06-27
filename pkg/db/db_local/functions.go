package db_local

import (
	"context"
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/schema"
	"github.com/turbot/steampipe/pkg/utils"
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

func refreshFunctions(ctx context.Context) error {
	utils.LogTime("db.refreshFunctions start")
	defer utils.LogTime("db.refreshFunctions end")

	queries := []string{
		fmt.Sprintf(`create schema if not exists %s;`, constants.FunctionSchema),
		fmt.Sprintf(`grant usage on schema %s to %s;`, constants.FunctionSchema, constants.DatabaseUsersRole),
	}
	queries = append(queries, getFunctionAddStrings(constants.Functions)...)
	if _, err := executeSqlAsRoot(ctx, queries...); err != nil {
		return err
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
		constants.FunctionSchema,
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
