package modconfig

// WorkspaceResourceMaps is a struct containing maps of all mod resource types
// This is provided to avoid db needing to reference workspace package
type WorkspaceResourceMaps struct {
	// the parent mod
	Mod *Mod
	// all mods (including deps)
	Mods              map[string]*Mod
	Queries           map[string]*Query
	Controls          map[string]*Control
	Benchmarks        map[string]*Benchmark
	Variables         map[string]*Variable
	Reports           map[string]*ReportContainer
	ReportContainers  map[string]*ReportContainer
	ReportCards       map[string]*ReportCard
	ReportCharts      map[string]*ReportChart
	ReportHierarchies map[string]*ReportHierarchy
	ReportImages      map[string]*ReportImage
	ReportInputs      map[string]*ReportInput
	ReportTables      map[string]*ReportTable
	ReportTexts       map[string]*ReportText
	References        map[string]*ResourceReference

	LocalQueries    map[string]*Query
	LocalControls   map[string]*Control
	LocalBenchmarks map[string]*Benchmark
}

func WorkspaceResourceMapFromMod(mod *Mod) *WorkspaceResourceMaps {
	resourceMaps := &WorkspaceResourceMaps{
		Mod:               mod,
		Mods:              make(map[string]*Mod),
		Queries:           mod.Queries,
		Controls:          mod.Controls,
		Benchmarks:        mod.Benchmarks,
		Variables:         mod.Variables,
		Reports:           mod.Reports,
		ReportContainers:  mod.ReportContainers,
		ReportCharts:      mod.ReportCharts,
		ReportCards:       mod.ReportCards,
		ReportHierarchies: mod.ReportHierarchies,
		ReportImages:      mod.ReportImages,
		ReportInputs:      mod.ReportInputs,
		ReportTables:      mod.ReportTables,
		ReportTexts:       mod.ReportTexts,
	}
	if !mod.IsDefaultMod() {
		resourceMaps.Mods[mod.Name()] = mod
	}
	return resourceMaps
}

func (m *WorkspaceResourceMaps) Equals(other *WorkspaceResourceMaps) bool {
	if other == nil {
		return false
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

	for name, cards := range m.ReportCards {
		if otherReport, ok := other.ReportCards[name]; !ok {
			return false
		} else if !cards.Equals(otherReport) {
			return false
		}
	}
	for name := range other.ReportCards {
		if _, ok := m.ReportCards[name]; !ok {
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

	for name, hierarchies := range m.ReportHierarchies {
		if otherHierarchy, ok := other.ReportHierarchies[name]; !ok {
			return false
		} else if !hierarchies.Equals(otherHierarchy) {
			return false
		}
	}
	for name := range other.ReportHierarchies {
		if _, ok := m.ReportHierarchies[name]; !ok {
			return false
		}
	}

	for name, images := range m.ReportImages {
		if otherImage, ok := other.ReportImages[name]; !ok {
			return false
		} else if !images.Equals(otherImage) {
			return false
		}
	}
	for name := range other.ReportImages {
		if _, ok := m.ReportImages[name]; !ok {
			return false
		}
	}

	for name, images := range m.ReportInputs {
		if otherImage, ok := other.ReportInputs[name]; !ok {
			return false
		} else if !images.Equals(otherImage) {
			return false
		}
	}
	for name := range other.ReportInputs {
		if _, ok := m.ReportInputs[name]; !ok {
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

	// TODO add other reference types - https://github.com/turbot/steampipe/issues/1331
}

func (m *WorkspaceResourceMaps) Empty() bool {
	return len(m.Mods)+
		len(m.Queries)+
		len(m.Controls)+
		len(m.Benchmarks)+
		len(m.Variables)+
		len(m.Reports)+
		len(m.ReportContainers)+
		len(m.ReportCards)+
		len(m.ReportCharts)+
		len(m.ReportHierarchies)+
		len(m.ReportImages)+
		len(m.ReportInputs)+
		len(m.ReportTables)+
		len(m.ReportTexts)+
		len(m.References) == 0
}
