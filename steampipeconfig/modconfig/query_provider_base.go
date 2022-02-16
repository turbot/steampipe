package modconfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

type QueryProviderBase struct {
	// the query provider which we are the base of
	queryProvider           QueryProvider
	preparedStatementSuffix string
	preparedStatementPrefix string
	preparedStatementName   string
}

// GetPreparedStatementExecuteSQL return the SQLs to run the query as a prepared statement
func (p *QueryProviderBase) GetPreparedStatementExecuteSQL(args *QueryArgs) (string, error) {
	paramsString, err := args.ResolveAsString(p.queryProvider)
	if err != nil {
		return "", fmt.Errorf("failed to resolve args for %s: %s", p.queryProvider.Name(), err.Error())
	}
	executeString := fmt.Sprintf("execute %s%s", p.GetPreparedStatementName(), paramsString)
	log.Printf("[TRACE] GetPreparedStatementExecuteSQL source: %s, sql: %s, args: %s", p.queryProvider.Name(), executeString, args)
	return executeString, nil
}

func (p *QueryProviderBase) initPreparedStatementName(queryProvider QueryProvider, modName, suffix string) {
	prefix := fmt.Sprintf("%s_", modName)
	prefix = strings.Replace(prefix, ".", "_", -1)
	prefix = strings.Replace(prefix, "@", "_", -1)

	p.preparedStatementPrefix = prefix
	p.preparedStatementSuffix = suffix
}

func (p *QueryProviderBase) GetPreparedStatementName() string {
	if p.preparedStatementName != "" {
		return p.preparedStatementName
	}
	var name string
	prefix := p.preparedStatementPrefix
	suffix := p.preparedStatementSuffix

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
