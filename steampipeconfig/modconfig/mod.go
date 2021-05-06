package modconfig

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
)

// mod name used if a default mod is created for a workspace which does not define one explicitly
const defaultModName = "local"

type Mod struct {
	ShortName string `hcl:"name,label"`
	FullName  string `cty:"name"`

	// attributes
	Color         *string            `cty:"color" hcl:"color" column:"color,text"`
	Description   *string            `cty:"description" hcl:"description" column:"description,text"`
	Documentation *string            `cty:"documentation" hcl:"documentation" column:"documentation,text"`
	Icon          *string            `cty:"icon" hcl:"icon" column:"icon,text"`
	Labels        *[]string          `cty:"labels" hcl:"labels" column:"labels,jsonb"`
	Tags          *map[string]string `cty:"tags" hcl:"tags" column:"tags,jsonb"`
	Title         *string            `cty:"title" hcl:"title" column:"title,text"`

	// blocks
	Requires  *Requires  `hcl:"requires,block"`
	OpenGraph *OpenGraph `hcl:"opengraph,block"`

	// TODO do we need this?
	Version *string

	Queries    map[string]*Query
	Controls   map[string]*Control
	Benchmarks map[string]*Benchmark
	ModPath    string
	DeclRange  hcl.Range

	children []ControlTreeItem
	metadata *ResourceMetadata
}

// Schema :: implementation of HclResource
func (m *Mod) Schema() *hcl.BodySchema {
	// todo this could be done fully generically if we had a tag for block properties
	schema := &hcl.BodySchema{Attributes: []hcl.AttributeSchema{
		{Name: "color"},
		{Name: "description"},
		{Name: "documentation"},
		{Name: "icon"},
		{Name: "labels"},
		{Name: "title"},
	}}
	schema.Blocks = []hcl.BlockHeaderSchema{
		{Type: BlockTypeRequires},
		{Type: BlockTypeOpengraph},
	}
	return schema

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

// CreateDefaultMod :: create a default mod created for a workspace with no mod definition
func CreateDefaultMod(modPath string) *Mod {
	return NewMod(defaultModName, modPath, hcl.Range{})
}

// IsDefaultMod :: is this mod a default mod created for a workspace with no mod definition
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
//Mod Dependencies: %s
//Plugin Dependencies: %s
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
		//modDependStr,
		//pluginDependStr,
		strings.Join(queryStrings, "\n"),
		strings.Join(controlStrings, "\n"),
		strings.Join(benchmarkStrings, "\n"),
	)
}

// IsControlTreeItem :: implementation of ControlTreeItem
// (mod is always top of the tree)
func (m *Mod) IsControlTreeItem() {}

func (m *Mod) BuildControlTree() error {
	for _, benchmark := range m.Benchmarks {
		if err := m.addItemIntoControlTree(benchmark); err != nil {
			return err
		}
	}
	for _, control := range m.Controls {
		if err := m.addItemIntoControlTree(control); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mod) addItemIntoControlTree(item ControlTreeItem) error {
	parent := m.getParent(item)

	// check this item does not exist in the parent path
	if helpers.StringSliceContains(parent.Path(), item.Name()) {
		return fmt.Errorf("cyclical dependency adding '%s' into control tree - parent '%s'", item.Name(), parent.Name())
	}
	// so we have a result - add into tree
	item.SetParent(parent)
	parent.AddChild(item)

	return nil
}

func (m *Mod) AddResource(item HclResource) bool {
	switch r := item.(type) {
	case *Query:
		name := r.Name()
		// check for dupes
		if _, ok := m.Queries[name]; ok {
			return false
		}
		m.Queries[name] = r
		return true
	case *Control:
		name := r.Name()
		// check for dupes
		if _, ok := m.Controls[name]; ok {
			return false
		}
		m.Controls[name] = r
		return true
	case *Benchmark:
		name := r.Name()
		// check for dupes
		if _, ok := m.Benchmarks[name]; ok {
			return false
		}
		m.Benchmarks[name] = r
		return true
	default:
		// mod does not store other resource types
		return true
	}
}

// AddChild  :: implementation of ControlTreeItem
func (m *Mod) AddChild(child ControlTreeItem) error {
	m.children = append(m.children, child)
	return nil
}

// SetParent :: implementation of ControlTreeItem
func (m *Mod) SetParent(ControlTreeItem) error {
	return errors.New("cannot set a parent on a mod")
}

// Name :: implementation of ControlTreeItem, HclResource
func (m *Mod) Name() string {

	if m.Version == nil {
		return m.FullName
	}
	return fmt.Sprintf("%s@%s", m.FullName, types.SafeString(m.Version))
}

// Path :: implementation of ControlTreeItem
func (m *Mod) Path() []string {
	return []string{m.Name()}
}

// AddPseudoResource :: add resource to parse results, if there is no resource of same name
func (m *Mod) AddPseudoResource(resource MappableResource) {
	switch r := resource.(type) {
	case *Query:
		// check there is not already a query with the same name
		if _, ok := m.Queries[r.ShortName]; !ok {
			m.Queries[r.ShortName] = r
			// set the mod on the query metadata
			r.GetMetadata().SetMod(m)
		}
	}
}

// GetMetadata :: implementation of HclResource
func (m *Mod) GetMetadata() *ResourceMetadata {
	return m.metadata
}

// OnDecoded :: implementation of HclResource
func (m *Mod) OnDecoded() {}

// SetMetadata :: implementation of HclResource
func (m *Mod) SetMetadata(metadata *ResourceMetadata) {
	m.metadata = metadata
}

// get the parent item for this ControlTreeItem
// first check all benchmarks - if they do not have this as child, default to the mod
func (m *Mod) getParent(item ControlTreeItem) ControlTreeItem {
	for _, benchmark := range m.Benchmarks {
		if benchmark.ChildNames == nil {
			continue
		}
		// check all child names of this benchmark for a matching name
		for _, childName := range *benchmark.ChildNames {
			if childName.Name == item.Name() {
				return benchmark
			}
		}
	}
	// fall back on mod
	return m
}
