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
	for _, benchmark := range sourceMod.Benchmarks {
		// add benchmark into control tree
		if err := m.addItemIntoResourceTree(benchmark); err != nil {
			return err
		}
		if len(benchmark.GetChildren()) == 0 {
			leafNodes = append(leafNodes, benchmark)
		}
	}
	for _, control := range sourceMod.Controls {
		if err := m.addItemIntoResourceTree(control); err != nil {
			return err
		}

		// controls cannot have children - all controls are leaves
		leafNodes = append(leafNodes, control)
	}
	// we add panels and reports into tree - as they may be children of Mod
	// any nested children of Container or Report (which may be Container or Panel) will already be in the tree)
	for _, panel := range sourceMod.Panels {
		if err := m.addItemIntoResourceTree(panel); err != nil {
			return err
		}
		// panels cannot have children
		leafNodes = append(leafNodes, panel)
	}

	for _, report := range sourceMod.Reports {
		if err := m.addItemIntoResourceTree(report); err != nil {
			return err
		}
		if len(report.GetChildren()) == 0 {
			leafNodes = append(leafNodes, report)
		}

	}
	for _, container := range sourceMod.Containers {
		if err := m.addItemIntoResourceTree(container); err != nil {
			return err
		}
		if len(container.GetChildren()) == 0 {
			leafNodes = append(leafNodes, container)
		}

	}
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

	case *Panel:
		name := r.Name()
		// check for dupes
		if _, ok := m.Panels[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.Panels[name] = r
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
			if _, ok := m.Containers[name]; ok {
				diags = append(diags, duplicateResourceDiagnostics(item))
				break
			} else {
				m.Containers[name] = r
			}
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
