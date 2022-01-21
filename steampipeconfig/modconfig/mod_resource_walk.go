package modconfig

import "fmt"

// WalkResources calls resourceFunc for every resource in the mod
// if any resourceFunc returns false, return immediately
func (m *Mod) WalkResources(resourceFunc func(item HclResource) bool) {
	for _, r := range m.Queries {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.Controls {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.Benchmarks {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.Reports {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.ReportContainers {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.ReportCharts {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.ReportControls {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.ReportCounters {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.ReportHierarchies {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.ReportImages {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.ReportTables {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.ReportTexts {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.Variables {
		if !resourceFunc(r) {
			return
		}
	}
	for _, r := range m.Locals {
		if !resourceFunc(r) {
			return
		}
	}
}

// get the parent item for this ModTreeItem
func (m *Mod) getParents(item ModTreeItem) []ModTreeItem {
	var parents []ModTreeItem

	resourceFunc := func(parent HclResource) bool {
		if treeItem, ok := parent.(ModTreeItem); ok {
			for _, child := range treeItem.GetChildren() {
				if child.Name() == item.Name() {
					parents = append(parents, treeItem)
				}
			}
		}
		// continue walking
		return true
	}
	m.WalkResources(resourceFunc)

	// if this item has no parents and is a child of the mod, set the mod as parent
	if len(parents) == 0 && m.containsResource(item.Name()) {
		parents = []ModTreeItem{m}

	}
	return parents
}

// does the mod contain a resource with this name?
func (m *Mod) containsResource(childName string) bool {
	var res bool

	resourceFunc := func(item HclResource) bool {
		if item.Name() == childName {
			res = true
			// break out of resource walk
			return false
		}
		// continue walking
		return true
	}
	m.WalkResources(resourceFunc)
	return res
}

func (m *Mod) GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool) {
	longName := fmt.Sprintf("%s.%s.%s", m.ShortName, parsedName.ItemType, parsedName.Name)

	switch parsedName.ItemType {
	case BlockTypeBenchmark:
		resource, found = m.Benchmarks[longName]
	case BlockTypeControl:
		resource, found = m.Controls[longName]
		if !found {
			resource, found = m.ReportControls[longName]
		}
	case BlockTypeReport:
		resource, found = m.Reports[longName]
	case BlockTypeContainer:
		resource, found = m.ReportContainers[longName]
	case BlockTypeChart:
		resource, found = m.ReportCharts[longName]
	case BlockTypeCounter:
		resource, found = m.ReportCounters[longName]
	case BlockTypeHierarchy:
		resource, found = m.ReportHierarchies[longName]
	case BlockTypeImage:
		resource, found = m.ReportImages[longName]
	case BlockTypeTable:
		resource, found = m.ReportTables[longName]
	case BlockTypeText:
		resource, found = m.ReportTexts[longName]
	}
	return resource, found

}
