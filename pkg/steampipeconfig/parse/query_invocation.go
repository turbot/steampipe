package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/type_conversion"
)

// ParseQueryInvocation parses a query invocation and extracts the args (if any)
// supported formats are:
//
// 1) positional args
// query.my_prepared_statement('val1','val1')
//
// 2) named args
// query.my_prepared_statement(my_arg1 => 'test', my_arg2 => 'test2')
func ParseQueryInvocation(arg string) (string, *modconfig.QueryArgs, error) {
	// TODO strip non printing chars
	args := &modconfig.QueryArgs{}

	arg = strings.TrimSpace(arg)
	query := arg
	var err error
	openBracketIdx := strings.Index(arg, "(")
	closeBracketIdx := strings.LastIndex(arg, ")")
	if openBracketIdx != -1 && closeBracketIdx == len(arg)-1 {
		argsString := arg[openBracketIdx+1 : len(arg)-1]
		args, err = parseArgs(argsString)
		query = strings.TrimSpace(arg[:openBracketIdx])
	}
	return query, args, err
}

// parse the actual args string, i.e. the contents of the bracket
// supported formats are:
//
// 1) positional args
// 'val1','val1'
//
// 2) named args
// my_arg1 => 'val1', my_arg2 => 'val2'
func parseArgs(argsString string) (*modconfig.QueryArgs, error) {
	res := modconfig.NewQueryArgs()
	if len(argsString) == 0 {
		return res, nil
	}

	// split on comma to get each arg string (taking quotes and brackets into account)
	splitArgs, err := splitArgString(argsString)
	if err != nil {
		// return empty result, even if we have an error
		return res, err
	}

	// first check for named args
	argMap, err := parseNamedArgs(splitArgs)
	if err != nil {
		return res, err
	}
	if err := res.SetArgMap(argMap); err != nil {
		return res, err
	}

	if res.Empty() {
		// no named args - fall back on positional
		argList, err := parsePositionalArgs(splitArgs)
		if err != nil {
			return res, err
		}
		if err := res.SetArgList(argList); err != nil {
			return res, err
		}
	}
	// return empty result, even if we have an error
	return res, err
}

func splitArgString(argsString string) ([]string, error) {
	var argsList []string
	openElements := map[string]int{
		"quote":  0,
		"curly":  0,
		"square": 0,
	}
	var currentWord string
	for _, c := range argsString {
		// should we split - are we in a block
		if c == ',' &&
			openElements["quote"] == 0 && openElements["curly"] == 0 && openElements["square"] == 0 {
			if len(currentWord) > 0 {
				argsList = append(argsList, currentWord)
				currentWord = ""
			}
		} else {
			currentWord = currentWord + string(c)
		}

		// handle brackets and quotes
		switch c {
		case '{':
			if openElements["quote"] == 0 {
				openElements["curly"]++
			}
		case '}':
			if openElements["quote"] == 0 {
				openElements["curly"]--
				if openElements["curly"] < 0 {
					return nil, fmt.Errorf("bad arg syntax")
				}
			}
		case '[':
			if openElements["quote"] == 0 {
				openElements["square"]++
			}
		case ']':
			if openElements["quote"] == 0 {
				openElements["square"]--
				if openElements["square"] < 0 {
					return nil, fmt.Errorf("bad arg syntax")
				}
			}
		case '"':
			if openElements["quote"] == 0 {
				openElements["quote"] = 1
			} else {
				openElements["quote"] = 0
			}
		}
	}
	if len(currentWord) > 0 {
		argsList = append(argsList, currentWord)
	}
	return argsList, nil
}

func parseArg(v string) (any, error) {
	b, diags := hclsyntax.ParseExpression([]byte(v), "", hcl.Pos{})
	if diags.HasErrors() {
		return "", plugin.DiagsToError("bad arg syntax", diags)
	}
	val, diags := b.Value(nil)
	if diags.HasErrors() {
		return "", plugin.DiagsToError("bad arg syntax", diags)
	}
	return type_conversion.CtyToGo(val)
}

func parseNamedArgs(argsList []string) (map[string]any, error) {
	var res = make(map[string]any)
	for _, p := range argsList {
		argTuple := strings.Split(strings.TrimSpace(p), "=>")
		if len(argTuple) != 2 {
			// not all args have valid syntax - give up
			return nil, nil
		}
		k := strings.TrimSpace(argTuple[0])
		val, err := parseArg(argTuple[1])
		if err != nil {
			return nil, err
		}
		res[k] = val
	}
	return res, nil
}

func parsePositionalArgs(argsList []string) ([]any, error) {
	// convert to pointer array
	res := make([]any, len(argsList))
	// just treat args as positional args
	// strip spaces
	for i, v := range argsList {
		valStr, err := parseArg(v)
		if err != nil {
			return nil, err
		}
		res[i] = valStr
	}

	return res, nil
}
