package modconfig

// ResolvedQuery contains the execute SQL, raw SQL and args string used to execute a query
type ResolvedQuery struct {
	SQL    string
	Args   []any
	Params []*ParamDef
}

func (r ResolvedQuery) QueryArgs() *QueryArgs {
	res := NewQueryArgs()
	res.ArgList = r.Args
	return res
}
