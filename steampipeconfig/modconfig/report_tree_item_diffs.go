package modconfig

// ReportTreeItemDiffs is a struct representing the differences between 2 ReportTreeItems (of same type)
type ReportTreeItemDiffs struct {
	Name              string
	Item              ModTreeItem
	ChangedProperties []string
	AddedItems        []string
	RemovedItems      []string
}

func (d *ReportTreeItemDiffs) AddPropertyDiff(propertyName string) {
	d.ChangedProperties = append(d.ChangedProperties, propertyName)
}

func (d *ReportTreeItemDiffs) AddAddedItem(name string) {
	d.AddedItems = append(d.AddedItems, name)
}

func (d *ReportTreeItemDiffs) AddRemovedItem(name string) {
	d.RemovedItems = append(d.RemovedItems, name)
}

func (d *ReportTreeItemDiffs) populateChildDiffs(old ModTreeItem, new ModTreeItem) {
	// build map of panel and report names
	oldChildMap := make(map[string]bool)
	newChildMap := make(map[string]bool)

	for _, child := range old.GetChildren() {
		oldChildMap[child.Name()] = true
	}
	for _, child := range new.GetChildren() {
		newChildMap[child.Name()] = true
	}

	for childName := range oldChildMap {
		if !newChildMap[childName] {
			d.AddRemovedItem(childName)
		}
	}
	for childName := range newChildMap {
		if !oldChildMap[childName] {
			d.AddAddedItem(childName)
		}
	}

}

func (d *ReportTreeItemDiffs) HasChanges() bool {
	return len(d.ChangedProperties)+
		len(d.AddedItems)+
		len(d.RemovedItems) > 0
}
