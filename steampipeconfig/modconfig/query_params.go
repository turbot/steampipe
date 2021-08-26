package modconfig

// QueryParams is a struct which contains the parameters used to invoke a prepared statement
// these may either be passed by name, in a map, or as a list of positional params
// NOTE: if both are present the named parameters are used
type QueryParams struct {
	Params     map[string]string
	ParamsList []string
}

func NewQueryParams() *QueryParams {
	return &QueryParams{
		Params: make(map[string]string),
	}
}

func (q *QueryParams) Equals(other *QueryParams) bool {
	if q.Empty() {
		return other.Empty()
	}
	if len(other.Params) != len(q.Params) || len(other.ParamsList) != len(q.ParamsList) {
		return false
	}
	for k, v := range q.Params {
		if other.Params[k] != v {
			return false
		}
	}
	for i, v := range q.ParamsList {
		if other.ParamsList[i] != v {
			return false
		}
	}
	return true
}

func (q *QueryParams) Empty() bool {
	return len(q.Params)+len(q.ParamsList) == 0
}
