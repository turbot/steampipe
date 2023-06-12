package modconfig

import (
	"encoding/json"
)

// ResolvedQuery contains the execute SQL, raw SQL and args string used to execute a query
type ResolvedQuery struct {
	ExecuteSQL  string
	RawSQL      string
	Args        []any
	IsMetaQuery bool
}

// QueryArgs converts the ResolvedQuery into  QueryArgs
func (r ResolvedQuery) QueryArgs() *QueryArgs {
	res := NewQueryArgs()

	res.ArgList = make([]*string, len(r.Args))

	for i, a := range r.Args {
		// TACTICAL convert to JSON representation
		jsonBytes, err := json.Marshal(a)
		argStr := string(jsonBytes)
		if err != nil {
			res.ArgList[i] = &argStr
		}
	}
	return res
}
