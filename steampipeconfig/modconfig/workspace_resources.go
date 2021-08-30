package modconfig

type WorkspaceResources struct {
	Query     map[string]bool
	Control   map[string]bool
	Benchmark map[string]bool
	//Panel  map[string]bool
}

func NewWorkspaceResources() *WorkspaceResources {
	return &WorkspaceResources{
		Query:     make(map[string]bool),
		Control:   make(map[string]bool),
		Benchmark: make(map[string]bool),
		//Panel: make(map[string]bool),
	}
}
func (r *WorkspaceResources) Merge(other *WorkspaceResources) *WorkspaceResources {
	for k := range other.Query {
		r.Query[k] = true
	}
	for k := range other.Control {
		r.Control[k] = true
	}
	for k := range other.Benchmark {
		r.Benchmark[k] = true
	}
	//for k := range other.Panel{
	//	r.Panel[k] = true
	//}
	return r
}
