package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ParsePreparedStatementInvocation parses a query invocation and extracts th eparams (if any)
// supported formats are:
//
// 1) positional params
// query.my_prepared_statement('val1','val1')
//
// 2) named params
// query.my_prepared_statement(my_param1 => 'test', my_param2 => 'test2')
func ParsePreparedStatementInvocation(arg string) (string, *modconfig.QueryArgs, error) {
	// TODO strip non printing chars
	params := &modconfig.QueryArgs{}

	// only parse args for named query or named control invocation
	if !(strings.HasPrefix(arg, "query.") || strings.HasPrefix(arg, "control.")) {
		return arg, params, nil
	}

	arg = strings.TrimSpace(arg)
	query := arg
	var err error
	openBracketIdx := strings.Index(arg, "(")
	closeBracketIdx := strings.LastIndex(arg, ")")
	if openBracketIdx != -1 && closeBracketIdx == len(arg)-1 {
		paramsString := arg[openBracketIdx+1 : len(arg)-1]
		params, err = parseParams(paramsString)
		query = strings.TrimSpace(arg[:openBracketIdx])
	}
	return query, params, err
}

// parse the actual params string, i.e. the contents of the bracket
// supported formats are:
//
// 1) positional params
// 'val1','val1'
//
// 2) named params
// my_param1 => 'val1', my_param2 => 'val2'
func parseParams(paramsString string) (*modconfig.QueryArgs, error) {
	res := modconfig.NewQueryArgs()
	if len(paramsString) == 0 {
		return res, nil
	}

	// split on comma to get each param string (taking quotes and brackets into account)
	paramsList, err := splitParamString(paramsString)
	if err != nil {
		// return empty result, even if we have an error
		return res, err
	}

	// first check for named parameters
	res.Args, err = parseNamedParams(paramsList)
	if err != nil {
		return nil, err
	}
	if res.Empty() {
		// no named params - fall back on positional
		res.ArgsList, err = parsePositionalParams(paramsList)
	}
	// return empty result, even if we have an error
	return res, err
}

func splitParamString(paramsString string) ([]string, error) {
	var paramsList []string
	openElements := map[string]int{
		"quote":  0,
		"curly":  0,
		"square": 0,
	}
	var currentWord string
	for _, c := range paramsString {
		// should we split - are we in a block
		if c == ',' &&
			openElements["quote"] == 0 && openElements["curly"] == 0 && openElements["square"] == 0 {
			if len(currentWord) > 0 {
				paramsList = append(paramsList, currentWord)
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
					return nil, fmt.Errorf("bad parameter syntax")
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
					return nil, fmt.Errorf("bad parameter syntax")
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
		paramsList = append(paramsList, currentWord)
	}
	return paramsList, nil
}

func parseParam(v string) (string, error) {
	b, diags := hclsyntax.ParseExpression([]byte(v), "", hcl.Pos{})
	if diags.HasErrors() {
		return "", plugin.DiagsToError("bad parameter syntax", diags)
	}
	val, diags := b.Value(nil)
	if diags.HasErrors() {
		return "", plugin.DiagsToError("bad parameter syntax", diags)
	}
	return ctyToPostgresString(val)
}

func parseNamedParams(paramsList []string) (map[string]string, error) {
	var res = make(map[string]string)
	for _, p := range paramsList {
		paramTuple := strings.Split(strings.TrimSpace(p), "=>")
		if len(paramTuple) != 2 {
			// not all params have valid syntax - give up
			return nil, nil
		}
		k := strings.TrimSpace(paramTuple[0])
		valStr, err := parseParam(paramTuple[1])
		if err != nil {
			return nil, err
		}
		res[k] = valStr
	}
	return res, nil
}

func parsePositionalParams(paramsList []string) ([]string, error) {
	// just treat params as positional parameters
	// strip spaces
	for i, v := range paramsList {
		valStr, err := parseParam(v)
		if err != nil {
			return nil, err
		}
		paramsList[i] = valStr
	}

	return paramsList, nil
}
