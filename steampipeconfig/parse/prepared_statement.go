package parse

import (
	"fmt"
	"log"
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
func ParsePreparedStatementInvocation(arg string) (string, *modconfig.QueryParams) {
	// TODO strip non printing chars
	params := &modconfig.QueryParams{}
	arg = strings.TrimSpace(arg)
	query := arg
	openBracketIdx := strings.Index(arg, "(")
	closeBracketIdx := strings.LastIndex(arg, ")")
	if openBracketIdx != -1 && closeBracketIdx == len(arg)-1 {
		paramsString := arg[openBracketIdx+1 : len(arg)-1]
		var err error
		params, err = parseParams(paramsString)
		if err == nil {
			query = strings.TrimSpace(arg[:openBracketIdx])
		} else {
			// if we failed to parse the query as a prepared statement invocation, just return the raw query to execute
			log.Printf("[TRACE] ParsePreparedStatementInvocation returned error, executing query as raw SQL: %v", err)
		}
	}
	return query, params
}

// parse the actual params string, i.e. the contents of the bracket
// supported formats are:
//
// 1) positional params
// 'val1','val1'
//
// 2) named params
// my_param1 => 'val1', my_param2 => 'val2'
func parseParams(paramsString string) (*modconfig.QueryParams, error) {
	res := modconfig.NewQueryParams()
	if len(paramsString) == 0 {
		return res, nil
	}

	// split on comma to get each param string (taking quotes and brackets into account)
	paramsList, err := splitParamString(paramsString)
	if err != nil {
		return nil, err
	}

	// first check for named parameters
	res.Params, err = parseNamedParams(paramsList)
	if err != nil {
		return nil, err
	}
	if res.Empty() {
		// no named params - fall back on positional
		res.ParamsList, err = parsePositionalParams(paramsList)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
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
