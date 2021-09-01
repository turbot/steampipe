package modconfig

// QueryArgs is a struct which contains the arguments used to invoke a prepared statement
// these may either be passed by name, in a map, or as a list of positional args
// NOTE: if both are present the named parameters are used
type QueryArgs struct {
	Args     map[string]string
	ArgsList []string
}

func NewQueryArgs() *QueryArgs {
	return &QueryArgs{
		Args: make(map[string]string),
	}
}

func (q *QueryArgs) Equals(other *QueryArgs) bool {
	if other == nil {
		return false
	}
	if q.Empty() {
		return other.Empty()
	}
	if len(other.Args) != len(q.Args) || len(other.ArgsList) != len(q.ArgsList) {
		return false
	}
	for k, v := range q.Args {
		if other.Args[k] != v {
			return false
		}
	}
	for i, v := range q.ArgsList {
		if other.ArgsList[i] != v {
			return false
		}
	}
	return true
}

func (q *QueryArgs) Empty() bool {
	return len(q.Args)+len(q.ArgsList) == 0
}
