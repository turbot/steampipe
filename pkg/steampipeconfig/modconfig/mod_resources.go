package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/utils"
)

// ResourceMaps is a struct containing maps of all mod resource types
// This is provided to avoid db needing to reference workspace package
type ResourceMaps struct {
	// the parent mod
	Mod *Mod

	// all mods (including deps)
	Benchmarks            map[string]*Benchmark
	Controls              map[string]*Control
	Dashboards            map[string]*Dashboard
	DashboardCategories   map[string]*DashboardCategory
	DashboardCards        map[string]*DashboardCard
	DashboardCharts       map[string]*DashboardChart
	DashboardContainers   map[string]*DashboardContainer
	DashboardEdges        map[string]*DashboardEdge
	DashboardFlows        map[string]*DashboardFlow
	DashboardGraphs       map[string]*DashboardGraph
	DashboardHierarchies  map[string]*DashboardHierarchy
	DashboardImages       map[string]*DashboardImage
	DashboardInputs       map[string]map[string]*DashboardInput
	DashboardTables       map[string]*DashboardTable
	DashboardTexts        map[string]*DashboardText
	DashboardNodes        map[string]*DashboardNode
	GlobalDashboardInputs map[string]*DashboardInput
	Locals                map[string]*Local
	Mods                  map[string]*Mod
	Queries               map[string]*Query
	References            map[string]*ResourceReference
	// map of snapshot paths, keyed by snapshot name
	Snapshots map[string]string
	Variables map[string]*Variable
}

func NewModResources(mod *Mod) *ResourceMaps {
	res := emptyModResources()
	res.Mod = mod
	res.Mods[mod.Name()] = mod
	return res
}

func NewSourceSnapshotModResources(snapshotPaths []string) *ResourceMaps {
	res := emptyModResources()
	res.AddSnapshots(snapshotPaths)
	return res
}

func emptyModResources() *ResourceMaps {
	return &ResourceMaps{
		Controls:              make(map[string]*Control),
		Benchmarks:            make(map[string]*Benchmark),
		Dashboards:            make(map[string]*Dashboard),
		DashboardCards:        make(map[string]*DashboardCard),
		DashboardCharts:       make(map[string]*DashboardChart),
		DashboardContainers:   make(map[string]*DashboardContainer),
		DashboardEdges:        make(map[string]*DashboardEdge),
		DashboardFlows:        make(map[string]*DashboardFlow),
		DashboardGraphs:       make(map[string]*DashboardGraph),
		DashboardHierarchies:  make(map[string]*DashboardHierarchy),
		DashboardImages:       make(map[string]*DashboardImage),
		DashboardInputs:       make(map[string]map[string]*DashboardInput),
		DashboardTables:       make(map[string]*DashboardTable),
		DashboardTexts:        make(map[string]*DashboardText),
		DashboardNodes:        make(map[string]*DashboardNode),
		DashboardCategories:   make(map[string]*DashboardCategory),
		GlobalDashboardInputs: make(map[string]*DashboardInput),
		Locals:                make(map[string]*Local),
		Mods:                  make(map[string]*Mod),
		Queries:               make(map[string]*Query),
		References:            make(map[string]*ResourceReference),
		Snapshots:             make(map[string]string),
		Variables:             make(map[string]*Variable),
	}
}

// ModResourcesForQueries creates a ResourceMaps object containing just the specified queries
// This is used to just create necessary prepared statements when executing batch queries
func ModResourcesForQueries(queryProviders []QueryProvider, mod *Mod) *ResourceMaps {
	res := NewModResources(mod)
	for _, p := range queryProviders {
		res.addControlOrQuery(p)
	}
	return res
}

// QueryProviders returns a slice of all QueryProviders
func (m *ResourceMaps) QueryProviders() []QueryProvider {
	res := make([]QueryProvider, m.queryProviderCount())
	idx := 0
	f := func(item HclResource) (bool, error) {
		if queryProvider, ok := item.(QueryProvider); ok {
			res[idx] = queryProvider
			idx++
		}
		return true, nil
	}

	m.WalkResources(f)

	return res
}

func (m *ResourceMaps) Equals(other *ResourceMaps) bool {
	//TODO use cmp.Equals or similar
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

	for name, flows := range m.DashboardFlows {
		if otherFlow, ok := other.DashboardFlows[name]; !ok {
			return false
		} else if !flows.Equals(otherFlow) {
			return false
		}
	}
	for name := range other.DashboardFlows {
		if _, ok := m.DashboardFlows[name]; !ok {
			return false
		}
	}

	for name, flows := range m.DashboardGraphs {
		if otherFlow, ok := other.DashboardGraphs[name]; !ok {
			return false
		} else if !flows.Equals(otherFlow) {
			return false
		}
	}
	for name := range other.DashboardGraphs {
		if _, ok := m.DashboardGraphs[name]; !ok {
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

	for name := range other.DashboardNodes {
		if _, ok := m.DashboardNodes[name]; !ok {
			return false
		}
	}

	for name := range other.DashboardEdges {
		if _, ok := m.DashboardEdges[name]; !ok {
			return false
		}
	}
	for name := range other.DashboardCategories {
		if _, ok := m.DashboardCategories[name]; !ok {
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

	for name, input := range m.GlobalDashboardInputs {
		if otherInput, ok := other.GlobalDashboardInputs[name]; !ok {
			return false
		} else if !input.Equals(otherInput) {
			return false
		}
	}
	for name := range other.DashboardInputs {
		if _, ok := m.DashboardInputs[name]; !ok {
			return false
		}
	}

	for dashboardName, inputsForDashboard := range m.DashboardInputs {
		if otherInputsForDashboard, ok := other.DashboardInputs[dashboardName]; !ok {
			return false
		} else {

			for name, input := range inputsForDashboard {
				if otherInput, ok := otherInputsForDashboard[name]; !ok {
					return false
				} else if !input.Equals(otherInput) {
					return false
				}
			}
			for name := range otherInputsForDashboard {
				if _, ok := inputsForDashboard[name]; !ok {
					return false
				}
			}

		}
	}
	for name := range other.DashboardInputs {
		if _, ok := m.DashboardInputs[name]; !ok {
			return false
		}
	}

	for name, table := range m.DashboardTables {
		if otherTable, ok := other.DashboardTables[name]; !ok {
			return false
		} else if !table.Equals(otherTable) {
			return false
		}
	}
	for name, category := range m.DashboardCategories {
		if otherCategory, ok := other.DashboardCategories[name]; !ok {
			return false
		} else if !category.Equals(otherCategory) {
			return false
		}
	}
	for name := range other.DashboardTables {
		if _, ok := m.DashboardTables[name]; !ok {
			return false
		}
	}

	for name, text := range m.DashboardTexts {
		if otherText, ok := other.DashboardTexts[name]; !ok {
			return false
		} else if !text.Equals(otherText) {
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

func (m *ResourceMaps) PopulateReferences() {
	m.References = make(map[string]*ResourceReference)

	resourceFunc := func(resource HclResource) (bool, error) {
		if resourceWithMetadata, ok := resource.(ResourceWithMetadata); ok {
			for _, ref := range resourceWithMetadata.GetReferences() {
				m.References[ref.String()] = ref
			}
		}

		// continue walking
		return true, nil
	}
	m.WalkResources(resourceFunc)
}

func (m *ResourceMaps) Empty() bool {
	return len(m.Mods)+
		len(m.Queries)+
		len(m.Controls)+
		len(m.Benchmarks)+
		len(m.Variables)+
		len(m.Dashboards)+
		len(m.DashboardContainers)+
		len(m.DashboardCards)+
		len(m.DashboardCharts)+
		len(m.DashboardFlows)+
		len(m.DashboardGraphs)+
		len(m.DashboardHierarchies)+
		len(m.DashboardNodes)+
		len(m.DashboardEdges)+
		len(m.DashboardCategories)+
		len(m.DashboardImages)+
		len(m.DashboardInputs)+
		len(m.DashboardTables)+
		len(m.DashboardTexts)+
		len(m.References) == 0
}

// this is used to create an optimized ResourceMaps containing only the queries which will be run
func (m *ResourceMaps) addControlOrQuery(provider QueryProvider) {
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

// WalkResources calls resourceFunc for every resource in the mod
// if any resourceFunc returns false or an error, return immediately
func (m *ResourceMaps) WalkResources(resourceFunc func(item HclResource) (bool, error)) error {
	for _, r := range m.Benchmarks {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Controls {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Dashboards {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardCards {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardCategories {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardCharts {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardContainers {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardEdges {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardFlows {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardGraphs {
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
	for _, inputsForDashboard := range m.DashboardInputs {
		for _, r := range inputsForDashboard {
			if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
				return err
			}
		}
	}
	for _, r := range m.DashboardNodes {
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
	for _, r := range m.GlobalDashboardInputs {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Locals {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Queries {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	// we cannot walk source snapshots as they are not a HclResource
	for _, r := range m.Variables {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}

	return nil
}

func (m *ResourceMaps) AddResource(item HclResource) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch r := item.(type) {
	case *Query:
		name := r.Name()
		if existing, ok := m.Queries[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Queries[name] = r

	case *Control:
		name := r.Name()
		if existing, ok := m.Controls[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Controls[name] = r

	case *Benchmark:
		name := r.Name()
		if existing, ok := m.Benchmarks[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Benchmarks[name] = r

	case *Dashboard:
		name := r.Name()
		if existing, ok := m.Dashboards[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Dashboards[name] = r

	case *DashboardContainer:
		name := r.Name()
		if existing, ok := m.DashboardContainers[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardContainers[name] = r

	case *DashboardCard:
		name := r.Name()
		if existing, ok := m.DashboardCards[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		} else {
			m.DashboardCards[name] = r
		}

	case *DashboardChart:
		name := r.Name()
		if existing, ok := m.DashboardCharts[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardCharts[name] = r

	case *DashboardFlow:
		name := r.Name()
		if existing, ok := m.DashboardFlows[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardFlows[name] = r

	case *DashboardGraph:
		name := r.Name()
		if existing, ok := m.DashboardGraphs[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardGraphs[name] = r

	case *DashboardHierarchy:
		name := r.Name()
		if existing, ok := m.DashboardHierarchies[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardHierarchies[name] = r

	case *DashboardNode:
		name := r.Name()
		if existing, ok := m.DashboardNodes[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardNodes[name] = r

	case *DashboardEdge:
		name := r.Name()
		if existing, ok := m.DashboardEdges[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardEdges[name] = r
	case *DashboardCategory:
		name := r.Name()
		if existing, ok := m.DashboardCategories[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardCategories[name] = r

	case *DashboardImage:
		name := r.Name()
		if existing, ok := m.DashboardImages[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardImages[name] = r

	case *DashboardInput:
		// if input has a dashboard asssigned, add to DashboardInputs
		name := r.Name()
		if dashboardName := r.DashboardName; dashboardName != "" {
			inputsForDashboard := m.DashboardInputs[dashboardName]
			if inputsForDashboard == nil {
				inputsForDashboard = make(map[string]*DashboardInput)
				m.DashboardInputs[dashboardName] = inputsForDashboard
			}
			// no need to check for dupes as we have already checked before adding the input to th m od
			inputsForDashboard[name] = r
			break
		}

		// so Dashboard Input must be global
		if existing, ok := m.GlobalDashboardInputs[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.GlobalDashboardInputs[name] = r

	case *DashboardTable:
		name := r.Name()
		if existing, ok := m.DashboardTables[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardTables[name] = r

	case *DashboardText:
		name := r.Name()
		if existing, ok := m.DashboardTexts[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardTexts[name] = r

	case *Variable:
		// NOTE: add variable by unqualified name
		name := r.UnqualifiedName
		if existing, ok := m.Variables[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Variables[name] = r

	case *Local:
		name := r.Name()
		if existing, ok := m.Locals[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Locals[name] = r

	}
	return diags
}

func (m *ResourceMaps) AddSnapshots(snapshotPaths []string) {
	for _, snapshotPath := range snapshotPaths {
		snapshotName := fmt.Sprintf("snapshot.%s", utils.FilenameNoExtension(snapshotPath))
		m.Snapshots[snapshotName] = snapshotPath
	}
}

func (m *ResourceMaps) Merge(others []*ResourceMaps) *ResourceMaps {
	res := NewModResources(m.Mod)
	sourceMaps := append([]*ResourceMaps{m}, others...)

	for _, source := range sourceMaps {
		for k, v := range source.Benchmarks {
			res.Benchmarks[k] = v
		}
		for k, v := range source.Controls {
			res.Controls[k] = v
		}
		for k, v := range source.Dashboards {
			res.Dashboards[k] = v
		}
		for k, v := range source.DashboardContainers {
			res.DashboardContainers[k] = v
		}
		for k, v := range source.DashboardCards {
			res.DashboardCards[k] = v
		}
		for k, v := range source.DashboardCategories {
			res.DashboardCategories[k] = v
		}
		for k, v := range source.DashboardCharts {
			res.DashboardCharts[k] = v
		}
		for k, v := range source.DashboardEdges {
			res.DashboardEdges[k] = v
		}
		for k, v := range source.DashboardFlows {
			res.DashboardFlows[k] = v
		}
		for k, v := range source.DashboardGraphs {
			res.DashboardGraphs[k] = v
		}
		for k, v := range source.DashboardHierarchies {
			res.DashboardHierarchies[k] = v
		}
		for k, v := range source.DashboardNodes {
			res.DashboardNodes[k] = v
		}
		for k, v := range source.DashboardImages {
			res.DashboardImages[k] = v
		}
		for k, v := range source.DashboardInputs {
			res.DashboardInputs[k] = v
		}
		for k, v := range source.DashboardTables {
			res.DashboardTables[k] = v
		}
		for k, v := range source.DashboardTexts {
			res.DashboardTexts[k] = v
		}
		for k, v := range source.GlobalDashboardInputs {
			res.GlobalDashboardInputs[k] = v
		}
		for k, v := range source.Locals {
			res.Locals[k] = v
		}
		for k, v := range source.Mods {
			res.Mods[k] = v
		}
		for k, v := range source.Queries {
			res.Queries[k] = v
		}
		for k, v := range source.Snapshots {
			res.Snapshots[k] = v
		}
		for k, v := range source.Variables {
			// NOTE: only include variables from root mod  - we add in the others separately
			if v.Mod.FullName == m.Mod.FullName {
				res.Variables[k] = v
			}
		}
	}

	return res
}

func (m *ResourceMaps) queryProviderCount() int {
	numDashboardInputs := 0
	for _, inputs := range m.DashboardInputs {
		numDashboardInputs += len(inputs)
	}

	numItems :=
		len(m.Controls) +
			len(m.DashboardCards) +
			len(m.DashboardCharts) +
			len(m.DashboardEdges) +
			len(m.DashboardFlows) +
			len(m.DashboardGraphs) +
			len(m.DashboardHierarchies) +
			len(m.DashboardImages) +
			numDashboardInputs +
			len(m.DashboardNodes) +
			len(m.DashboardTables) +
			len(m.GlobalDashboardInputs) +
			len(m.Queries)
	return numItems
}
