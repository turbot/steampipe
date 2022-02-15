package modconfig

// WorkspaceResourceMaps is a struct containing maps of all mod resource types
// This is provided to avoid db needing to reference workspace package
type WorkspaceResourceMaps struct {
	// the parent mod
	Mod *Mod
	// all mods (including deps)
	Mods                 map[string]*Mod
	Queries              map[string]*Query
	Controls             map[string]*Control
	Benchmarks           map[string]*Benchmark
	Variables            map[string]*Variable
	Dashboards           map[string]*Dashboard
	DashboardContainers  map[string]*DashboardContainer
	DashboardCards       map[string]*DashboardCard
	DashboardCharts      map[string]*DashboardChart
	DashboardHierarchies map[string]*DashboardHierarchy
	DashboardImages      map[string]*DashboardImage
	DashboardInputs      map[string]*DashboardInput
	DashboardTables      map[string]*DashboardTable
	DashboardTexts       map[string]*DashboardText
	References           map[string]*ResourceReference

	LocalQueries    map[string]*Query
	LocalControls   map[string]*Control
	LocalBenchmarks map[string]*Benchmark
}

func WorkspaceResourceMapFromMod(mod *Mod) *WorkspaceResourceMaps {
	resourceMaps := &WorkspaceResourceMaps{
		Mod:                  mod,
		Mods:                 make(map[string]*Mod),
		Queries:              mod.Queries,
		Controls:             mod.Controls,
		Benchmarks:           mod.Benchmarks,
		Variables:            mod.Variables,
		Dashboards:           mod.Dashboards,
		DashboardContainers:  mod.DashboardContainers,
		DashboardCharts:      mod.DashboardCharts,
		DashboardCards:       mod.DashboardCards,
		DashboardHierarchies: mod.DashboardHierarchies,
		DashboardImages:      mod.DashboardImages,
		DashboardInputs:      mod.DashboardInputs,
		DashboardTables:      mod.DashboardTables,
		DashboardTexts:       mod.DashboardTexts,
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

	for name, dashboard := range m.Dashboards {
		if otherDashboard, ok := other.Dashboards[name]; !ok {
			return false
		} else if !dashboard.Equals(otherDashboard) {
			return false
		}
	}
	for name := range other.Dashboards {
		if _, ok := m.Dashboards[name]; !ok {
			return false
		}
	}

	for name, container := range m.DashboardContainers {
		if otherContainer, ok := other.DashboardContainers[name]; !ok {
			return false
		} else if !container.Equals(otherContainer) {
			return false
		}
	}
	for name := range other.DashboardContainers {
		if _, ok := m.DashboardContainers[name]; !ok {
			return false
		}
	}

	for name, cards := range m.DashboardCards {
		if otherCard, ok := other.DashboardCards[name]; !ok {
			return false
		} else if !cards.Equals(otherCard) {
			return false
		}
	}
	for name := range other.DashboardCards {
		if _, ok := m.DashboardCards[name]; !ok {
			return false
		}
	}

	for name, charts := range m.DashboardCharts {
		if otherChart, ok := other.DashboardCharts[name]; !ok {
			return false
		} else if !charts.Equals(otherChart) {
			return false
		}
	}
	for name := range other.DashboardCharts {
		if _, ok := m.DashboardCharts[name]; !ok {
			return false
		}
	}

	for name, hierarchies := range m.DashboardHierarchies {
		if otherHierarchy, ok := other.DashboardHierarchies[name]; !ok {
			return false
		} else if !hierarchies.Equals(otherHierarchy) {
			return false
		}
	}
	for name := range other.DashboardHierarchies {
		if _, ok := m.DashboardHierarchies[name]; !ok {
			return false
		}
	}

	for name, images := range m.DashboardImages {
		if otherImage, ok := other.DashboardImages[name]; !ok {
			return false
		} else if !images.Equals(otherImage) {
			return false
		}
	}
	for name := range other.DashboardImages {
		if _, ok := m.DashboardImages[name]; !ok {
			return false
		}
	}

	for name, images := range m.DashboardInputs {
		if otherImage, ok := other.DashboardInputs[name]; !ok {
			return false
		} else if !images.Equals(otherImage) {
			return false
		}
	}
	for name := range other.DashboardInputs {
		if _, ok := m.DashboardInputs[name]; !ok {
			return false
		}
	}

	for name, tables := range m.DashboardTables {
		if otherTable, ok := other.DashboardTables[name]; !ok {
			return false
		} else if !tables.Equals(otherTable) {
			return false
		}
	}
	for name := range other.DashboardTables {
		if _, ok := m.DashboardTables[name]; !ok {
			return false
		}
	}

	for name, texts := range m.DashboardTexts {
		if otherText, ok := other.DashboardTexts[name]; !ok {
			return false
		} else if !texts.Equals(otherText) {
			return false
		}
	}
	for name := range other.DashboardTexts {
		if _, ok := m.DashboardTexts[name]; !ok {
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
		len(m.Dashboards)+
		len(m.DashboardContainers)+
		len(m.DashboardCards)+
		len(m.DashboardCharts)+
		len(m.DashboardHierarchies)+
		len(m.DashboardImages)+
		len(m.DashboardInputs)+
		len(m.DashboardTables)+
		len(m.DashboardTexts)+
		len(m.References) == 0
}
