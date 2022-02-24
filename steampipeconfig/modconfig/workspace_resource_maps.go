package modconfig

import "github.com/turbot/go-kit/helpers"

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
	Locals               map[string]*Local

	LocalQueries    map[string]*Query
	LocalControls   map[string]*Control
	LocalBenchmarks map[string]*Benchmark
}

func NewWorkspaceResourceMaps(mod *Mod) *WorkspaceResourceMaps {
	return &WorkspaceResourceMaps{
		Mod:                  mod,
		Mods:                 make(map[string]*Mod),
		Queries:              make(map[string]*Query),
		Controls:             make(map[string]*Control),
		Benchmarks:           make(map[string]*Benchmark),
		Variables:            make(map[string]*Variable),
		Dashboards:           make(map[string]*Dashboard),
		DashboardContainers:  make(map[string]*DashboardContainer),
		DashboardCards:       make(map[string]*DashboardCard),
		DashboardCharts:      make(map[string]*DashboardChart),
		DashboardHierarchies: make(map[string]*DashboardHierarchy),
		DashboardImages:      make(map[string]*DashboardImage),
		DashboardInputs:      make(map[string]*DashboardInput),
		DashboardTables:      make(map[string]*DashboardTable),
		DashboardTexts:       make(map[string]*DashboardText),
		References:           make(map[string]*ResourceReference),
		LocalQueries:         make(map[string]*Query),
		LocalControls:        make(map[string]*Control),
		LocalBenchmarks:      make(map[string]*Benchmark),
	}
}

func CreateWorkspaceResourceMapForMod(mod *Mod) *WorkspaceResourceMaps {
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
		Locals:               mod.Locals,
	}
	// if mod is not a default mod (i.e. if there is a mod.sp), add it into the resource maps
	if !mod.IsDefaultMod() {
		resourceMaps.Mods[mod.Name()] = mod
	}

	return resourceMaps
}

func CreateWorkspaceResourceMapForQueryProviders(queryProviders []QueryProvider) *WorkspaceResourceMaps {
	res := &WorkspaceResourceMaps{
		Mods:       make(map[string]*Mod),
		Queries:    make(map[string]*Query),
		Controls:   make(map[string]*Control),
		Benchmarks: make(map[string]*Benchmark),
		Variables:  make(map[string]*Variable),
		References: make(map[string]*ResourceReference),
	}
	for _, p := range queryProviders {
		res.addQueryProvider(p)
	}
	return res
}

func (m *WorkspaceResourceMaps) QueryProviders() []QueryProvider {
	res := make([]QueryProvider,
		len(m.Queries)+
			len(m.Controls)+
			len(m.DashboardCards)+
			len(m.DashboardCharts)+
			len(m.DashboardHierarchies)+
			len(m.DashboardInputs)+
			len(m.DashboardTables))

	idx := 0
	for _, p := range m.Queries {
		res[idx] = p
		idx++
	}
	for _, p := range m.Controls {
		res[idx] = p
		idx++
	}
	for _, p := range m.DashboardCards {
		res[idx] = p
		idx++
	}
	for _, p := range m.DashboardCharts {
		res[idx] = p
		idx++
	}
	for _, p := range m.DashboardHierarchies {
		res[idx] = p
		idx++
	}
	for _, p := range m.DashboardInputs {
		res[idx] = p
		idx++
	}
	for _, p := range m.DashboardTables {
		res[idx] = p
		idx++
	}
	return res
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

	for name := range other.Locals {
		if _, ok := m.Locals[name]; !ok {
			return false
		}
	}
	return true
}

func (m *WorkspaceResourceMaps) PopulateReferences() {
	m.References = make(map[string]*ResourceReference)

	resourceFunc := func(resource HclResource) (bool, error) {

		parsedName, _ := ParseResourceName(resource.Name())
		if helpers.StringSliceContains(ReferenceBlocks, parsedName.ItemType) {
			for _, ref := range resource.GetReferences() {
				m.References[ref.String()] = ref
			}
		}

		// continue walking
		return true, nil
	}
	m.WalkResources(resourceFunc)
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

func (m *WorkspaceResourceMaps) addQueryProvider(provider QueryProvider) {
	switch p := provider.(type) {
	case *Query:
		if p != nil {
			m.Queries[p.FullName] = p
		}
	case *Control:
		if p != nil {
			m.Controls[p.FullName] = p
		}
	case *DashboardCard:
		if p != nil {
			m.DashboardCards[p.FullName] = p
		}
	case *DashboardChart:
		if p != nil {
			m.DashboardCharts[p.FullName] = p
		}
	case *DashboardHierarchy:
		if p != nil {
			m.DashboardHierarchies[p.FullName] = p
		}
	case *DashboardInput:
		if p != nil {
			m.DashboardInputs[p.FullName] = p
		}
	}
}

// WalkResources calls resourceFunc for every resource in the mod
// if any resourceFunc returns false or an error, return immediately
func (m *WorkspaceResourceMaps) WalkResources(resourceFunc func(item HclResource) (bool, error)) error {
	for _, r := range m.Queries {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}

	}
	for _, r := range m.Controls {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Benchmarks {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Dashboards {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardContainers {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardCards {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardCharts {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardHierarchies {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardImages {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardInputs {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardTables {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardTexts {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Variables {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Locals {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	return nil
}
