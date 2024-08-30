package modconfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
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
// it returns the arg values as a csv string which can be used in a query invocation
// (the arg values and param defaults will already have been converted to postgres format)
func ResolveArgs(qp QueryProvider, runtimeArgs *QueryArgs) ([]any, error) {
	var argVals []any
	var missingParams []string
	var err error
	// validate args
	if runtimeArgs == nil {
		runtimeArgs = &QueryArgs{}
	}

	log.Printf("[TRACE] ResolveArgs: resolving args for %s", qp.Name())

	// merge the query provider args (if any) with the runtime args
	sourceArgs := qp.GetArgs()
	if sourceArgs == nil {
		sourceArgs = &QueryArgs{}
	}
	mergedArgs, err := sourceArgs.Merge(runtimeArgs, qp)
	if err != nil {
		log.Printf("[WARN] ResolveArgs failed to merge args for %s: %s", qp.Name(), err.Error())
		return nil, err
	}

	if namedArgCount := len(mergedArgs.ArgMap); namedArgCount > 0 {
		log.Printf("[TRACE] %s defines %d named %s", qp.Name(), namedArgCount, utils.Pluralize("arg", namedArgCount))
		// if named args are provided and the query does not define params, we cannot resolve the args
		if len(qp.GetParams()) == 0 {
			log.Printf("[TRACE] %s defines %d named %s but has no parameters definitions", qp.Name(), namedArgCount, utils.Pluralize("arg", namedArgCount))
		} else {
			// do params contain named params?
			argVals, missingParams, err = mergedArgs.resolveNamedParameters(qp)
			log.Printf("[TRACE] resolved %d named %s for %s", len(argVals), utils.Pluralize("params", len(argVals)), qp.Name())
		}
	} else {
		// resolve as positional parameters
		// (or fall back to defaults if no positional params are present)
		argVals, missingParams, err = mergedArgs.resolvePositionalParameters(qp)
		log.Printf("[TRACE] resolved %d positional %s for %s", len(argVals), utils.Pluralize("params", len(argVals)), qp.Name())
	}
	if err != nil {
		log.Printf("[WARN] ResolveArgs failed to resolve args for %s: %s", qp.Name(), err.Error())
		return nil, err
	}

	// did we resolve them all?
	if len(missingParams) > 0 {
		log.Printf("[WARN] ResolveArgs: args missing for %s: %s", qp.Name(), strings.Join(missingParams, ","))
		// a better error will be constructed by the calling code
		return nil, fmt.Errorf("%s", strings.Join(missingParams, ","))
	}

	// are there any params?
	if len(argVals) == 0 {
		return nil, nil
	}

	// convert any array args into a strongly typed array
	for i, v := range argVals {
		argVals[i] = helpers.AnySliceToTypedSlice(v)
	}

	// success!
	return argVals, nil
}
