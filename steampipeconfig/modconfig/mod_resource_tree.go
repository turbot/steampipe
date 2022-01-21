package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// BuildResourceTree builds the control tree structure by setting the parent property for each control and benchmar
// NOTE: this also builds the sorted benchmark list
func (m *Mod) BuildResourceTree(loadedDependencyMods ModMap) error {
	if err := m.addResourcesIntoTree(m); err != nil {
		return err
	}
	defer m.validateResourceTree()

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

func (m *Mod) addResourcesIntoTree(sourceMod *Mod) error {
	var leafNodes []ModTreeItem
	var err error

	resourceFunc := func(item HclResource) bool {
		if treeItem, ok := item.(ModTreeItem); ok {
			if err = sourceMod.addItemIntoResourceTree(treeItem); err != nil {
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
		item.AddParent(p)
		if p == m {
			m.children = append(m.children, item)
		}
	}

	return nil
}

func (m *Mod) AddResource(item HclResource) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch r := item.(type) {
	case *Query:
		name := r.Name()
		// check for dupes
		if _, ok := m.Queries[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		}
		m.Queries[name] = r

	case *Control:
		name := r.Name()
		// check for dupes
		if _, ok := m.Controls[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		}
		m.Controls[name] = r

	case *Benchmark:
		name := r.Name()
		// check for dupes
		if _, ok := m.Benchmarks[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.Benchmarks[name] = r
		}

	case *ReportContainer:
		name := r.Name()
		// report struct may either be a `report` or a `container`
		if r.IsReport() {
			// check for dupes
			if _, ok := m.Reports[name]; ok {
				diags = append(diags, duplicateResourceDiagnostics(item))
				break
			} else {
				m.Reports[name] = r
			}
		} else {
			// check for dupes
			if _, ok := m.ReportContainers[name]; ok {
				diags = append(diags, duplicateResourceDiagnostics(item))
				break
			} else {
				m.ReportContainers[name] = r
			}
		}
	case *ReportChart:
		name := r.Name()
		// check for dupes
		if _, ok := m.ReportCharts[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.ReportCharts[name] = r
		}
	case *ReportCounter:
		name := r.Name()
		// check for dupes
		if _, ok := m.ReportCounters[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.ReportCounters[name] = r
		}
	case *ReportControl:
		name := r.Name()
		// check for dupes
		if _, ok := m.ReportControls[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.ReportControls[name] = r
		}
	case *ReportImage:
		name := r.Name()
		// check for dupes
		if _, ok := m.ReportImages[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.ReportImages[name] = r
		}
	case *ReportTable:
		name := r.Name()
		// check for dupes
		if _, ok := m.ReportTables[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.ReportTables[name] = r
		}
	case *ReportText:
		name := r.Name()
		// check for dupes
		if _, ok := m.ReportTexts[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.ReportTexts[name] = r
		}

	case *Variable:
		// NOTE: add variable by unqualified name
		name := r.UnqualifiedName
		// check for dupes
		if _, ok := m.Variables[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.Variables[name] = r
		}

	case *Local:
		name := r.Name()
		// check for dupes
		if _, ok := m.Locals[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.Locals[name] = r
		}
	}
	return diags
}

func duplicateResourceDiagnostics(item HclResource) *hcl.Diagnostic {
	return &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("mod defines more than one resource named %s", item.Name()),
		Subject:  item.GetDeclRange(),
	}
}
