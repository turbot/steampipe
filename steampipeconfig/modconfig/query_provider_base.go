package modconfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

type QueryProviderBase struct {
	runtimeDependencies map[string]*RuntimeDependency
}

// VerifyQuery returns an error if neither sql or query are set
// it is overidden by resource types for which sql is optional
func (b *QueryProviderBase) VerifyQuery(queryProvider QueryProvider) error {
	// verify we have either SQL or a Query defined
	if queryProvider.GetQuery() == nil && queryProvider.GetSQL() == nil {
		// this should never happen as we should catch it in the parsing stage
		return fmt.Errorf("%s must define either a 'sql' property or a 'query' property", queryProvider.Name())
	}
	return nil
}

// IsParameterised returns whether the query provider has a parameterised query
// the query is parameterised if either there are any param defintions, or any positional arguments passed,
// or it has runtime dependencies (which must be args)
func (b *QueryProviderBase) IsParameterised(baseArgs *QueryArgs, params []*ParamDef) bool {
	return baseArgs != nil && len(baseArgs.ArgList) > 0 ||
		len(b.runtimeDependencies) > 0 ||
		len(params) > 0
}

func (b *QueryProviderBase) RequiresExecution(queryProvider QueryProvider) bool {
	return queryProvider.GetQuery() != nil || queryProvider.GetSQL() != nil
}

func (b *QueryProviderBase) buildPreparedStatementName(queryName, modName, suffix string) string {
	// build prefix from mod name
	prefix := b.buildPreparedStatementPrefix(modName)

	// build the hash from the query/control name, mod name and suffix and take the first 4 bytes
	str := fmt.Sprintf("%s%s%s", prefix, queryName, suffix)
	hash := utils.GetMD5Hash(str)[:4]
	// add hash to suffix
	suffix += hash

	// truncate the name if necessary
	nameLength := len(queryName)
	maxNameLength := constants.MaxPreparedStatementNameLength - (len(prefix) + len(suffix))
	if nameLength > maxNameLength {
		nameLength = maxNameLength
	}

	// construct the name
	return fmt.Sprintf("%s%s%s", prefix, queryName[:nameLength], suffix)
}

// set the prepared statement suffix and prefix
// and also store the parent resource object as a QueryProvider interface (base struct cannot cast itself to this)
func (b *QueryProviderBase) buildPreparedStatementPrefix(modName string) string {
	prefix := fmt.Sprintf("%s_", modName)
	prefix = strings.Replace(prefix, ".", "_", -1)
	prefix = strings.Replace(prefix, "@", "_", -1)

	return prefix
}

// return the SQLs to run the query as a prepared statement
func (b *QueryProviderBase) getPreparedStatementExecuteSQL(queryProvider QueryProvider, runtimeArgs *QueryArgs) (string, error) {
	paramsString, err := ResolveArgsAsString(queryProvider, runtimeArgs)
	if err != nil {
		return "", fmt.Errorf("failed to resolve args for %s: %s", queryProvider.Name(), err.Error())
	}
	executeString := fmt.Sprintf("execute %s%s", queryProvider.GetPreparedStatementName(), paramsString)
	log.Printf("[TRACE] GetPreparedStatementExecuteSQL source: %s, sql: %s, args: %s", queryProvider.Name(), executeString, runtimeArgs)
	return executeString, nil
}

func (b *QueryProviderBase) AddRuntimeDependencies(dependencies []*RuntimeDependency) {
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}
	for _, dependency := range dependencies {
		b.runtimeDependencies[dependency.String()] = dependency
	}
}

func (b *QueryProviderBase) MergeRuntimeDependencies(other QueryProvider) {
	dependencies := other.GetRuntimeDependencies()
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}
	for _, dependency := range dependencies {
		if _, ok := b.runtimeDependencies[dependency.String()]; !ok {
			b.runtimeDependencies[dependency.String()] = dependency
		}
	}
}

func (b *QueryProviderBase) GetRuntimeDependencies() map[string]*RuntimeDependency {
	return b.runtimeDependencies
}
