package modconfig

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/utils"
	"golang.org/x/exp/maps"
)

// DashboardTreeItemDiffs is a struct representing the differences between 2 DashboardTreeItems (of same type)
type DashboardTreeItemDiffs struct {
	Name              string
	Item              ModTreeItem
	ChangedProperties []string
	AddedItems        []string
	RemovedItems      []string
}

func (d *DashboardTreeItemDiffs) AddPropertyDiff(propertyName string) {
	if !helpers.StringSliceContains(d.ChangedProperties, propertyName) {
		d.ChangedProperties = append(d.ChangedProperties, propertyName)
	}
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

	oldChildren := old.GetChildren()
	newChildren := new.GetChildren()

	for i, child := range oldChildren {
		// check for child ordering
		if i < len(newChildren) && newChildren[i].Name() != child.Name() {
			d.AddPropertyDiff("Children")
		}
		oldChildMap[child.Name()] = true
	}
	for _, child := range newChildren {
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
		for i, lParam := range lParams {
			if !lParam.Equals(rParams[i]) {
				d.AddPropertyDiff("Params")
			}
		}
	}

	// with
	lWiths := l.GetWiths()
	rWiths := r.GetWiths()
	if len(lWiths) != len(rWiths) {
		d.AddPropertyDiff("With")
	} else {
		for i, lWith := range lWiths {
			if !lWith.Equals(rWiths[i]) {
				d.AddPropertyDiff("With")
			}
		}
	}

}

func (d *DashboardTreeItemDiffs) dashboardLeafNodeDiff(l DashboardLeafNode, r DashboardLeafNode) {
	if l.Name() != r.Name() {
		d.AddPropertyDiff("Name")
	}
	if l.GetTitle() != r.GetTitle() {
		d.AddPropertyDiff("Title")
	}
	if l.GetWidth() != r.GetWidth() {
		d.AddPropertyDiff("Width")
	}
	if l.GetDisplay() != r.GetDisplay() {
		d.AddPropertyDiff("Display")
	}
	if l.GetDocumentation() != r.GetDocumentation() {
		d.AddPropertyDiff("Documentation")
	}
	if l.GetType() != r.GetType() {
		d.AddPropertyDiff("Type")
	}
	if !maps.Equal(l.GetTags(), r.GetTags()) {
		d.AddPropertyDiff("Tags")
	}
}
