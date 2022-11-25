package modconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// QueryArgs is a struct which contains the arguments used to invoke a prepared statement
// these may either be passed by name, in a map, or as a list of positional args
// NOTE: if both are present the named parameters are used
type QueryArgs struct {
	ArgMap map[string]string `cty:"args" json:"args,omitempty"`
	// args list may be sparsely populated (in case of runtime dependencies)
	// so use *string
	ArgList    []*string            `cty:"args_list" json:"args_list"`
	References []*ResourceReference `cty:"refs" json:"refs"`
	// TACTICAL: map of positional and named args which are strings and therefor do NOT need JSON serialising
	// (can be removed when we move to cty)
	stringNamedArgs      map[string]struct{}
	stringPositionalArgs map[int]struct{}
}

func (q *QueryArgs) String() string {
	if q == nil {
		return "<nil>"
	}
	if len(q.ArgList) > 0 {
		argsStringList := q.ArgsStringList()
		return fmt.Sprintf("Args list: %s", strings.Join(argsStringList, ","))
	}
	if len(q.ArgMap) > 0 {
		var strs = make([]string, len(q.ArgMap))
		idx := 0
		for k, v := range q.ArgMap {
			strs[idx] = fmt.Sprintf("%s = %s", k, v)
			idx++
		}
		return fmt.Sprintf("args:\n\t%s", strings.Join(strs, "\n\t"))
	}
	return "<empty>"
}

// ArgsStringList convert ArgLists into list of strings
func (q *QueryArgs) ArgsStringList() []string {
	var argsStringList = make([]string, len(q.ArgList))
	for i, a := range q.ArgList {
		argsStringList[i] = typehelpers.SafeString(a)
	}
	return argsStringList
}

// ConvertArgsList convert argList into list of interface{} by unmarshalling
func (q *QueryArgs) ConvertArgsList() ([]any, error) {
	var argList = make([]any, len(q.ArgList))

	for i, a := range q.ArgList {
		if a != nil {
			// do we need to unmarshal?
			if _, stringArg := q.stringPositionalArgs[i]; stringArg {
				argList[i] = *a
			} else {
				// so this arg is stored as json - we need to deserialize
				err := json.Unmarshal([]byte(*a), &argList[i])
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return argList, nil
}

func NewQueryArgs() *QueryArgs {
	return &QueryArgs{
		ArgMap:               make(map[string]string),
		stringNamedArgs:      make(map[string]struct{}),
		stringPositionalArgs: make(map[int]struct{}),
	}
}

func (q *QueryArgs) Equals(other *QueryArgs) bool {
	if other == nil {
		return false
	}
	if q.Empty() {
		return other.Empty()
	}
	if len(other.ArgMap) != len(q.ArgMap) || len(other.ArgList) != len(q.ArgList) {
		return false
	}
	for k, v := range q.ArgMap {
		if !utils.SafeStringsEqual(other.ArgMap[k], v) {
			return false
		}
	}
	for i, v := range q.ArgList {
		if !utils.SafeStringsEqual(other.ArgList[i], v) {
			return false
		}
	}
	return true
}

func (q *QueryArgs) Empty() bool {
	return len(q.ArgMap)+len(q.ArgList) == 0
}

func (q *QueryArgs) Validate() error {
	if len(q.ArgMap) > 0 && len(q.ArgList) > 0 {
		return fmt.Errorf("args contain both positional and named parameters")
	}
	return nil
}

// Merge merges the other args with ourselves, creating and returning a new QueryArgs with the result
// NOTE: other has precedence
func (q *QueryArgs) Merge(other *QueryArgs, source QueryProvider) (*QueryArgs, error) {
	if other == nil {
		return q, nil
	}

	// ensure we valid before trying to merge (i.e. cannot define both arg list and arg map)
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("argument validation failed for '%s': %s", source.Name(), err.Error())
	}

	// ensure the other args are valid
	if err := other.Validate(); err != nil {
		return nil, fmt.Errorf("runtime argument validation failed for '%s': %s", source.Name(), err.Error())
	}

	// create a new query args to store the merged result
	result := NewQueryArgs()
	result.stringNamedArgs = other.stringNamedArgs
	result.stringPositionalArgs = other.stringPositionalArgs

	// named args
	// first set values from other
	for k, v := range other.ArgMap {
		result.ArgMap[k] = v

	}
	// now set any unset values from our map
	for k, v := range q.ArgMap {
		if _, ok := result.ArgMap[k]; !ok {
			result.ArgMap[k] = v
			if _, ok := q.stringNamedArgs[k]; ok {
				result.stringNamedArgs[k] = struct{}{}
			}
		}
	}

	// positional args
	// so we must have an args list - figure out how long
	listLength := len(q.ArgList)
	if otherLen := len(other.ArgList); otherLen > listLength {
		listLength = otherLen
	}
	if listLength > 0 {
		result.ArgList = make([]*string, listLength)

		// first set values from other
		copy(result.ArgList, other.ArgList)

		// now set any unset values from base list
		for i, a := range q.ArgList {
			if result.ArgList[i] == nil {
				result.ArgList[i] = a
				if _, ok := q.stringPositionalArgs[i]; ok {
					result.stringPositionalArgs[i] = struct{}{}
				}
			}
		}
	}

	// validate the merged result
	// runtime args must specify args in same way as base args (i.e. both must define either map or list)
	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("runtime argument validation failed when merging runtime args into '%s': %s", source.Name(), err.Error())
	}

	return result, nil
}

func (q *QueryArgs) SetNamedArgVal(value any, name string) (err error) {
	strVal, ok := value.(string)
	if ok {
		q.stringNamedArgs[name] = struct{}{}
	} else {
		strVal, err = q.ToString(value)
		if err != nil {
			return err
		}
	}
	q.ArgMap[name] = strVal
	return nil
}

func (q *QueryArgs) SetPositionalArgVal(value any, idx int) (err error) {
	if idx >= len(q.ArgList) {
		return fmt.Errorf("positional arg index %d out of range", idx)
	}
	strVal, ok := value.(string)
	if ok {
		// no need to convert toi string - make a note
		q.stringPositionalArgs[idx] = struct{}{}
	} else {
		strVal, err = q.ToString(value)
		if err != nil {
			return err
		}
	}
	q.ArgList[idx] = &strVal
	return nil
}

func (q *QueryArgs) ToString(value any) (string, error) {
	// format the arg value as a JSON string
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (q *QueryArgs) SetArgMap(argMap map[string]any) error {
	for k, v := range argMap {
		if err := q.SetNamedArgVal(v, k); err != nil {
			return err
		}
	}
	return nil
}

func (q *QueryArgs) SetArgList(argList []any) error {
	q.ArgList = make([]*string, len(argList))
	for i, v := range argList {
		if err := q.SetPositionalArgVal(v, i); err != nil {
			return err
		}
	}
	return nil
}

func (q *QueryArgs) GetNamedArg(name string) (interface{}, bool, error) {
	argStr, ok := q.ArgMap[name]
	if !ok {
		return nil, false, nil
	}
	// do we need to deserialise?
	if _, isStringArg := q.stringNamedArgs[name]; isStringArg {
		return argStr, true, nil
	}

	var res any
	if err := json.Unmarshal([]byte(argStr), &res); err != nil {
		return nil, false, err
	}
	return res, true, nil
}

func (q *QueryArgs) GetPositionalArg(idx int) (interface{}, bool, error) {
	if idx > len(q.ArgList) {
		return nil, false, fmt.Errorf("positional arg index %d out of range", idx)
	}
	argStrPtr := q.ArgList[idx]
	if argStrPtr == nil {
		return nil, false, nil
	}

	// do we need to deserialise?
	if _, isStringArg := q.stringPositionalArgs[idx]; isStringArg {
		return *argStrPtr, true, nil
	}

	var res any
	if err := json.Unmarshal([]byte(*argStrPtr), &res); err != nil {
		return nil, false, err
	}
	return res, true, nil
}
func (q *QueryArgs) resolveNamedParameters(queryProvider QueryProvider) (argVals []any, missingParams []string, err error) {
	// if query params contains both positional and named params, error out
	params := queryProvider.GetParams()

	argVals = make([]any, len(params))

	// iterate through each param def and resolve the value
	// build a map of which args have been matched (used to validate all args have param defs)
	argsWithParamDef := make(map[string]bool)
	for i, param := range params {
		// first set default
		defaultValue, err := param.GetDefault()
		if err != nil {
			return nil, nil, err
		}

		// can we resolve a value for this param?
		if argVal, ok, err := q.GetNamedArg(param.Name); ok {
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
	for arg := range q.ArgMap {
		if _, ok := argsWithParamDef[arg]; !ok {
			log.Printf("[TRACE] no parameter definition found for argument '%s'", arg)
		}
	}

	return argVals, missingParams, nil
}

func (q *QueryArgs) resolvePositionalParameters(queryProvider QueryProvider) (argValues []any, missingParams []string, err error) {
	// if query params contains both positional and named params, error out
	// if there are param defs - we must be able to resolve all params
	// if there are MORE defs than provided parameters, all remaining defs MUST provide a default
	params := queryProvider.GetParams()

	// if no param defs are defined, just use the given values, using runtime dependencies where available
	if len(params) == 0 {
		// no params defined, so we return as many args as are provided
		// (convert arg vals from json)
		argValues, err = q.ConvertArgsList()
		if err != nil {
			return nil, nil, err
		}
		return argValues, nil, nil
	}

	// verify we have enough args
	if len(params) < len(q.ArgList) {
		err = fmt.Errorf("resolvePositionalParameters failed for '%s' - %d %s were provided but there %s %d parameter %s",
			queryProvider.Name(),
			len(q.ArgList),
			utils.Pluralize("argument", len(q.ArgList)),
			utils.Pluralize("is", len(params)),
			len(params),
			utils.Pluralize("definition", len(params)),
		)
		return
	}

	// so there are param definitions - use these to populate argValues
	argValues = make([]any, len(params))

	for i, param := range params {
		// first set default
		defaultValue, err := param.GetDefault()
		if err != nil {
			return nil, nil, err
		}

		if i < len(q.ArgList) && q.ArgList[i] != nil {
			argVal, _, err := q.GetPositionalArg(i)
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
