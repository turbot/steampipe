package modconfig

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
	"strings"

	typehelpers "github.com/turbot/go-kit/types"
)

// MergeArgs ensures base and runtime args are non nil and merges them into single args
func MergeArgs(queryProvider QueryProvider, runtimeArgs *QueryArgs) (*QueryArgs, error) {

	baseArgs := queryProvider.GetArgs()
	// ensure non nil
	if baseArgs == nil {
		baseArgs = NewQueryArgs()
	}
	if runtimeArgs == nil {
		runtimeArgs = NewQueryArgs()
	}

	return baseArgs.Merge(runtimeArgs, queryProvider)
}

// ResolveArgs resolves the argument values,
// falling back on defaults from param definitions in the source (if present)
// it returns the arg values as a csv string which can be used in a prepared statement invocation
// (the arg values and param defaults will already have been converted to postgres format)
func ResolveArgs(source QueryProvider, runtimeArgs *QueryArgs) ([]any, error) {
	var paramVals []any
	var missingParams []string
	var err error
	// validate args
	if runtimeArgs == nil {
		runtimeArgs = &QueryArgs{}
	}

	// merge the query provider args (if any) with the runtime args
	sourceArgs := source.GetArgs()
	if sourceArgs == nil {
		sourceArgs = &QueryArgs{}
	}
	mergedArgs, err := sourceArgs.Merge(runtimeArgs, source)
	if err != nil {
		return nil, err
	}
	if len(mergedArgs.ArgMap) > 0 {
		// do params contain named params?
		paramVals, missingParams, err = resolveNamedParameters(source, mergedArgs)
	} else {
		// resolve as positional parameters
		// (or fall back to defaults if no positional params are present)
		paramVals, missingParams, err = resolvePositionalParameters(source, mergedArgs)
	}
	if err != nil {
		return nil, err
	}

	// did we resolve them all?
	if len(missingParams) > 0 {
		// a better error will be constructed by the calling code
		return nil, fmt.Errorf("%s", strings.Join(missingParams, ","))
	}

	// are there any params?
	if len(paramVals) == 0 {
		return nil, nil
	}

	// success!
	return paramVals, nil
}

func resolveNamedParameters(queryProvider QueryProvider, args *QueryArgs) (argVals []any, missingParams []string, err error) {
	// if query params contains both positional and named params, error out
	params := queryProvider.GetParams()

	// so params contain named params - if this query has no param defs, error out
	if len(params) < len(args.ArgMap) {
		err = fmt.Errorf("resolveNamedParameters failed for '%s' - %d named arguments were provided but there are %d parameter definitions",
			queryProvider.Name(), len(args.ArgMap), len(queryProvider.GetParams()))
		return
	}

	// to get here, we must have param defs for all provided named params
	argVals = make([]any, len(params))

	// iterate through each param def and resolve the value
	// build a map of which args have been matched (used to validate all args have param defs)
	argsWithParamDef := make(map[string]bool)
	for i, param := range params {
		// first set default
		defaultValue := typehelpers.SafeString(param.Default)

		// can we resolve a value for this param?
		if val, ok := args.ArgMap[param.Name]; ok {
			// convert from cty to go
			goVal, err := utils.CtyToGo(val)
			if err != nil {
				return nil, nil, err
			}
			argVals[i] = goVal
			argsWithParamDef[param.Name] = true

		} else if defaultValue != "" {
			// is there a default
			argVals[i] = defaultValue
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

	return argVals, missingParams, nil
}

func resolvePositionalParameters(queryProvider QueryProvider, args *QueryArgs) (argValues []any, missingParams []string, err error) {
	// if query params contains both positional and named params, error out

	// if there are param defs - we must be able to resolve all params
	// if there are MORE defs than provided parameters, all remaining defs MUST provide a default
	params := queryProvider.GetParams()

	// if no param defs are defined, just use the given values, using runtime dependencies where available
	if len(params) == 0 {
		// no params defined, so we return as many args as are provided
		// (convert from *string to string)
		argValues = args.SafeArgsList()
		return argValues, nil, nil
	}

	// so there are param defintions - use these to populate argStrs
	argValues = make([]any, len(params))

	for i, param := range params {
		// first set default
		defaultValue := typehelpers.SafeString(param.Default)

		if i < len(args.ArgList) && args.ArgList[i] != cty.Zero {
			// convert from cty to go
			goVal, err := utils.CtyToGo(args.ArgList[i])
			if err != nil {
				return nil, nil, err
			}

			argValues[i] = goVal
		} else if defaultValue != "" {
			// so we have run out of provided params - is there a default?
			argValues[i] = defaultValue
		} else {
			// no value provided and no default defined - add to missing list
			missingParams = append(missingParams, param.Name)
		}
	}
	return argValues, missingParams, nil
}

// QueryProviderIsParameterised returns whether the query provider has a parameterised query
// the query is parameterised if either there are any param defintions, or any positional arguments passed,
// or it has runtime dependencies (which must be args)
func QueryProviderIsParameterised(queryProvider QueryProvider) bool {
	// no sql, NOT parameterised
	if queryProvider.GetSQL() == nil {
		return false
	}

	args := queryProvider.GetArgs()
	params := queryProvider.GetParams()
	runtimeDependencies := queryProvider.GetRuntimeDependencies()

	return args != nil || len(params) > 0 || len(runtimeDependencies) > 0
}
