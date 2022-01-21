package modconfig

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
