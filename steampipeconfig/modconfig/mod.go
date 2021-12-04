package modconfig

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// mod name used if a default mod is created for a workspace which does not define one explicitly
const defaultModName = "local"

// Mod is a struct representing a Mod resource
type Mod struct {
	// ShortName is the mod name, e.g. azure_thrifty
	ShortName string `cty:"short_name" hcl:"name,label"`
	// FullName is the mod name prefixed with 'mod', e.g. mod.azure_thrifty
	FullName string `cty:"name"`
	// ModDependencyPath is the fully qualified mod name, which can be used to 'require'  the mod,
	// e.g. github.com/turbot/steampipe-mod-azure-thrifty
	// This is only set if the mod is installed as a dependency
	ModDependencyPath string `cty:"mod_dependency_path"`

	// attributes
	Categories    *[]string          `cty:"categories" hcl:"categories" column:"categories,jsonb"`
	Color         *string            `cty:"color" hcl:"color" column:"color,text"`
	Description   *string            `cty:"description" hcl:"description" column:"description,text"`
	Documentation *string            `cty:"documentation" hcl:"documentation" column:"documentation,text"`
	Icon          *string            `cty:"icon" hcl:"icon" column:"icon,text"`
	Tags          *map[string]string `cty:"tags" hcl:"tags" column:"tags,jsonb"`
	Title         *string            `cty:"title" hcl:"title" column:"title,text"`

	// list of all blocks referenced by the resource
	References []*ResourceReference

	// blocks
	Requires  *Requires  `hcl:"requires,block"`
	OpenGraph *OpenGraph `hcl:"opengraph,block" column:"open_graph,jsonb"`

	VersionString string `cty:"version"`
	Version       *semver.Version

	Queries    map[string]*Query
	Controls   map[string]*Control
	Benchmarks map[string]*Benchmark
	Reports    map[string]*Report
	Panels     map[string]*Panel
	Variables  map[string]*Variable
	Locals     map[string]*Local

	// flat list of all resources
	AllResources map[string]HclResource

	// list of benchmark names, sorted alphabetically
	benchmarksOrdered []string

	// ModPath is the installation location of the mod
	ModPath   string
	DeclRange hcl.Range

	children []ModTreeItem
	metadata *ResourceMetadata
}

func NewMod(shortName, modPath string, defRange hcl.Range) *Mod {
	mod := &Mod{
		ShortName:    shortName,
		FullName:     fmt.Sprintf("mod.%s", shortName),
		Queries:      make(map[string]*Query),
		Controls:     make(map[string]*Control),
		Benchmarks:   make(map[string]*Benchmark),
		Reports:      make(map[string]*Report),
		Panels:       make(map[string]*Panel),
		Variables:    make(map[string]*Variable),
		Locals:       make(map[string]*Local),
		ModPath:      modPath,
		DeclRange:    defRange,
		AllResources: make(map[string]HclResource),
		Requires:     new(Requires),
	}
	// try to derive mod version from the path
	mod.setVersion()
	return mod
}

func (m *Mod) setVersion() {
	segments := strings.Split(m.ModPath, "@")
	if len(segments) == 1 {
		return
	}
	versionString := segments[len(segments)-1]
	// try to set version, ignoring error
	version, err := semver.NewVersion(versionString)
	if err == nil {
		m.Version = version
		m.VersionString = fmt.Sprintf("%d.%d", version.Major(), version.Minor())
	}
}

func (m *Mod) Equals(other *Mod) bool {
	res := m.ShortName == other.ShortName &&
		m.FullName == other.FullName &&
		typehelpers.SafeString(m.Color) == typehelpers.SafeString(other.Color) &&
		typehelpers.SafeString(m.Description) == typehelpers.SafeString(other.Description) &&
		typehelpers.SafeString(m.Documentation) == typehelpers.SafeString(other.Documentation) &&
		typehelpers.SafeString(m.Icon) == typehelpers.SafeString(other.Icon) &&
		typehelpers.SafeString(m.Title) == typehelpers.SafeString(other.Title)
	if !res {
		return res
	}
	// categories
	if m.Categories == nil {
		if other.Categories != nil {
			return false
		}
	} else {
		// we have categories
		if other.Categories == nil {
			return false
		}

		if len(*m.Categories) != len(*other.Categories) {
			return false
		}
		for i, c := range *m.Categories {
			if (*other.Categories)[i] != c {
				return false
			}
		}
	}
	// tags
	if m.Tags == nil {
		if other.Tags != nil {
			return false
		}
	} else {
		// we have tags
		if other.Tags == nil {
			return false
		}
		for k, v := range *m.Tags {
			if otherVal, ok := (*other.Tags)[k]; !ok && v != otherVal {
				return false
			}
		}
	}

	// controls
	for k := range m.Controls {
		if _, ok := other.Controls[k]; !ok {
			return false
		}
	}
	for k := range m.Queries {
		if _, ok := other.Queries[k]; !ok {
			return false
		}
	}
	for k := range other.Queries {
		if _, ok := m.Queries[k]; !ok {
			return false
		}
	}
	// queries
	for k := range other.Controls {
		if _, ok := m.Controls[k]; !ok {
			return false
		}
	}
	// benchmarks
	for k := range m.Benchmarks {
		if _, ok := other.Benchmarks[k]; !ok {
			return false
		}
	}
	for k := range other.Benchmarks {
		if _, ok := m.Benchmarks[k]; !ok {
			return false
		}
	}
	// reports
	for k := range m.Reports {
		if _, ok := other.Reports[k]; !ok {
			return false
		}
	}
	for k := range other.Reports {
		if _, ok := m.Reports[k]; !ok {
			return false
		}
	}
	// panels
	for k := range m.Panels {
		if _, ok := other.Panels[k]; !ok {
			return false
		}
	}
	for k := range other.Panels {
		if _, ok := m.Panels[k]; !ok {
			return false
		}
	}
	// variables
	for k := range m.Variables {
		if _, ok := other.Variables[k]; !ok {
			return false
		}
	}
	for k := range other.Variables {
		if _, ok := m.Variables[k]; !ok {
			return false
		}
	}
	// locals
	for k := range m.Locals {
		if _, ok := other.Locals[k]; !ok {
			return false
		}
	}
	for k := range other.Locals {
		if _, ok := m.Locals[k]; !ok {
			return false
		}
	}
	return true

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
	if m.VersionString == "" {
		versionString = fmt.Sprintf("\nVersion: v%s", m.VersionString)
	}
	var requiresStrings []string
	var requiresString string
	if m.Requires != nil {
		if m.Requires.SteampipeVersionString != "" {
			requiresStrings = append(requiresStrings, fmt.Sprintf("Steampipe %s", m.Requires.SteampipeVersionString))
		}
		for _, m := range m.Requires.Mods {
			requiresStrings = append(requiresStrings, m.String())
		}
		for _, p := range m.Requires.Plugins {
			requiresStrings = append(requiresStrings, p.String())
		}
		requiresString = fmt.Sprintf("Requires: \n%s", strings.Join(requiresStrings, "\n"))
	}

	return fmt.Sprintf(`Name: %s
Title: %s
Description: %s 
Version: %s
Queries: 
%s
Controls: 
%s
Benchmarks: 
%s
%s`,
		m.FullName,
		types.SafeString(m.Title),
		types.SafeString(m.Description),
		versionString,
		strings.Join(queryStrings, "\n"),
		strings.Join(controlStrings, "\n"),
		strings.Join(benchmarkStrings, "\n"),
		requiresString,
	)
}

// BuildResourceTree builds the control tree structure by setting the parent property for each control and benchmar
// NOTE: this also builds the sorted benchmark list
func (m *Mod) BuildResourceTree() error {
	// build sorted list of benchmarks
	m.benchmarksOrdered = make([]string, len(m.Benchmarks))
	idx := 0
	for name, benchmark := range m.Benchmarks {
		// save this benchmark name
		m.benchmarksOrdered[idx] = name
		idx++

		// add benchmark into control tree
		if err := m.addItemIntoResourceTree(benchmark); err != nil {
			return err
		}
	}
	// now sort the benchmark names
	sort.Strings(m.benchmarksOrdered)

	for _, control := range m.Controls {
		if err := m.addItemIntoResourceTree(control); err != nil {
			return err
		}
	}
	for _, panel := range m.Panels {
		if err := m.addItemIntoResourceTree(panel); err != nil {
			return err
		}
	}
	for _, report := range m.Reports {
		if err := m.addItemIntoResourceTree(report); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mod) addItemIntoResourceTree(item ModTreeItem) error {
	parents := m.getParents(item)

	// so we have a result - add into tree
	for _, p := range parents {
		// TODO validity checking
		//for _, parentPath := range p.GetPaths() {
		//	// check this item does not exist in the parent path
		//	if helpers.StringSliceContains(parentPath, item.Name()) {
		//		return fmt.Errorf("cyclical dependency adding '%s' into control tree - parent '%s'", item.Name(), p.Name())
		//	}
		item.AddParent(p)
		p.AddChild(item)
		//}
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
		name := r.Name()
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
	m.AllResources[item.Name()] = item
	return diags
}

func duplicateResourceDiagnostics(item HclResource) *hcl.Diagnostic {
	return &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("mod defines more than one resource named %s", item.Name()),
		Subject:  item.GetDeclRange(),
	}
}

func (m *Mod) NameWithVersion() string {
	if m.VersionString == "" {
		return m.ShortName
	}
	return fmt.Sprintf("%s@%s", m.ShortName, m.VersionString)
}

// AddChild  implements ModTreeItem
func (m *Mod) AddChild(child ModTreeItem) error {
	m.children = append(m.children, child)
	return nil
}

// AddParent implements ModTreeItem
func (m *Mod) AddParent(ModTreeItem) error {
	return errors.New("cannot set a parent on a mod")
}

// GetParents implements ModTreeItem
func (m *Mod) GetParents() []ModTreeItem {
	return nil
}

// Name implements ModTreeItem, HclResource
func (m *Mod) Name() string {
	return m.FullName
}

// GetTitle implements ModTreeItem
func (m *Mod) GetTitle() string {
	return typehelpers.SafeString(m.Title)
}

// GetDescription implements ModTreeItem
func (m *Mod) GetDescription() string {
	return typehelpers.SafeString(m.Description)
}

// GetTags implements ModTreeItem
func (m *Mod) GetTags() map[string]string {
	if m.Tags != nil {
		return *m.Tags
	}
	return map[string]string{}
}

// GetChildren implements ModTreeItem
func (m *Mod) GetChildren() []ModTreeItem {
	return m.children
}

// GetPaths implements ModTreeItem
func (m *Mod) GetPaths() []NodePath {
	return []NodePath{{m.Name()}}
}

// AddPseudoResource adds the pseudo resource to the mod,
// as long as there is no existing resource of same name
//
// A pseudo resource ids a resource created by loading a content file (e.g. a SQL file),
// rather than parsing a HCL definition
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

// OnDecoded implements HclResource
func (m *Mod) OnDecoded(*hcl.Block) hcl.Diagnostics {
	// if VersionString is set, set Version
	if m.VersionString != "" && m.Version == nil {
		m.Version, _ = semver.NewVersion(m.VersionString)
	}
	// initialise our Requires
	if m.Requires == nil {
		return nil
	}
	return m.Requires.Initialise()
}

// AddReference implements HclResource
func (m *Mod) AddReference(ref *ResourceReference) {
	m.References = append(m.References, ref)
}

// SetMod implements HclResource
func (m *Mod) SetMod(*Mod) {}

// GetMod implements HclResource
func (m *Mod) GetMod() *Mod {
	return nil
}

// GetDeclRange implements HclResource
func (m *Mod) GetDeclRange() *hcl.Range {
	return &m.DeclRange
}

// GetMetadata implements ResourceWithMetadata
func (m *Mod) GetMetadata() *ResourceMetadata {
	return m.metadata
}

// SetMetadata implements ResourceWithMetadata
func (m *Mod) SetMetadata(metadata *ResourceMetadata) {
	m.metadata = metadata
}

// get the parent item for this ModTreeItem
// first check all benchmarks - if they do not have this as child, default to the mod
func (m *Mod) getParents(item ModTreeItem) []ModTreeItem {
	var parents []ModTreeItem
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
	for _, report := range m.Reports {
		// check all child names of this benchmark for a matching name
		for _, child := range report.GetChildren() {
			if child.Name() == item.Name() {
				parents = append(parents, report)
			}
		}
	}
	for _, panel := range m.Panels {
		// check all child names of this benchmark for a matching name
		for _, child := range panel.GetChildren() {
			if child.Name() == item.Name() {
				parents = append(parents, panel)
			}
		}
	}
	if len(parents) == 0 {
		// fall back on mod
		parents = []ModTreeItem{m}
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
