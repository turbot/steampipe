package modconfig

// WorkspaceResourceMaps is a struct containing maps of all mod resource types
// This is provided to avoid db needing to reference workspace package
type WorkspaceResourceMaps struct {
	Mods       map[string]*Mod
	Queries    map[string]*Query
	Controls   map[string]*Control
	Benchmarks map[string]*Benchmark
	Variables  map[string]*Variable
}

func NewWorkspaceResourceMaps() *WorkspaceResourceMaps {
	return &WorkspaceResourceMaps{
		Mods:       make(map[string]*Mod),
		Queries:    make(map[string]*Query),
		Controls:   make(map[string]*Control),
		Benchmarks: make(map[string]*Benchmark),
		Variables:  make(map[string]*Variable),
	}
}
func (m WorkspaceResourceMaps) Equals(other *WorkspaceResourceMaps) bool {
	for name, mod := range m.Mods {
		if otherMod, ok := other.Mods[name]; !ok {
			return false
		} else if !mod.Equals(otherMod) {
			return false
		}
	}
	for name := range other.Mods {
		if _, ok := m.Mods[name]; !ok {
			return false
		}
	}
	for name, query := range m.Queries {
		if otherQuery, ok := other.Queries[name]; !ok {
			return false
		} else if !query.Equals(otherQuery) {
			return false
		}
	}

	for name := range other.Queries {
		if _, ok := m.Queries[name]; !ok {
			return false
		}
	}
	for name, control := range m.Controls {
		if otherControl, ok := other.Controls[name]; !ok {
			return false
		} else if !control.Equals(otherControl) {
			return false
		}
	}
	for name := range other.Controls {
		if _, ok := m.Controls[name]; !ok {
			return false
		}
	}
	for name, benchmark := range m.Benchmarks {
		if otherBenchmark, ok := other.Benchmarks[name]; !ok {
			return false
		} else if !benchmark.Equals(otherBenchmark) {
			return false
		}
	}
	for name := range other.Benchmarks {
		if _, ok := m.Benchmarks[name]; !ok {
			return false
		}
	}
	for name, variable := range m.Variables {
		if otherVariable, ok := other.Variables[name]; !ok {
			return false
		} else if !variable.Equals(otherVariable) {
			return false
		}
	}
	for name := range other.Variables {
		if _, ok := m.Variables[name]; !ok {
			return false
		}
	}
	return true
}

func (m WorkspaceResourceMaps) AddPreparedStatementProvider(provider PreparedStatementProvider) {
	switch p := provider.(type) {
	case *Query:
		if p != nil {
			m.Queries[p.FullName] = p
		}
	case *Control:
		if p != nil {
			m.Controls[p.FullName] = p
		}
	}
}
