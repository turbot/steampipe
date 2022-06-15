package modconfig

// ResolvedQuery contains the execute SQL, raw SQL and args string used to execute a query
type ResolvedQuery struct {
	ExecuteSQL string
	RawSQL     string
	Args       []string
	Params     []*ParamDef
}
