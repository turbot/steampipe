package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// BuildResourceTree builds the control tree structure by setting the parent property for each control and benchmar
// NOTE: this also builds the sorted benchmark list
func (m *Mod) BuildResourceTree(loadedDependencyMods ModMap) error {
	m.buildFlatChilden()
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
	for _, benchmark := range sourceMod.Benchmarks {
		// add benchmark into control tree
		if err := m.addItemIntoResourceTree(benchmark); err != nil {
			return err
		}
	}
	for _, control := range sourceMod.Controls {
		if err := m.addItemIntoResourceTree(control); err != nil {
			return err
		}
	}
	for _, panel := range sourceMod.Panels {
		if err := m.addItemIntoResourceTree(panel); err != nil {
			return err
		}
	}
	for _, report := range sourceMod.Reports {
		if err := m.addItemIntoResourceTree(report); err != nil {
			return err
		}
	}
	// now initialise all Paths properties
	setPaths(m)
	return nil
}

func setPaths(i ModTreeItem) {
	i.SetPaths()
	for _, c := range i.GetChildren() {
		setPaths(c)
	}
}

func (m *Mod) addItemIntoResourceTree(item ModTreeItem) error {
	for _, p := range m.getParents(item) {
		item.AddParent(p)
		p.AddChild(item)
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

	case *Report:
		name := r.Name()
		// check for dupes
		if _, ok := m.Reports[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item))
			break
		} else {
			m.Reports[name] = r
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
