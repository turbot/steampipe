package modconfig

// ResolvedQuery contains the execute SQL, raw SQL and args string used to execute a query
type ResolvedQuery struct {
	SQL    string
	Args   []any
	Params []*ParamDef
}

func (r ResolvedQuery) QueryArgs() *QueryArgs {
	res := NewQueryArgs()

	// TODO KAI this assumes string args
	res.ArgList = make([]*string, len(r.Args))

	for i, a := range r.Args {
		if argStr, ok := a.(string); ok {
			res.ArgList[i] = &argStr
		}
	}
	return res
}
