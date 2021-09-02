package modconfig

import (
	"fmt"
	"log"
)

const preparesStatementQuerySuffix = "_psq"
const preparesStatementControlSuffix = "_psc"

// GetPreparedStatementExecuteSQL return the SQLs to run the query as a prepared statement
func GetPreparedStatementExecuteSQL(source PreparedStatementProvider, args *QueryArgs) (string, error) {
	paramsString, err := args.ResolveAsString(source)
	if err != nil {
		return "", fmt.Errorf("failed to resolve args for %s: %s", source.Name(), err.Error())
	}
	executeString := fmt.Sprintf("execute %s%s", source.PreparedStatementName(), paramsString)
	log.Printf("[TRACE] GetPreparedStatementExecuteSQL source: %s, sql: %s, args: %s", source.Name(), executeString, args)
	return executeString, nil
}

// return the prepared statement name for the given source
func preparedStatementName(source PreparedStatementProvider) string {
	var name, suffix string

	switch t := source.(type) {
	case *Query:
		name = t.ShortName
		suffix = preparesStatementQuerySuffix
	case *Control:
		name = t.ShortName
		suffix = preparesStatementControlSuffix
	}
	maxNameLength := 64 - len(suffix)
	nameLength := len(name)
	if nameLength > maxNameLength {
		nameLength = maxNameLength
	}
	return fmt.Sprintf("%s%s", name[:nameLength], suffix)
}
