package db_common

type QueryWithArgs struct {
	Query string
	Args  []any
}

type QueriesWithArgs []QueryWithArgs

// Add returns a new instance of QueriesWithArgs with the given QueryWithArgs appended
func (q QueriesWithArgs) Add(another ...QueryWithArgs) QueriesWithArgs {
	return append(q, another...)
}
