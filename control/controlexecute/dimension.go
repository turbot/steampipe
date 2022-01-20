package controlexecute

// Dimension is a struct representing an attribute returned by a control run.
// An attribute is stored as a dimension if it's not a standard attribute (reason, resource, status).
type Dimension struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
