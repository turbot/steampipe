package modconfig

import (
	"fmt"
	"log"
	"strings"

	typehelpers "github.com/turbot/go-kit/types"

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
	paramsString, err := queryProvider.ResolveArgsAsString(queryProvider, runtimeArgs)
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

// ResolveArgsAsString resolves the argument values,
// falling back on defaults from param definitions in the source (if present)
// it returns the arg values as a csv string which can be used in a prepared statement invocation
// (the arg values and param defaults will already have been converted to postgres format)
func (b *QueryProviderBase) ResolveArgsAsString(source QueryProvider, runtimeArgs *QueryArgs) (string, error) {
	var paramStrs, missingParams []string

	// validate args
	mergedArgs, err := b.mergeArgs(source.GetArgs(), runtimeArgs, source)
	if err != nil {
		return "", err
	}

	if len(mergedArgs.ArgMap) > 0 {
		// do params contain named params?
		paramStrs, missingParams, err = b.resolveNamedParameters(source, mergedArgs)
	} else {
		// resolve as positional parameters
		// (or fall back to defaults if no positional params are present)
		paramStrs, missingParams, err = b.resolvePositionalParameters(source, mergedArgs)
	}
	if err != nil {
		return "", err
	}

	// did we resolve them all?
	if len(missingParams) > 0 {
		return "", fmt.Errorf("ResolveAsString failed for %s - failed to resolve value for %d %s: %s",
			source.Name(),
			len(missingParams),
			utils.Pluralize("parameter", len(missingParams)),
			strings.Join(missingParams, ","))
	}

	// are there any params?
	if len(paramStrs) == 0 {
		return "", nil
	}

	// success!
	return fmt.Sprintf("(%s)", strings.Join(paramStrs, ",")), nil
}

// TODO [report] do not support runtime default and convert this to merge
// validate base and runtime args and merge into single args
func (b *QueryProviderBase) mergeArgs(baseArgs, runtimeArgs *QueryArgs, source QueryProvider) (*QueryArgs, error) {
	// ensure non nil
	if baseArgs == nil {
		baseArgs = NewQueryArgs()
	}
	if runtimeArgs == nil {
		runtimeArgs = NewQueryArgs()
	}

	return baseArgs.Merge(runtimeArgs, source)
}

func (b *QueryProviderBase) resolveNamedParameters(source QueryProvider, args *QueryArgs) (argStrs []string, missingParams []string, err error) {
	// if query params contains both positional and named params, error out
	params := source.GetParams()

	// so params contain named params - if this query has no param defs, error out
	if len(params) < len(args.ArgMap) {
		err = fmt.Errorf("resolveNamedParameters failed for '%s' - %d named arguments were provided but there are %d parameter definitions",
			source.Name(), len(args.ArgMap), len(source.GetParams()))
		return
	}

	// to get here, we must have param defs for all provided named params
	argStrs = make([]string, len(params))

	// iterate through each param def and resolve the value
	// build a map of which args have been matched (used to validate all args have param defs)
	argsWithParamDef := make(map[string]bool)
	for i, param := range params {
		// first set default
		defaultValue := typehelpers.SafeString(param.Default)

		// can we resolve a value for this param?
		if val, ok := args.ArgMap[param.Name]; ok {
			argStrs[i] = val
			argsWithParamDef[param.Name] = true

		} else if defaultValue != "" {
			// is there a default
			argStrs[i] = defaultValue
		} else {
			// no value provided and no default defined - add to missing list
			missingParams = append(missingParams, param.Name)
		}
	}

	// verify we have param defs for all provided args
	for arg := range args.ArgMap {
		if _, ok := argsWithParamDef[arg]; !ok {
			return nil, nil, fmt.Errorf("no parameter definition found for argument '%s'", arg)
		}
	}

	return argStrs, missingParams, nil
}

func (b *QueryProviderBase) resolvePositionalParameters(source QueryProvider, args *QueryArgs) (argStrs []string, missingParams []string, err error) {
	// if query params contains both positional and named params, error out

	// if there are param defs - we must be able to resolve all params
	// if there are MORE defs than provided parameters, all remaining defs MUST provide a default
	params := source.GetParams()

	// if no param defs are defined, just use the given values, using runtime dependencies where available
	if len(params) == 0 {

		// no params defined, so we return as many args as are provided
		// (convert from *string to string)
		argStrs = args.ArgsStringList()
		return
	}

	// so there are param defintions - use these to populate argStrs
	argStrs = make([]string, len(params))

	for i, param := range params {
		// first set default
		defaultValue := typehelpers.SafeString(param.Default)

		if i < len(args.ArgList) && args.ArgList[i] != nil {
			argStrs[i] = typehelpers.SafeString(args.ArgList[i])
		} else if defaultValue != "" {
			// so we have run out of provided params - is there a default?
			argStrs[i] = defaultValue
		} else {
			// no value provided and no default defined - add to missing list
			missingParams = append(missingParams, param.Name)
		}

	}
	return
}
