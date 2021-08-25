package parse

import (
	"regexp"
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
		// if we successfully parsed params, extract the query name from the start of the string
		if !params.Empty() {
			queryName = strings.TrimSpace(arg[:openBracketIdx])
		}
	}
	return queryName, params
}

// parse the actual params string, i.e. the contents of the bracket
// supported formats are:
//
// 1) positional params
// array['val1','val1']
//
// 2) named params
// my_param1 => 'val1', my_param2 => 'val2'
func parseParams(paramsString string) *modconfig.QueryParams {
	res := modconfig.NewQueryParams()
	// first check for positional parameters
	r := *regexp.MustCompile(`array\[(.*)\]`)
	regexResult := r.FindStringSubmatch(paramsString)
	if len(regexResult) > 0 {
		paramsList := strings.Split(regexResult[1], ",")
		for i, p := range paramsList {
			paramsList[i] = strings.Trim(strings.TrimSpace(p), "'")
		}

		res.ParamsList = paramsList
		return res
	}

	// otherwise check for named parameters
	paramsList := strings.Split(paramsString, ",")
	for _, p := range paramsList {
		paramTuple := strings.Split(strings.TrimSpace(p), "=>")
		if len(paramTuple) != 2 {
			// not all params have valid syntax - give up
			return res
		}
		k := strings.TrimSpace(paramTuple[0])
		v := strings.Trim(strings.TrimSpace(paramTuple[1]), "'")
		res.Params[k] = v
	}

	return res
}
