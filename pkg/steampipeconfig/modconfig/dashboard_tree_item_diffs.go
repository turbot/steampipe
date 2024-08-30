package modconfig

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
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
	oldChildMap := make(map[string]ModTreeItem)
	newChildMap := make(map[string]ModTreeItem)

	oldChildren := old.GetChildren()
	newChildren := new.GetChildren()

	for i, child := range oldChildren {
		// check for child ordering
		if i < len(newChildren) && newChildren[i].Name() != child.Name() {
			d.AddPropertyDiff("Children")
		}
		oldChildMap[child.Name()] = child
	}
	for _, child := range newChildren {
		newChildMap[child.Name()] = child
	}

	for childName, prevChild := range oldChildMap {
		if child, existInNew := newChildMap[childName]; !existInNew {
			d.AddRemovedItem(childName)
		} else {
			// so this resource exists on old and new

			// TACTICAL
			// some child resources are not added to the mod but we must consider them for the diff
			var childDiff = &DashboardTreeItemDiffs{}
			switch t := child.(type) {
			case *DashboardWith:
				childDiff = t.Diff(prevChild.(*DashboardWith))
			case *DashboardNode:
				childDiff = t.Diff(prevChild.(*DashboardNode))
			case *DashboardEdge:
				childDiff = t.Diff(prevChild.(*DashboardEdge))
			}
			if childDiff.HasChanges() {
				d.AddPropertyDiff("Children")
			}

		}
	}
	for childName := range newChildMap {
		if _, existsInOld := oldChildMap[childName]; !existsInOld {
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
		} else if !lArgs.Equals(rArgs) {
			d.AddPropertyDiff("Args")
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
	if lwp, ok := l.(WithProvider); ok {
		rwp := r.(WithProvider)
		lWiths := lwp.GetWiths()
		rWiths := rwp.GetWiths()
		if len(lWiths) != len(rWiths) {
			d.AddPropertyDiff("With")
		} else {
			for i, lWith := range lWiths {
				if !lWith.Equals(rWiths[i]) {
					d.AddPropertyDiff("With")
				}
			}
		}

		// have BASE withs changed
		lbase := l.GetBase()
		rbase := r.GetBase()
		var lbaseWiths []*DashboardWith
		var rbaseWiths []*DashboardWith
		if lbase != nil {
			lbaseWiths = lbase.(WithProvider).GetWiths()
		}
		if rbase != nil {
			rbaseWiths = rbase.(WithProvider).GetWiths()
		}
		if len(lbaseWiths) != len(rbaseWiths) {
			d.AddPropertyDiff("With")
		} else {
			for i, lBaseWith := range lbaseWiths {
				if !lBaseWith.Equals(rbaseWiths[i]) {
					d.AddPropertyDiff("With")
				}
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
