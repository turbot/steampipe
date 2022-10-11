package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/utils"
)

// BuildResourceTree builds the control tree structure by setting the parent property for each control and benchmark
// NOTE: this also builds the sorted benchmark list
func (m *Mod) BuildResourceTree(loadedDependencyMods ModMap) (err error) {
	utils.LogTime("BuildResourceTree start")
	defer utils.LogTime("BuildResourceTree end")
	defer func() {
		if err == nil {
			err = m.validateResourceTree()
		}
	}()

	if err := m.addResourcesIntoTree(m); err != nil {
		return err
	}

	if !m.HasDependentMods() {
		return nil
	}
	// add dependent mods into tree
	for _, requiredMod := range m.Require.Mods {
		// find this mod in installed dependency mods
		depMod, ok := loadedDependencyMods[requiredMod.Name]
		if !ok {
			return fmt.Errorf("dependency mod %s is not loaded", requiredMod.Name)
		}
		if err := m.addResourcesIntoTree(depMod); err != nil {
			return err
		}
	}

	return nil
}

// add all resource in sourceMod into _our_ resource tree
func (m *Mod) addResourcesIntoTree(sourceMod *Mod) error {
	var leafNodes []ModTreeItem
	var err error

	resourceFunc := func(item HclResource) (bool, error) {
		if treeItem, ok := item.(ModTreeItem); ok {
			// NOTE: add resource into _our_ resource tree, i.e. mod 'm'
			if err = m.addItemIntoResourceTree(treeItem); err != nil {
				// stop walking
				return false, nil
			}
			if len(treeItem.GetChildren()) == 0 {
				leafNodes = append(leafNodes, treeItem)
			}
		}
		// continue walking
		return true, nil
	}

	// iterate through all resources in source mod
	sourceMod.WalkResources(resourceFunc)

	// now initialise all Paths properties
	for _, l := range leafNodes {
		l.SetPaths()
	}

	return nil
}

func (m *Mod) addItemIntoResourceTree(item ModTreeItem) error {
	for _, p := range m.getParents(item) {
		// if we are the parent, add as a child
		if err := item.AddParent(p); err != nil {
			return err
		}
		if p == m {
			m.children = append(m.children, item)
		}
	}

	return nil
}

// check whether a resource with the same name has already been added to the mod
// (it is possible to add the same resource to a mod more than once as the parent resource
// may have dependency errors and so be decoded again)
func checkForDuplicate(existing, new HclResource) hcl.Diagnostics {
	if existing.GetDeclRange().String() == new.GetDeclRange().String() {
		// decl range is the same - this is the same resource - allowable
		return nil
	}
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("Mod defines more than one resource named '%s'", new.Name()),
		Detail:   fmt.Sprintf("\n- %s\n- %s", existing.GetDeclRange(), new.GetDeclRange()),
	}}
}

func (m *Mod) AddResource(item HclResource) hcl.Diagnostics {
	return m.ResourceMaps.AddResource(item)
}
