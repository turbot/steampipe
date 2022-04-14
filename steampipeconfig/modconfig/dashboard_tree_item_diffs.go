package modconfig

import "github.com/turbot/steampipe/utils"

// DashboardTreeItemDiffs is a struct representing the differences between 2 DashboardTreeItems (of same type)
type DashboardTreeItemDiffs struct {
	Name              string
	Item              ModTreeItem
	ChangedProperties []string
	AddedItems        []string
	RemovedItems      []string
}

func (d *DashboardTreeItemDiffs) AddPropertyDiff(propertyName string) {
	d.ChangedProperties = append(d.ChangedProperties, propertyName)
}

func (d *DashboardTreeItemDiffs) AddAddedItem(name string) {
	d.AddedItems = append(d.AddedItems, name)
}

func (d *DashboardTreeItemDiffs) AddRemovedItem(name string) {
	d.RemovedItems = append(d.RemovedItems, name)
}

func (d *DashboardTreeItemDiffs) populateChildDiffs(old ModTreeItem, new ModTreeItem) {
	// build map of child names
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

func (d *DashboardTreeItemDiffs) HasChanges() bool {
	return len(d.ChangedProperties)+
		len(d.AddedItems)+
		len(d.RemovedItems) > 0
}

func (d *DashboardTreeItemDiffs) queryProviderDiff(l QueryProvider, r QueryProvider) {
	// sql
	if !utils.SafeStringsEqual(l.GetSQL(), r.GetSQL()) {
		d.AddPropertyDiff("SQL")
	}

	// args
	if lArgs := l.GetArgs(); lArgs == nil {
		if r.GetArgs() != nil {
			d.AddPropertyDiff("Args")
		}
	} else {
		// we have args
		if rArgs := r.GetArgs(); rArgs == nil {
			d.AddPropertyDiff("Args")
		} else {
			if !lArgs.Equals(rArgs) {
				d.AddPropertyDiff("Args")
			}
		}
	}

	// query
	if lQuery := l.GetQuery(); lQuery == nil {
		if r.GetQuery() != nil {
			d.AddPropertyDiff("Query")
		}
	} else {
		// we have query
		if rQuery := r.GetQuery(); rQuery == nil {
			d.AddPropertyDiff("Query")
		} else {
			if !lQuery.Equals(rQuery) {
				d.AddPropertyDiff("Query")
			}
		}
	}

	// params
	lParams := l.GetParams()
	rParams := r.GetParams()
	if len(lParams) != len(rParams) {
		d.AddPropertyDiff("Params")
	} else {
		for i, p := range lParams {
			if !p.Equals(rParams[i]) {
				d.AddPropertyDiff("Params")
			}
		}
	}

}

func (d *DashboardTreeItemDiffs) dashboardLeafNodeDiff(l DashboardLeafNode, r DashboardLeafNode) {
	if !utils.SafeStringsEqual(l.Name(), r.Name()) {
		d.AddPropertyDiff("Name")
	}
	if !utils.SafeStringsEqual(l.GetTitle(), r.GetTitle()) {
		d.AddPropertyDiff("Title")
	}
	if l.GetWidth() != r.GetWidth() {
		d.AddPropertyDiff("Width")
	}
	if !utils.SafeStringsEqual(l.GetDisplay(), r.GetDisplay()) {
		d.AddPropertyDiff("Display")
	}
}
