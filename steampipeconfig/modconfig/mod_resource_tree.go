package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// BuildResourceTree builds the control tree structure by setting the parent property for each control and benchmark
// NOTE: this also builds the sorted benchmark list
func (m *Mod) BuildResourceTree(loadedDependencyMods ModMap) (err error) {
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

	resourceFunc := func(item HclResource) bool {
		if treeItem, ok := item.(ModTreeItem); ok {
			// NOTE: add resource into _our_ resource tree, i.e. mod 'm'
			if err = m.addItemIntoResourceTree(treeItem); err != nil {
				// stop walking
				return false
			}
			if len(treeItem.GetChildren()) == 0 {
				leafNodes = append(leafNodes, treeItem)
			}
		}
		// continue walking
		return true
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

// check whether a resource with the same name has already been added
// (it is possible to add the same resource to a mod more than once as the parent resource
// may have dependency errors and so be decoded again)
func checkForDuplicate(existing, new HclResource) hcl.Diagnostics {
	if existing.GetDeclRange().String() == new.GetDeclRange().String() {
		// decl range is the same - this is the same resource - allowable
		return nil
	}
	return duplicateResourceDiagnostics(new)
}

func (m *Mod) AddResource(item HclResource) hcl.Diagnostics {
	// TODO generics would make this must more compact
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

	case *ReportContainer:
		name := r.Name()
		// report struct may either be a `report` or a `container`
		if r.IsReport() {
			if existing, ok := m.Reports[name]; ok {
				diags = append(diags, checkForDuplicate(existing, item)...)
				break
			}
			m.Reports[name] = r

		} else {
			if existing, ok := m.ReportContainers[name]; ok {
				diags = append(diags, checkForDuplicate(existing, item)...)
				break
			}
			m.ReportContainers[name] = r

		}
	case *ReportCard:
		name := r.Name()
		if existing, ok := m.ReportCards[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		} else {
			m.ReportCards[name] = r
		}
	case *ReportChart:
		name := r.Name()
		if existing, ok := m.ReportCharts[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.ReportCharts[name] = r

	case *ReportHierarchy:
		name := r.Name()
		if existing, ok := m.ReportHierarchies[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.ReportHierarchies[name] = r

	case *ReportImage:
		name := r.Name()
		if existing, ok := m.ReportImages[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.ReportImages[name] = r

	case *ReportInput:
		name := r.Name()
		if existing, ok := m.ReportInputs[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.ReportInputs[name] = r

	case *ReportTable:
		name := r.Name()
		if existing, ok := m.ReportTables[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.ReportTables[name] = r

	case *ReportText:
		name := r.Name()
		if existing, ok := m.ReportTexts[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.ReportTexts[name] = r

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

func duplicateResourceDiagnostics(item HclResource) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("mod defines more than one resource named %s", item.Name()),
		Subject:  item.GetDeclRange(),
	}}
}
