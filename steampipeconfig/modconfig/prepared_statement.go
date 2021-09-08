package modconfig

import (
	"fmt"
	"log"

	"github.com/turbot/steampipe/utils"
)

const maxPreparedStatementNameLength = 63
const preparesStatementQuerySuffix = "_q"
const preparesStatementControlSuffix = "_c"

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

func preparedStatementName(source PreparedStatementProvider) string {
	var name, suffix string
	prefix := fmt.Sprintf("%s_", source.ModName())

	// build the hash of the source object and take first 4 bytes
	str := fmt.Sprintf("%v", source)
	hash := utils.GetMD5Hash(str)[:4]

	switch t := source.(type) {
	case *Query:
		name = t.ShortName
		suffix = preparesStatementQuerySuffix + hash
	case *Control:
		name = t.ShortName
		suffix = preparesStatementControlSuffix + hash
	}

	nameLength := len(name)
	maxNameLength := maxPreparedStatementNameLength - (len(prefix) + len(suffix))
	if nameLength > maxNameLength {
		nameLength = maxNameLength
	}

	preparedStatementName := fmt.Sprintf("%s%s%s", prefix, name[:nameLength], suffix)

	return preparedStatementName
}
