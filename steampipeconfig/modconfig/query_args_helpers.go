package modconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/utils"

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

// ResolveArgsAsString resolves the argument values,
// falling back on defaults from param definitions in the source (if present)
// it returns the arg values as a csv string which can be used in a prepared statement invocation
// (the arg values and param defaults will already have been converted to postgres format)
func ResolveArgsAsString(source QueryProvider, runtimeArgs *QueryArgs) (string, error) {
	var paramStrs, missingParams []string
	var err error
	// validate args

	if len(runtimeArgs.ArgMap) > 0 {
		// do params contain named params?
		paramStrs, missingParams, err = resolveNamedParameters(source, runtimeArgs)
	} else {
		// resolve as positional parameters
		// (or fall back to defaults if no positional params are present)
		paramStrs, missingParams, err = resolvePositionalParameters(source, runtimeArgs)
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

func resolveNamedParameters(queryProvider QueryProvider, args *QueryArgs) (argStrs []string, missingParams []string, err error) {
	// if query params contains both positional and named params, error out
	params := queryProvider.GetParams()

	// so params contain named params - if this query has no param defs, error out
	if len(params) < len(args.ArgMap) {
		err = fmt.Errorf("resolveNamedParameters failed for '%s' - %d named arguments were provided but there are %d parameter definitions",
			queryProvider.Name(), len(args.ArgMap), len(queryProvider.GetParams()))
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

func resolvePositionalParameters(queryProvider QueryProvider, args *QueryArgs) (argStrs []string, missingParams []string, err error) {
	// if query params contains both positional and named params, error out

	// if there are param defs - we must be able to resolve all params
	// if there are MORE defs than provided parameters, all remaining defs MUST provide a default
	params := queryProvider.GetParams()

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
