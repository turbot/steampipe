package modconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/filepaths"
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
	Categories    []string          `cty:"categories" hcl:"categories,optional" column:"categories,jsonb"`
	Color         *string           `cty:"color" hcl:"color" column:"color,text"`
	Description   *string           `cty:"description" hcl:"description" column:"description,text"`
	Documentation *string           `cty:"documentation" hcl:"documentation" column:"documentation,text"`
	Icon          *string           `cty:"icon" hcl:"icon" column:"icon,text"`
	Tags          map[string]string `cty:"tags" hcl:"tags,optional" column:"tags,jsonb"`
	Title         *string           `cty:"title" hcl:"title" column:"title,text"`

	// list of all blocks referenced by the resource
	References []*ResourceReference

	// blocks
	Require       *Require   `hcl:"require,block"`
	LegacyRequire *Require   `hcl:"requires,block"`
	OpenGraph     *OpenGraph `hcl:"opengraph,block" column:"open_graph,jsonb"`

	VersionString string `cty:"version"`
	Version       *semver.Version

	Queries    map[string]*Query
	Controls   map[string]*Control
	Benchmarks map[string]*Benchmark
	Reports    map[string]*ReportContainer
	Containers map[string]*ReportContainer
	Panels     map[string]*Panel
	Variables  map[string]*Variable
	Locals     map[string]*Local

	// ModPath is the installation location of the mod
	ModPath   string
	DeclRange hcl.Range

	// array of direct mod children - excludes resources which are children of other resources
	children []ModTreeItem
	metadata *ResourceMetadata
}

func NewMod(shortName, modPath string, defRange hcl.Range) *Mod {
	mod := &Mod{
		ShortName:  shortName,
		FullName:   fmt.Sprintf("mod.%s", shortName),
		Queries:    make(map[string]*Query),
		Controls:   make(map[string]*Control),
		Benchmarks: make(map[string]*Benchmark),
		Reports:    make(map[string]*ReportContainer),
		Containers: make(map[string]*ReportContainer),
		Panels:     make(map[string]*Panel),
		Variables:  make(map[string]*Variable),
		Locals:     make(map[string]*Local),

		ModPath:   modPath,
		DeclRange: defRange,
		Require:   newRequire(),
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

		if len(m.Categories) != len(other.Categories) {
			return false
		}
		for i, c := range m.Categories {
			if (other.Categories)[i] != c {
				return false
			}
		}
	}
	// tags
	if len(m.Tags) != len(other.Tags) {
		return false
	}
	for k, v := range m.Tags {
		if otherVal := other.Tags[k]; v != otherVal {
			return false
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
	// containers
	for k := range m.Containers {
		if _, ok := other.Containers[k]; !ok {
			return false
		}
	}
	for k := range other.Containers {
		if _, ok := m.Containers[k]; !ok {
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
	if m.Require != nil {
		if m.Require.SteampipeVersionString != "" {
			requiresStrings = append(requiresStrings, fmt.Sprintf("Steampipe %s", m.Require.SteampipeVersionString))
		}
		for _, m := range m.Require.Mods {
			requiresStrings = append(requiresStrings, m.String())
		}
		for _, p := range m.Require.Plugins {
			requiresStrings = append(requiresStrings, p.String())
		}
		requiresString = fmt.Sprintf("Require: \n%s", strings.Join(requiresStrings, "\n"))
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

func (m *Mod) NameWithVersion() string {
	if m.VersionString == "" {
		return m.ShortName
	}
	return fmt.Sprintf("%s@%s", m.ShortName, m.VersionString)
}

// AddChild implements ModTreeItem
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

// GetModDependencyPath ModDependencyPath if it is set. If not it returns NameWithVersion()
func (m *Mod) GetModDependencyPath() string {
	if m.ModDependencyPath != "" {
		return m.ModDependencyPath
	}
	return m.NameWithVersion()
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
		return m.Tags
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

// SetPaths implements ModTreeItem
func (m *Mod) SetPaths() {}

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
func (m *Mod) OnDecoded(block *hcl.Block) hcl.Diagnostics {
	// if VersionString is set, set Version
	if m.VersionString != "" && m.Version == nil {
		m.Version, _ = semver.NewVersion(m.VersionString)
	}

	// handle legacy requires block
	if m.LegacyRequire != nil && !m.Require.Empty() {
		if m.Require != nil && !m.Require.Empty() {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Both 'require' and legacy 'requires' blocks are defined",
				Subject:  &block.DefRange,
			}}
		}
		m.Require = m.LegacyRequire
	}

	// initialise our Require
	if m.Require == nil {
		return nil
	}
	return m.Require.initialise()
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
func (m *Mod) getParents(item ModTreeItem) []ModTreeItem {
	var parents []ModTreeItem

	for _, benchmark := range m.Benchmarks {
		if benchmark.ChildNames == nil {
			continue
		}
		// check all child names of this benchmark for a matching name
		for _, childName := range benchmark.ChildNames {
			if childName.Name == item.Name() {
				parents = append(parents, benchmark)
			}
		}
	}
	for _, report := range m.Reports {
		// check all child names of this report for a matching name
		for _, child := range report.GetChildren() {
			if child.Name() == item.Name() {
				parents = append(parents, report)
			}
		}
	}
	for _, container := range m.Containers {
		// check all child names of this container for a matching name
		for _, child := range container.GetChildren() {
			if child.Name() == item.Name() {
				parents = append(parents, container)
			}
		}
	}
	// if this item has no parents and is a child of the mod, set the mod as parent
	if len(parents) == 0 && m.containsResource(item.Name()) {
		parents = []ModTreeItem{m}

	}
	return parents
}

// does the mod contain a resource with this name?
func (m *Mod) containsResource(childName string) bool {
	if _, ok := m.Queries[childName]; ok {
		return true
	}
	if _, ok := m.Controls[childName]; ok {
		return true
	}
	if _, ok := m.Benchmarks[childName]; ok {
		return true
	}
	if _, ok := m.Reports[childName]; ok {
		return true
	}
	if _, ok := m.Panels[childName]; ok {
		return true
	}
	if _, ok := m.Containers[childName]; ok {
		return true
	}
	if _, ok := m.Variables[childName]; ok {
		return true
	}
	if _, ok := m.Locals[childName]; ok {
		return true
	}

	return false
}

// GetChildControls return a flat list of controls underneath the mod
func (m *Mod) GetChildControls() []*Control {
	var res []*Control
	for _, control := range m.Controls {
		res = append(res, control)
	}
	return res
}

func (m *Mod) AddModDependencies(modVersions map[string]*ModVersionConstraint) {
	m.Require.AddModDependencies(modVersions)
}

func (m *Mod) RemoveModDependencies(modVersions map[string]*ModVersionConstraint) {
	m.Require.RemoveModDependencies(modVersions)
}

func (m *Mod) RemoveAllModDependencies() {
	m.Require.RemoveAllModDependencies()
}

func (m *Mod) Save() error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	modBody := rootBody.AppendNewBlock("mod", []string{m.ShortName}).Body()
	if m.Title != nil {
		modBody.SetAttributeValue("title", cty.StringVal(*m.Title))
	}
	if m.Description != nil {
		modBody.SetAttributeValue("description", cty.StringVal(*m.Description))
	}
	if m.Color != nil {
		modBody.SetAttributeValue("color", cty.StringVal(*m.Color))
	}
	if m.Documentation != nil {
		modBody.SetAttributeValue("documentation", cty.StringVal(*m.Documentation))
	}
	if m.Icon != nil {
		modBody.SetAttributeValue("icon", cty.StringVal(*m.Icon))
	}
	if len(m.Categories) > 0 {
		categoryValues := make([]cty.Value, len(m.Categories))
		for i, c := range m.Categories {
			categoryValues[i] = cty.StringVal(typehelpers.SafeString(c))
		}
		modBody.SetAttributeValue("categories", cty.ListVal(categoryValues))
	}

	if len(m.Tags) > 0 {
		tagMap := make(map[string]cty.Value, len(m.Tags))
		for k, v := range m.Tags {
			tagMap[k] = cty.StringVal(v)
		}
		modBody.SetAttributeValue("tags", cty.MapVal(tagMap))
	}

	// opengraph
	if opengraph := m.OpenGraph; opengraph != nil {
		opengraphBody := modBody.AppendNewBlock("opengraph", nil).Body()
		if opengraph.Title != nil {
			opengraphBody.SetAttributeValue("title", cty.StringVal(*opengraph.Title))
		}
		if opengraph.Description != nil {
			opengraphBody.SetAttributeValue("description", cty.StringVal(*opengraph.Description))
		}
		if opengraph.Image != nil {
			opengraphBody.SetAttributeValue("image", cty.StringVal(*opengraph.Image))
		}

	}

	// require
	if require := m.Require; require != nil && !m.Require.Empty() {
		requiresBody := modBody.AppendNewBlock("require", nil).Body()
		if require.SteampipeVersionString != "" {
			requiresBody.SetAttributeValue("steampipe", cty.StringVal(require.SteampipeVersionString))
		}
		if len(require.Plugins) > 0 {
			pluginValues := make([]cty.Value, len(require.Plugins))
			for i, p := range require.Plugins {
				pluginValues[i] = cty.StringVal(typehelpers.SafeString(p))
			}
			requiresBody.SetAttributeValue("plugins", cty.ListVal(pluginValues))
		}
		if len(require.Mods) > 0 {
			for _, m := range require.Mods {
				modBody := requiresBody.AppendNewBlock("mod", []string{m.Name}).Body()
				modBody.SetAttributeValue("version", cty.StringVal(m.VersionString))
			}
		}
	}

	// load existing mod data and remove the mod definitions from it
	nonModData, err := m.loadNonModDataInModFile()
	if err != nil {
		return err
	}
	modData := append(f.Bytes(), nonModData...)
	return os.WriteFile(filepaths.ModFilePath(m.ModPath), modData, 0644)
}

func (m *Mod) HasDependentMods() bool {
	return m.Require != nil && len(m.Require.Mods) > 0
}

func (m *Mod) GetModDependency(modName string) *ModVersionConstraint {
	if m.Require == nil {
		return nil
	}
	return m.Require.GetModDependency(modName)
}

func (m *Mod) loadNonModDataInModFile() ([]byte, error) {
	modFilePath := filepaths.ModFilePath(m.ModPath)
	if !helpers.FileExists(modFilePath) {
		return nil, nil
	}

	fileData, err := os.ReadFile(modFilePath)
	if err != nil {
		return nil, err
	}

	fileLines := strings.Split(string(fileData), "\n")
	decl := m.DeclRange
	// just use line positions
	start := decl.Start.Line - 1
	end := decl.End.Line - 1

	var resLines []string
	for i, line := range fileLines {
		if (i < start || i > end) && line != "" {
			resLines = append(resLines, line)
		}
	}
	return []byte(strings.Join(resLines, "\n")), nil
}
