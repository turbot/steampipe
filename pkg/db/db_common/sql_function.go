package db_common

// SQLFunction is a struct for an sqlFunc
type SQLFunction struct {
	Name     string
	Params   map[string]string
	Returns  string
	Body     string
	Language string
}
