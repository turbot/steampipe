package modconfig

// get the parent item for this ModTreeItem
func (m *Mod) getParents(item ModTreeItem) []ModTreeItem {
	var parents []ModTreeItem

	resourceFunc := func(parent HclResource) (bool, error) {
		if treeItem, ok := parent.(ModTreeItem); ok {
			for _, child := range treeItem.GetChildren() {
				if child.Name() == item.Name() {
					parents = append(parents, treeItem)
				}
			}
		}
		// continue walking
		return true, nil
	}
	m.ResourceMaps.WalkResources(resourceFunc)

	// if this item has no parents and is a child of the mod, set the mod as parent
	if len(parents) == 0 && m.containsResource(item.Name()) {
		parents = []ModTreeItem{m}

	}
	return parents
}

// does the mod contain a resource with this name?
func (m *Mod) containsResource(childName string) bool {
	var res bool

	resourceFunc := func(item HclResource) (bool, error) {
		if item.Name() == childName {
			res = true
			// break out of resource walk
			return false, nil
		}
		// continue walking
		return true, nil
	}
	m.ResourceMaps.WalkResources(resourceFunc)
	return res
}
