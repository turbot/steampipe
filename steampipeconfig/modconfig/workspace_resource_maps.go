package modconfig

// WorkspaceResourceMaps is a struct containing maps of all mod resource types
// This is provided to avoid db needing to reference workspace package
type WorkspaceResourceMaps struct {
	Mods       map[string]*Mod
	Queries    map[string]*Query
	Controls   map[string]*Control
	Benchmarks map[string]*Benchmark
	Variables  map[string]*Variable
	References map[string]*ResourceReference
}

func NewWorkspaceResourceMaps() *WorkspaceResourceMaps {
	return &WorkspaceResourceMaps{
		Mods:       make(map[string]*Mod),
		Queries:    make(map[string]*Query),
		Controls:   make(map[string]*Control),
		Benchmarks: make(map[string]*Benchmark),
		Variables:  make(map[string]*Variable),
		References: make(map[string]*ResourceReference),
	}
}

func (m *WorkspaceResourceMaps) Equals(other *WorkspaceResourceMaps) bool {
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
	for name, reference := range m.References {
		if otherReference, ok := other.References[name]; !ok {
			return false
		} else if !reference.Equals(otherReference) {
			return false
		}
	}
	for name := range other.References {
		if _, ok := m.References[name]; !ok {
			return false
		}
	}

	return true
}

func (m *WorkspaceResourceMaps) AddQueryProvider(provider QueryProvider) {
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

func (m *WorkspaceResourceMaps) PopulateReferences() {
	m.References = make(map[string]*ResourceReference)
	for _, mod := range m.Mods {
		for _, ref := range mod.References {
			m.References[ref.String()] = ref
		}
	}

	for _, query := range m.Queries {
		for _, ref := range query.References {
			m.References[ref.String()] = ref
		}
	}

	for _, control := range m.Controls {
		for _, ref := range control.References {
			m.References[ref.String()] = ref
		}
	}

	for _, benchmark := range m.Benchmarks {
		for _, ref := range benchmark.References {
			m.References[ref.String()] = ref
		}
	}
}

func (m *WorkspaceResourceMaps) Empty() bool {
	return len(m.Mods)+
		len(m.Queries)+
		len(m.Controls)+
		len(m.Benchmarks)+
		len(m.Variables)+
		len(m.References) == 0
}
