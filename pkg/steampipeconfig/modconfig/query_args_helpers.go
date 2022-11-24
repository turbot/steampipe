package modconfig

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/type_conversion"
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
	if namedArgCount := len(mergedArgs.ArgMap); namedArgCount > 0 {
		// if named args are provided and the query does not define params, we cannot resolve the args
		if len(source.GetParams()) == 0 {
			log.Printf("[TRACE] %s defines %d named %s but has no parameters definitions", source.Name(), namedArgCount, utils.Pluralize("arg", namedArgCount))
		} else {
			// do params contain named params?
			paramVals, missingParams, err = mergedArgs.resolveNamedParameters(source)
		}
	} else {
		// resolve as positional parameters
		// (or fall back to defaults if no positional params are present)
		paramVals, missingParams, err = mergedArgs.resolvePositionalParameters(source)
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

	// convert any array args into a strongly typed array
	for i, v := range paramVals {
		paramVals[i] = type_conversion.AnySliceToTypedSlice(v)
	}

	// success!
	return paramVals, nil
}
