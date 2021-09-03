package modconfig

import (
	"fmt"
	"log"

	"github.com/turbot/steampipe/utils"
)

const maxPreparedStatementNameLength = 63
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
	preparedStatementName := name + suffix

	// is this name too long?
	if len(preparedStatementName) > maxPreparedStatementNameLength {
		// if the name is longer than the max length, truncate it and add a truncated hash
		// NOTE: as we are truncating the hash there is a theoretical possibility of name clash
		// however as this only applies for very long control/query names, it's considered an acceptable risk
		// NOTE 2: hash the name WITH the suffix to avoid clashed between controls and queries with the same long name
		suffix = fmt.Sprintf("_%s", utils.GetMD5Hash(name + suffix)[:8])

		// rebuild the name -  determine how much of the original name to include
		nameLength := 63 - len(suffix)
		preparedStatementName = fmt.Sprintf("%s%s", name[:nameLength], suffix)
	}
	return preparedStatementName
}
