package modconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/turbot/steampipe/pkg/utils"
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

	argVals = make([]any, len(params))

	// iterate through each param def and resolve the value
	// build a map of which args have been matched (used to validate all args have param defs)
	argsWithParamDef := make(map[string]bool)
	for i, param := range params {
		// first set default
		var defaultValue any = nil
		if param.Default == nil {
			defaultValue = ""
		} else {
			if param.Default != nil {
				err := json.Unmarshal([]byte(*param.Default), &defaultValue)
				if err != nil {
					return nil, nil, err
				}
			}
		}
		// can we resolve a value for this param?
		if val, ok := args.ArgMap[param.Name]; ok {
			// convert from json
			var argVal any
			err := json.Unmarshal([]byte(val), &argVal)
			if err != nil {
				return nil, nil, err
			}
			argVals[i] = argVal
			argsWithParamDef[param.Name] = true

		} else if defaultValue != nil {
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
			log.Printf("[TRACE] no parameter definition found for argument '%s'", arg)
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

	// so there are param definitions - use these to populate argStrs

	if len(params) < len(args.ArgList) {
		err = fmt.Errorf("resolvePositionalParameters failed for '%s' - %d %s were provided but there %s %d parameter %s",
			queryProvider.Name(),
			len(args.ArgList),
			utils.Pluralize("argument", len(args.ArgList)),
			utils.Pluralize("is", len(params)),
			len(params),
			utils.Pluralize("definition", len(params)),
		)
		return
	}

	// so there are param definitions - use these to populate argStrs
	argValues = make([]any, len(params))

	for i, param := range params {
		// first set default
		var defaultValue any = nil
		if param.Default != nil {
			err := json.Unmarshal([]byte(*param.Default), &defaultValue)
			if err != nil {
				return nil, nil, err
			}
		}

		if i < len(args.ArgList) && args.ArgList[i] != nil {
			// convert from json
			var argVal any
			err := json.Unmarshal([]byte(*args.ArgList[i]), &argVal)
			if err != nil {
				return nil, nil, err
			}

			argValues[i] = argVal
		} else if defaultValue != nil {
			// so we have run out of provided params - is there a default?
			argValues[i] = defaultValue
		} else {
			// no value provided and no default defined - add to missing list
			missingParams = append(missingParams, param.Name)
		}
	}
	return argValues, missingParams, nil
}
