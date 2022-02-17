package modconfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

type QueryProviderBase struct {
	preparedStatementName string
}

// GetPreparedStatementExecuteSQL return the SQLs to run the query as a prepared statement
func (p *QueryProviderBase) GetPreparedStatementExecuteSQL(args *QueryArgs) (string, error) {
	paramsString, err := args.ResolveAsString(p.queryProvider)
	if err != nil {
		return "", fmt.Errorf("failed to resolve args for %s: %s", p.queryProvider.Name(), err.Error())
	}
	executeString := fmt.Sprintf("execute %s%s", p.buildPreparedStatementName(), paramsString)
	log.Printf("[TRACE] GetPreparedStatementExecuteSQL source: %s, sql: %s, args: %s", p.queryProvider.Name(), executeString, args)
	return executeString, nil
}

func (p *QueryProviderBase) buildPreparedStatementName(modName, suffix, name string) string {
	// build prefix from mod name
	prefix := p.buildPreparedStatementPrefix(modName)

	// build the hash from the query/control name, mod name and suffix and take the first 4 bytes
	str := fmt.Sprintf("%s%s%s", prefix, name, suffix)
	hash := utils.GetMD5Hash(str)[:4]
	// add hash to suffix
	suffix += hash

	// truncate the name if necessary
	nameLength := len(name)
	maxNameLength := constants.MaxPreparedStatementNameLength - (len(prefix) + len(suffix))
	if nameLength > maxNameLength {
		nameLength = maxNameLength
	}

	// construct the name
	p.preparedStatementName = fmt.Sprintf("%s%s%s", prefix, name[:nameLength], suffix)
	return p.preparedStatementName
}

// set the prepared statement suffix and prefix
// and also store the parent resource object as a QueryProvider interface (base struct cannot cast itself to this)
func (p *QueryProviderBase) buildPreparedStatementPrefix(modName string) string {
	prefix := fmt.Sprintf("%s_", modName)
	prefix = strings.Replace(prefix, ".", "_", -1)
	prefix = strings.Replace(prefix, "@", "_", -1)

	return prefix
}
