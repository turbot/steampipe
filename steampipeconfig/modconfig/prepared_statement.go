package modconfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/turbot/steampipe/utils"
)

const maxPreparedStatementNameLength = 63
const preparesStatementQuerySuffix = "_q"
const preparesStatementControlSuffix = "_c"

// GetPreparedStatementExecuteSQL return the SQLs to run the query as a prepared statement
func GetPreparedStatementExecuteSQL(source QueryProvider, args *QueryArgs) (string, error) {
	paramsString, err := args.ResolveAsString(source)
	if err != nil {
		return "", fmt.Errorf("failed to resolve args for %s: %s", source.Name(), err.Error())
	}
	executeString := fmt.Sprintf("execute %s%s", source.GetPreparedStatementName(), paramsString)
	log.Printf("[TRACE] GetPreparedStatementExecuteSQL source: %s, sql: %s, args: %s", source.Name(), executeString, args)
	return executeString, nil
}

func preparedStatementName(source QueryProvider) string {
	var name, suffix string
	prefix := fmt.Sprintf("%s_", source.GetModName())
	prefix = strings.Replace(prefix, ".", "_", -1)
	prefix = strings.Replace(prefix, "@", "_", -1)

	// build suffix using a char to indicate control or query, and the truncated hash
	switch t := source.(type) {
	case *Query:
		name = t.ShortName
		suffix = preparesStatementQuerySuffix
	case *Control:
		name = t.ShortName
		suffix = preparesStatementControlSuffix
	}
	// build the hash from the query/control name, mod name and suffix and take the first 4 bytes
	str := fmt.Sprintf("%s%s%s", prefix, name, suffix)
	hash := utils.GetMD5Hash(str)[:4]
	// add hash to suffix
	suffix += hash

	// truncate the name if necessary
	nameLength := len(name)
	maxNameLength := maxPreparedStatementNameLength - (len(prefix) + len(suffix))
	if nameLength > maxNameLength {
		nameLength = maxNameLength
	}

	// construct the name
	preparedStatementName := fmt.Sprintf("%s%s%s", prefix, name[:nameLength], suffix)
	return preparedStatementName
}
