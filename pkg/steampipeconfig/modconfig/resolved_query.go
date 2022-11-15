package modconfig

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// ResolvedQuery contains the execute SQL, raw SQL and args string used to execute a query
type ResolvedQuery struct {
	SQL    string
	Args   []any
	Params []*ParamDef
}

func (r ResolvedQuery) QueryArgs() *QueryArgs {
	res := NewQueryArgs()

	res.ArgList = make([]cty.Value, len(r.Args))

	for i, a := range r.Args {
		// TODO KAI assume string - also support array
		if _, ok := a.(string); ok {
			ctyVal, err := gocty.ToCtyValue(a, cty.String)
			if err != nil {
				res.ArgList[i] = ctyVal
			}
		}
	}
	return res
}
