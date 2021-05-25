package modconfig

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"

	"github.com/turbot/go-kit/helpers"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
)

// mod name used if a default mod is created for a workspace which does not define one explicitly
const defaultModName = "local"

// Mod is a struct representing a Mod resource
type Mod struct {
	ShortName string `hcl:"name,label"`
	FullName  string `cty:"name"`

	// attributes
	Categories    *[]string          `cty:"categories" hcl:"categories" column:"categories,jsonb"`
	Color         *string            `cty:"color" hcl:"color" column:"color,text"`
	Description   *string            `cty:"description" hcl:"description" column:"description,text"`
	Documentation *string            `cty:"documentation" hcl:"documentation" column:"documentation,text"`
	Icon          *string            `cty:"icon" hcl:"icon" column:"icon,text"`
	Tags          *map[string]string `cty:"tags" hcl:"tags" column:"tags,jsonb"`
	Title         *string            `cty:"title" hcl:"title" column:"title,text"`

	// list of all block referenced by the resource
	References []string `column:"refs,jsonb"`

	// blocks
	Requires  *Requires  `hcl:"requires,block"`
	OpenGraph *OpenGraph `hcl:"opengraph,block"`

	// TODO do we need this?
	Version *string

	Queries    map[string]*Query
	Controls   map[string]*Control
	Benchmarks map[string]*Benchmark
	// list of benchmark names, sorted alphabetically
	benchmarksOrdered []string

	ModPath   string
	DeclRange hcl.Range

	children []ControlTreeItem
	metadata *ResourceMetadata
}

func (m *Mod) ParseRequiredPluginVersions() error {
	if m.Requires != nil {
		requiredPluginVersions := m.Requires.Plugins

		for _, v := range requiredPluginVersions {
			err := v.parseProperties()
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func (m *Mod) CtyValue() (cty.Value, error) {
	return getCtyValue(m)
}

func NewMod(shortName, modPath string, defRange hcl.Range) *Mod {
	return &Mod{
		ShortName:  shortName,
		FullName:   fmt.Sprintf("mod.%s", shortName),
		Queries:    make(map[string]*Query),
		Controls:   make(map[string]*Control),
		Benchmarks: make(map[string]*Benchmark),
		ModPath:    modPath,
		DeclRange:  defRange,
	}
}

// CreateDefaultMod creates a default mod created for a workspace with no mod definition
func CreateDefaultMod(modPath string) *Mod {
	m := NewMod(defaultModName, modPath, hcl.Range{})
	folderName := filepath.Base(modPath)
	m.Title = &folderName
	return m
}

// IsDefaultMod returns whether this mod is a default mod created for a workspace with no mod definition
func (m *Mod) IsDefaultMod() bool {
	return m.ShortName == defaultModName
}

func (m *Mod) String() string {
	if m == nil {
		return ""
	}
	// build ordered list of query names
	var queryNames []string
	for name := range m.Queries {
		queryNames = append(queryNames, name)
	}
	sort.Strings(queryNames)

	var queryStrings []string
	for _, name := range queryNames {
		queryStrings = append(queryStrings, m.Queries[name].String())
	}
	// build ordered list of control names
	var controlNames []string
	for name := range m.Controls {
		controlNames = append(controlNames, name)
	}
	sort.Strings(controlNames)

	var controlStrings []string
	for _, name := range controlNames {
		controlStrings = append(controlStrings, m.Controls[name].String())
	}
	// build ordered list of control group names
	var benchmarkNames []string
	for name := range m.Benchmarks {
		benchmarkNames = append(benchmarkNames, name)
	}
	sort.Strings(benchmarkNames)

	var benchmarkStrings []string
	for _, name := range benchmarkNames {
		benchmarkStrings = append(benchmarkStrings, m.Benchmarks[name].String())
	}

	versionString := ""
	if m.Version != nil {
		versionString = fmt.Sprintf("\nVersion: %s", types.SafeString(m.Version))
	}
	return fmt.Sprintf(`Name: %s
Title: %s
Description: %s %s
Queries: 
%s
Controls: 
%s
Control Groups: 
%s`,
		m.FullName,
		types.SafeString(m.Title),
		types.SafeString(m.Description),
		versionString,
		strings.Join(queryStrings, "\n"),
		strings.Join(controlStrings, "\n"),
		strings.Join(benchmarkStrings, "\n"),
	)
}

// IsControlTreeItem implements ControlTreeItem
// (mod is always top of the tree)
func (m *Mod) IsControlTreeItem() {}

// BuildControlTree builds the control tree structure by setting the parent property for each control and benchmar
// NOTE: this also builds the sorted benchmark list
func (m *Mod) BuildControlTree() error {
	// build sorted list of benchmarks
	m.benchmarksOrdered = make([]string, len(m.Benchmarks))
	idx := 0
	for name, benchmark := range m.Benchmarks {
		// save this benchmark name
		m.benchmarksOrdered[idx] = name
		idx++

		// add benchmark into control tree
		if err := m.addItemIntoControlTree(benchmark); err != nil {
			return err
		}
	}
	// now sort the benchmark names
	sort.Strings(m.benchmarksOrdered)

	for _, control := range m.Controls {
		if err := m.addItemIntoControlTree(control); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mod) addItemIntoControlTree(item ControlTreeItem) error {
	parents := m.getParents(item)

	// so we have a result - add into tree
	for _, p := range parents {
		// check this item does not exist in the parent path
		if helpers.StringSliceContains(p.Path(), item.Name()) {
			return fmt.Errorf("cyclical dependency adding '%s' into control tree - parent '%s'", item.Name(), p.Name())
		}
		item.AddParent(p)
		p.AddChild(item)
	}

	return nil
}

func (m *Mod) AddResource(item HclResource, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch r := item.(type) {
	case *Query:
		name := r.Name()
		// check for dupes
		if _, ok := m.Queries[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item, block))
			break
		}
		m.Queries[name] = r

	case *Control:
		name := r.Name()
		// check for dupes
		if _, ok := m.Controls[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item, block))
			break
		}
		m.Controls[name] = r
	case *Benchmark:
		name := r.Name()
		// check for dupes
		if _, ok := m.Benchmarks[name]; ok {
			diags = append(diags, duplicateResourceDiagnostics(item, block))
			break
		} else {
			m.Benchmarks[name] = r
		}

	}
	return diags

}

func duplicateResourceDiagnostics(item HclResource, block *hcl.Block) *hcl.Diagnostic {
	return &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("mod defines more that one resource named %s", item.Name()),
		Subject:  &block.DefRange,
	}
}

// AddChild  implements ControlTreeItem
func (m *Mod) AddChild(child ControlTreeItem) error {
	m.children = append(m.children, child)
	return nil
}

// AddParent implements ControlTreeItem
func (m *Mod) AddParent(ControlTreeItem) error {
	return errors.New("cannot set a parent on a mod")
}

// GetParents implements ControlTreeItem
func (m *Mod) GetParents() []ControlTreeItem {
	return nil
}

// Name implements ControlTreeItem, HclResource
func (m *Mod) Name() string {

	if m.Version == nil {
		return m.FullName
	}
	return fmt.Sprintf("%s@%s", m.FullName, types.SafeString(m.Version))
}

// GetTitle implements ControlTreeItem
func (m *Mod) GetTitle() string {
	return typehelpers.SafeString(m.Title)
}

// GetDescription implements ControlTreeItem
func (m *Mod) GetDescription() string {
	return typehelpers.SafeString(m.Description)
}

// GetTags implements ControlTreeItem
func (m *Mod) GetTags() map[string]string {
	if m.Tags != nil {
		return *m.Tags
	}
	return map[string]string{}
}

// GetChildren implements ControlTreeItem
func (m *Mod) GetChildren() []ControlTreeItem {
	return m.children
}

// Path implements ControlTreeItem
func (m *Mod) Path() []string {
	return []string{m.Name()}
}

// AddPseudoResource adds the pseudo resource to the mod,
// as long as there is no existing resource of same name
//
// A pseudo resource ids a resource created by loading a content file (e.g. a SQL file),
// rather than parsing a HCL defintion
func (m *Mod) AddPseudoResource(resource MappableResource) {
	switch r := resource.(type) {
	case *Query:
		// check there is not already a query with the same name
		if _, ok := m.Queries[r.Name()]; !ok {
			m.Queries[r.Name()] = r
			// set the mod on the query metadata
			r.GetMetadata().SetMod(m)
		}
	}
}

// CtyValue implements HclResource
func (m *Mod) CtyValue() (cty.Value, error) {
	return getCtyValue(m)
}

// GetMetadata implements HclResource
func (m *Mod) GetMetadata() *ResourceMetadata {
	return m.metadata
}

// OnDecoded implements HclResource
func (m *Mod) OnDecoded(*hcl.Block) {}

// AddReference implements HclResource
func (m *Mod) AddReference(reference string) {
	m.References = append(m.References, reference)
}

// SetMetadata implements ResourceWithMetadata
func (m *Mod) SetMetadata(metadata *ResourceMetadata) {
	m.metadata = metadata
}

// get the parent item for this ControlTreeItem
// first check all benchmarks - if they do not have this as child, default to the mod
func (m *Mod) getParents(item ControlTreeItem) []ControlTreeItem {
	var parents []ControlTreeItem
	for _, benchmark := range m.Benchmarks {
		if benchmark.ChildNames == nil {
			continue
		}
		// check all child names of this benchmark for a matching name
		for _, childName := range *benchmark.ChildNames {
			if childName.Name == item.Name() {
				parents = append(parents, benchmark)
			}
		}
	}
	if len(parents) == 0 {
		// fall back on mod
		parents = []ControlTreeItem{m}
	}
	return parents
}

// GetChildControls return a flat list of controls underneath the mod
func (m *Mod) GetChildControls() []*Control {
	var res []*Control
	for _, control := range m.Controls {
		res = append(res, control)
	}
	return res
}

//// BuildControTree :: populate the parent fields for all mods and benchmarslks
//func (m *Mod) BuildControTree() {
//
//	for name, benchmark := range m.Benchmarks{
//
//		for _, childName := range benchmark.ChildNameStrings{
//			parsedName, _ := ParseResourceName(childName)
//			if parsedName.ItemType ==BlockTypeControl{
//				child := m.Controls[childName]
//				child.A
//			}
//			child :=
//		}
//	},
//}
