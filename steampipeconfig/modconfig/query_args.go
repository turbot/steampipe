package modconfig

import (
	"fmt"
	"strings"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
)

// QueryArgs is a struct which contains the arguments used to invoke a prepared statement
// these may either be passed by name, in a map, or as a list of positional args
// NOTE: if both are present the named parameters are used
type QueryArgs struct {
	Args       map[string]string    `cty:"args" json:"args"`
	ArgsList   []string             `cty:"args_list" json:"args_list"`
	References []*ResourceReference `cty:"refs" json:"refs"`
}

func (q *QueryArgs) String() string {
	if q == nil {
		return "<nil>"
	}
	if len(q.ArgsList) > 0 {
		return fmt.Sprintf("Args list: %s", strings.Join(q.ArgsList, ","))
	}
	if len(q.Args) > 0 {
		var strs = make([]string, len(q.Args))
		idx := 0
		for k, v := range q.Args {
			strs[idx] = fmt.Sprintf("%s = %s", k, v)
			idx++
		}
		return fmt.Sprintf("args:\n\t%s", strings.Join(strs, "\n\t"))
	}
	return "<empty>"
}

func NewQueryArgs() *QueryArgs {
	return &QueryArgs{
		Args: make(map[string]string),
	}
}

func (q *QueryArgs) Equals(other *QueryArgs) bool {
	if other == nil {
		return false
	}
	if q.Empty() {
		return other.Empty()
	}
	if len(other.Args) != len(q.Args) || len(other.ArgsList) != len(q.ArgsList) {
		return false
	}
	for k, v := range q.Args {
		if other.Args[k] != v {
			return false
		}
	}
	for i, v := range q.ArgsList {
		if other.ArgsList[i] != v {
			return false
		}
	}
	return true
}

func (q *QueryArgs) Empty() bool {
	return len(q.Args)+len(q.ArgsList) == 0
}

// ResolveAsString resolves the argument values,
// falling back on defaults from param definitions in the source (if present)
// it returns the arg values as a csv string which can be used in a prepared statement invocation
// (the arg values and param defaults will already have been converted to postgres format)
func (q *QueryArgs) ResolveAsString(source QueryProvider) (string, error) {
	var paramStrs, missingParams []string
	var err error
	if len(q.Args) > 0 {
		// do params contain named params?
		paramStrs, missingParams, err = q.resolveNamedParameters(source)
	} else {
		// resolve as positional parameters
		// (or fall back to defaults if no positional params are present)
		paramStrs, missingParams, err = q.resolvePositionalParameters(source)
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
	return fmt.Sprintf("(%s)", strings.Join(paramStrs, ",")), err
}

func (q *QueryArgs) resolveNamedParameters(source QueryProvider) (argStrs []string, missingParams []string, err error) {
	// if query params contains both positional and named params, error out
	if len(q.ArgsList) > 0 {
		err = fmt.Errorf("ResolveAsString failed for %s - params data contain both positional and named parameters", source.Name())
		return
	}
	params := source.GetParams()
	// so params contain named params - if this query has no param defs, error out
	if len(params) < len(q.Args) {
		err = fmt.Errorf("ResolveAsString failed for %s - params data contain %d named parameters but this query %d parameter definitions",
			source.Name(), len(q.Args), len(source.GetParams()))
		return
	}

	// to get here, we must have param defs for all provided named params
	argStrs = make([]string, len(params))

	// iterate through each param def and resolve the value
	// build a map of which args have been matched (used to validate all args have poaram defs)
	argsWithParamDef := make(map[string]bool)
	for i, param := range params {
		defaultValue := typehelpers.SafeString(param.Default)

		// can we resolve a value for this param?
		if val, ok := q.Args[param.Name]; ok {
			argStrs[i] = val
			argsWithParamDef[param.Name] = true
		} else if defaultValue != "" {
			argStrs[i] = defaultValue
		} else {
			// no value provided and no default defined - add to missing list
			missingParams = append(missingParams, param.Name)
		}
	}

	// verify we have param defs for all provided args
	for arg := range q.Args {
		if _, ok := argsWithParamDef[arg]; !ok {
			return nil, nil, fmt.Errorf("no parameter definition found for argument '%s'", arg)
		}
	}

	return argStrs, missingParams, nil
}

func (q *QueryArgs) resolvePositionalParameters(source QueryProvider) (argStrs []string, missingParams []string, err error) {
	// if query params contains both positional and named params, error out
	if len(q.Args) > 0 {
		err = fmt.Errorf("resolvePositionalParameters failed for %s - args data contain both positional and named parameters", source.Name())
		return
	}
	params := source.GetParams()
	// if no param defs are defined, just use the given values
	if len(params) == 0 {
		argStrs = q.ArgsList
		return
	}

	// so there are param defs - we must be able to resolve all params
	// if there are MORE defs than provided parameters, all remaining defs MUST provide a default
	argStrs = make([]string, len(params))

	for i, param := range params {
		defaultValue := typehelpers.SafeString(param.Default)

		if i < len(q.ArgsList) {
			argStrs[i] = q.ArgsList[i]
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
