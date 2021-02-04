package schema

// SQLFunc :: struct for an sqlFunc
type SQLFunc struct {
	Name     string
	Params   map[string]string
	Returns  string
	Body     string
	Language string
}
