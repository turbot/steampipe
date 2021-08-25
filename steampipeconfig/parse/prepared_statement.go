package parse

import (
	"strings"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ParsePreparedStatementInvocation parses a query invocation and extracts th eparams (if any)
// supported formats are:
//
// 1) positional params
// query.my_prepared_statement(array['val1','val1'])
//
// 2) named params
// query.my_prepared_statement(my_param1 => 'test', my_param2 => 'test2')
func ParsePreparedStatementInvocation(arg string) (string, *modconfig.QueryParams) {
	params := &modconfig.QueryParams{}
	arg = strings.TrimSpace(arg)
	queryName := arg
	openBracketIdx := strings.Index(arg, "(")
	closeBracketIdx := strings.LastIndex(arg, ")")
	if openBracketIdx != -1 && closeBracketIdx == len(arg)-1 {
		paramsString := arg[openBracketIdx+1 : len(arg)-1]
		params = parseParams(paramsString)
		queryName = strings.TrimSpace(arg[:openBracketIdx])
	}
	return queryName, params
}

// parse the actual params string, i.e. the contents of the bracket
// supported formats are:
//
// 1) positional params
// 'val1','val1'
//
// 2) named params
// my_param1 => 'val1', my_param2 => 'val2'
func parseParams(paramsString string) *modconfig.QueryParams {
	res := modconfig.NewQueryParams()
	if len(paramsString) == 0 {
		return res
	}

	// split on comma to get each param string
	paramsList := strings.Split(paramsString, ",")

	// first check for named parameters
	res.Params = parseNamedParams(paramsList)
	if !res.Empty() {
		return res
	}

	// just treat params as positional parameters
	// strip spaces
	for i, v := range paramsList {
		paramsList[i] = strings.TrimSpace(v)
	}
	res.ParamsList = paramsList
	return res
}

func parseNamedParams(paramsList []string) map[string]string {
	var res = make(map[string]string)
	for _, p := range paramsList {
		paramTuple := strings.Split(strings.TrimSpace(p), "=>")
		if len(paramTuple) != 2 {
			// not all params have valid syntax - give up
			return nil
		}
		k := strings.TrimSpace(paramTuple[0])
		v := strings.Trim(strings.TrimSpace(paramTuple[1]), "'")
		res[k] = v
	}
	return res
}
