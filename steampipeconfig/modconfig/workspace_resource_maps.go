package modconfig

// WorkspaceResourceMaps is a struct containing maps of all mod resource types
// This is provided to avoid db needing to reference workspace package
type WorkspaceResourceMaps struct {
	Mods             map[string]*Mod
	Queries          map[string]*Query
	Controls         map[string]*Control
	Benchmarks       map[string]*Benchmark
	Variables        map[string]*Variable
	Reports          map[string]*ReportContainer
	ReportContainers map[string]*ReportContainer
	ReportTables     map[string]*ReportTable
	ReportTexts      map[string]*ReportText
	ReportCounters   map[string]*ReportCounter
	ReportCharts     map[string]*ReportChart
	References       map[string]*ResourceReference
}

func NewWorkspaceResourceMaps() *WorkspaceResourceMaps {
	return &WorkspaceResourceMaps{
		Mods:             make(map[string]*Mod),
		Queries:          make(map[string]*Query),
		Controls:         make(map[string]*Control),
		Benchmarks:       make(map[string]*Benchmark),
		Variables:        make(map[string]*Variable),
		Reports:          make(map[string]*ReportContainer),
		ReportContainers: make(map[string]*ReportContainer),
		ReportTables:     make(map[string]*ReportTable),
		ReportTexts:      make(map[string]*ReportText),
		ReportCounters:   make(map[string]*ReportCounter),
		ReportCharts:     make(map[string]*ReportChart),
		References:       make(map[string]*ResourceReference),
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

	for name, report := range m.Reports {
		if otherReport, ok := other.Reports[name]; !ok {
			return false
		} else if !report.Equals(otherReport) {
			return false
		}
	}
	for name := range other.Reports {
		if _, ok := m.Reports[name]; !ok {
			return false
		}
	}

	for name, container := range m.ReportContainers {
		if otherReport, ok := other.ReportContainers[name]; !ok {
			return false
		} else if !container.Equals(otherReport) {
			return false
		}
	}
	for name := range other.ReportContainers {
		if _, ok := m.ReportContainers[name]; !ok {
			return false
		}
	}

	for name, tables := range m.ReportTables {
		if otherReport, ok := other.ReportTables[name]; !ok {
			return false
		} else if !tables.Equals(otherReport) {
			return false
		}
	}
	for name := range other.ReportTables {
		if _, ok := m.ReportTables[name]; !ok {
			return false
		}
	}

	for name, texts := range m.ReportTexts {
		if otherReport, ok := other.ReportTexts[name]; !ok {
			return false
		} else if !texts.Equals(otherReport) {
			return false
		}
	}
	for name := range other.ReportTexts {
		if _, ok := m.ReportTexts[name]; !ok {
			return false
		}
	}

	for name, counters := range m.ReportCounters {
		if otherReport, ok := other.ReportCounters[name]; !ok {
			return false
		} else if !counters.Equals(otherReport) {
			return false
		}
	}
	for name := range other.ReportCounters {
		if _, ok := m.ReportCounters[name]; !ok {
			return false
		}
	}

	for name, charts := range m.ReportCharts {
		if otherReport, ok := other.ReportCharts[name]; !ok {
			return false
		} else if !charts.Equals(otherReport) {
			return false
		}
	}
	for name := range other.ReportCharts {
		if _, ok := m.ReportCharts[name]; !ok {
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

	// TODO KAI OTHER REFERENCEQ TYPES
}

func (m *WorkspaceResourceMaps) Empty() bool {
	return len(m.Mods)+
		len(m.Queries)+
		len(m.Controls)+
		len(m.Benchmarks)+
		len(m.Variables)+
		len(m.Reports)+
		len(m.ReportContainers)+
		len(m.ReportTables)+
		len(m.ReportTexts)+
		len(m.ReportCounters)+
		len(m.ReportCharts)+
		len(m.References) == 0
}
