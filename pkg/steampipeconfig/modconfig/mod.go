package modconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	filehelpers "github.com/turbot/go-kit/files"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/zclconf/go-cty/cty"
)

// mod name used if a default mod is created for a workspace which does not define one explicitly
const defaultModName = "local"

// Mod is a struct representing a Mod resource
type Mod struct {
	ResourceWithMetadataImpl
	ModTreeItemImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	// ModDependencyPath is the fully qualified mod name, which can be used to 'require'  the mod,
	// e.g. github.com/turbot/steampipe-mod-azure-thrifty
	// This is only set if the mod is installed as a dependency
	ModDependencyPath string `cty:"mod_dependency_path"`

	// attributes
	Categories []string `cty:"categories" hcl:"categories,optional" column:"categories,jsonb"`
	Color      *string  `cty:"color" hcl:"color" column:"color,text"`
	Icon       *string  `cty:"icon" hcl:"icon" column:"icon,text"`

	// blocks
	Require       *Require
	LegacyRequire *Require   `hcl:"requires,block"`
	OpenGraph     *OpenGraph `hcl:"opengraph,block" column:"open_graph,jsonb"`

	VersionString string `cty:"version"`
	Version       *semver.Version

	// ModPath is the installation location of the mod
	ModPath string

	// the filepath of the mod.sp file (will be empty for default mod)
	modFilePath string
	// convenient aggregation of all resources
	// NOTE: this resource map object references the same set of resources
	ResourceMaps *ResourceMaps
}

func NewMod(shortName, modPath string, defRange hcl.Range) *Mod {
	require := NewRequire()
	name := fmt.Sprintf("mod.%s", shortName)
	mod := &Mod{
		ModTreeItemImpl: ModTreeItemImpl{
			HclResourceImpl: HclResourceImpl{
				ShortName:       shortName,
				FullName:        name,
				UnqualifiedName: name,
				DeclRange:       defRange,
				blockType:       BlockTypeMod,
			},
		},
		ModPath: modPath,
		Require: require,
	}
	mod.ResourceMaps = NewModResources(mod)

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

	// now check the child resources
	return m.ResourceMaps.Equals(other.ResourceMaps)
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
	return m.modFilePath == ""
}

func (m *Mod) NameWithVersion() string {
	if m.VersionString == "" {
		return m.ShortName
	}
	return fmt.Sprintf("%s@%s", m.ShortName, m.VersionString)
}

// GetModDependencyPath ModDependencyPath if it is set. If not it returns NameWithVersion()
func (m *Mod) GetModDependencyPath() string {
	if m.ModDependencyPath != "" {
		return m.ModDependencyPath
	}
	return m.NameWithVersion()
}

// GetPaths implements ModTreeItem (override base functionality)
func (m *Mod) GetPaths() []NodePath {
	return []NodePath{{m.Name()}}
}

// SetPaths implements ModTreeItem (override base functionality)
func (m *Mod) SetPaths() {}

// OnDecoded implements HclResource
func (m *Mod) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	// if VersionString is set, set Version
	if m.VersionString != "" && m.Version == nil {
		m.Version, _ = semver.NewVersion(m.VersionString)
	}

	// handle legacy requires block
	if m.LegacyRequire != nil && !m.LegacyRequire.Empty() {
		// ensure that both 'require' and 'requires' were not set
		for _, b := range block.Body.(*hclsyntax.Body).Blocks {
			if b.Type == BlockTypeRequire {
				subject := BlockRange(block)
				return hcl.Diagnostics{&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Both 'require' and legacy 'requires' blocks are defined",
					Subject:  &subject,
				}}
			}
		}
		m.Require = m.LegacyRequire
	}

	// initialise our Require
	if m.Require == nil {
		return nil
	}
	err := m.Require.initialise()
	if err != nil {
		subject := BlockRange(block)
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
			Subject:  &subject,
		}}
	}
	return nil

}

// AddReference implements ResourceWithMetadata (overridden from ResourceWithMetadataImpl)
func (m *Mod) AddReference(ref *ResourceReference) {
	m.ResourceMaps.References[ref.Name()] = ref
}

// GetReferences implements ResourceWithMetadata (overridden from ResourceWithMetadataImpl)
func (m *Mod) GetReferences() []*ResourceReference {
	var res = make([]*ResourceReference, len(m.ResourceMaps.References))
	// convert from map to array
	idx := 0
	for _, ref := range m.ResourceMaps.References {
		res[idx] = ref
		idx++
	}
	return res
}

// GetResourceMaps implements ResourceMapsProvider
func (m *Mod) GetResourceMaps() *ResourceMaps {
	return m.ResourceMaps
}

func (m *Mod) GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool) {
	return m.ResourceMaps.GetResource(parsedName)
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
	if !filehelpers.FileExists(modFilePath) {
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

func (m *Mod) WalkResources(resourceFunc func(item HclResource) (bool, error)) error {
	return m.ResourceMaps.WalkResources(resourceFunc)
}

func (m *Mod) SetFilePath(modFilePath string) {
	m.modFilePath = modFilePath
}

func (m *Mod) ValidateSteampipeVersion() error {
	if m.Require == nil {
		return nil
	}
	return m.Require.ValidateSteampipeVersion(m.Name())
}

// CtyValue implements CtyValueProvider
func (m *Mod) CtyValue() (cty.Value, error) {
	return GetCtyValue(m)
}
