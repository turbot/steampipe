package modconfig

import "sort"

type WorkspaceResources struct {
	Query     map[string]bool
	Control   map[string]bool
	Benchmark map[string]bool
}

func NewWorkspaceResources() *WorkspaceResources {
	return &WorkspaceResources{
		Query:     make(map[string]bool),
		Control:   make(map[string]bool),
		Benchmark: make(map[string]bool),
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
	return r
}

// GetSortedBenchmarksAndControlNames gives back a list of the benchmarks
// and controls in the current workspace.
// The list is sorted alphabetically - with the benchmarks first
func (w *WorkspaceResources) GetSortedBenchmarksAndControlNames() []string {
	benchmarkList := []string{}
	controlList := []string{}

	for key := range w.Benchmark {
		benchmarkList = append(benchmarkList, key)
	}

	for key := range w.Control {
		controlList = append(controlList, key)
	}

	sort.Strings(benchmarkList)
	sort.Strings(controlList)

	return append(benchmarkList, controlList...)
}

func (w *WorkspaceResources) GetSortedNamedQueryNames() []string {
	namedQueries := []string{}
	for key := range w.Query {
		namedQueries = append(namedQueries, key)
	}
	sort.Strings(namedQueries)
	return namedQueries
}
