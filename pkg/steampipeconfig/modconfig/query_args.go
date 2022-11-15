package modconfig

import (
	"fmt"
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

// SafeArgsList convert ArgLists into list of strings but return as an interface slice
func (q *QueryArgs) SafeArgsList() []any {
	var argsStringList = make([]any, len(q.ArgList))
	for i, a := range q.ArgList {
		argsStringList[i] = typehelpers.SafeString(a)
	}
	return argsStringList
}

func NewQueryArgs() *QueryArgs {
	return &QueryArgs{
		ArgMap: make(map[string]string),
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

	// ensure valid (i.e. cannot define both arg list and arg map)
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("argument validation failed for '%s': %s", source.Name(), err.Error())
	}

	if err := other.Validate(); err != nil {
		return nil, fmt.Errorf("runtime argument validation failed for '%s': %s", source.Name(), err.Error())
	}

	// create a new query args to store the merged result
	result := NewQueryArgs()

	// runtime args must specify args in same way as base args (i.e. both must define either map or list)
	if len(q.ArgMap)+len(other.ArgMap) > 0 {
		if len(other.ArgList) > 0 {
			return nil, fmt.Errorf("runtime argument validation failed for '%s': runtime args must be provided in same format (map or list) as base args", source.Name())
		}
		// first set values from other
		for k, v := range other.ArgMap {
			result.ArgMap[k] = v
		}
		// now set any unset values from our map
		for k, v := range q.ArgMap {
			if _, ok := result.ArgMap[k]; !ok {
				result.ArgMap[k] = v
			}
		}
	} else {
		// so we must have an args list - figure out how long
		listLength := len(q.ArgList)
		if otherLen := len(other.ArgList); otherLen > listLength {
			listLength = otherLen
		}
		result.ArgList = make([]*string, listLength)

		// first set values from other
		copy(result.ArgList, other.ArgList)

		// now set any unset values from base list
		for i, a := range q.ArgList {
			if result.ArgList[i] == nil {
				result.ArgList[i] = a
			}
		}
	}

	return result, nil
}
