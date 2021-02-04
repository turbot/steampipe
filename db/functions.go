package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
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

func refreshFunctions(client *Client) error {
	sql := []string{
		fmt.Sprintf(`create schema if not exists %s;`, constants.FunctionSchema),
		fmt.Sprintf(`grant usage on schema %s to steampipe_users;`, constants.FunctionSchema),
	}
	sql = append(sql, getFunctionAddStrings(constants.Functions)...)
	if err := executeAddFunctionQuery(strings.Join(sql, ";")); err != nil {
		// panic - this should never happen,
		// since the function definitions are
		// tightly bound to development
		panic(err)
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

func executeAddFunctionQuery(functionQueryString string) error {
	log.Println("[TRACE]", "executeAddFunctionQuery", "executing", functionQueryString)
	client, err := createSteampipeRootDbClient()
	if err != nil {
		return err
	}
	defer func() {
		client.Close()
	}()
	_, err = client.Exec(functionQueryString)
	if err != nil {
		return err
	}
	return nil
}
