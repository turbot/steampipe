package modconfig

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/version"
	"github.com/turbot/steampipe/sperr"
)

// Require is a struct representing mod dependencies
type Require struct {
	Plugins                          []*PluginVersion        `hcl:"plugin,block"`
	DeprecatedSteampipeVersionString string                  `hcl:"steampipe,optional"`
	Steampipe                        *SteampipeRequire       `hcl:"steampipe,block"`
	Mods                             []*ModVersionConstraint `hcl:"mod,block"`
	// map keyed by name [and alias]
	modMap map[string]*ModVersionConstraint
	// range of the definition of the require block
	DeclRange hcl.Range
	// range of the body of the require block
	BodyRange hcl.Range
}

func NewRequire() *Require {
	return &Require{
		modMap: make(map[string]*ModVersionConstraint),
	}
}
func (r *Require) Clone() *Require {
	require := NewRequire()
	require.Steampipe = r.Steampipe
	require.Plugins = r.Plugins
	require.Mods = r.Mods
	require.DeclRange = r.DeclRange
	require.BodyRange = r.BodyRange

	// we need to shallow copy the map
	// if we don't, when the other one gets
	// modified - this one gets as well
	for k, mvc := range r.modMap {
		require.modMap[k] = mvc
	}
	return require
}

func (r *Require) initialise(modBlock *hcl.Block) hcl.Diagnostics {
	// This will actually be called twice - once when we load the mod definition,
	// and again when we load the mod resources (and set the mod metadata, references etc)
	// If we have already initialised, return (we can tell by checking the DeclRange)
	if !r.DeclRange.Empty() {
		return nil
	}

	var diags hcl.Diagnostics
	// handle deprecated properties
	moreDiags := r.handleDeprecations()
	diags = append(diags, moreDiags...)

	// find the require block
	requireBlock := hclhelpers.FindFirstChildBlock(modBlock, BlockTypeRequire)
	if requireBlock == nil {
		// was this the legacy 'requires' block?
		requireBlock = hclhelpers.FindFirstChildBlock(modBlock, BlockTypeLegacyRequires)
	}
	if requireBlock == nil {
		// nothing else to populate
		return nil
	}

	// set our Ranges
	r.DeclRange = requireBlock.DefRange
	r.BodyRange = requireBlock.Body.(*hclsyntax.Body).SrcRange

	// build maps of plugin and mod blocks
	pluginBlockMap := hclhelpers.BlocksToMap(hclhelpers.FindChildBlocks(requireBlock, BlockTypePlugin))
	modBlockMap := hclhelpers.BlocksToMap(hclhelpers.FindChildBlocks(requireBlock, BlockTypeMod))

	if r.Steampipe != nil {
		moreDiags := r.Steampipe.initialise(requireBlock)
		diags = append(diags, moreDiags...)
	}

	for _, p := range r.Plugins {
		moreDiags := p.Initialise(pluginBlockMap[p.RawName])
		diags = append(diags, moreDiags...)
	}
	for _, m := range r.Mods {
		moreDiags := m.Initialise(modBlockMap[m.Name])
		diags = append(diags, moreDiags...)
		if !diags.HasErrors() {
			// key map entry by name [and alias]
			r.modMap[m.Name] = m
		}
	}

	return diags
}

func (r *Require) handleDeprecations() hcl.Diagnostics {
	var diags hcl.Diagnostics
	// the 'steampipe' property is deprecated and replace with a steampipe block
	if r.DeprecatedSteampipeVersionString != "" {
		// if there is both a steampipe block and property, fail
		if r.Steampipe != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Both 'steampipe' block and deprecated 'steampipe' property are set",
				Subject:  &r.DeclRange,
			})
		} else {
			r.Steampipe = &SteampipeRequire{MinVersionString: r.DeprecatedSteampipeVersionString}
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  "Property 'steampipe' is deprecated for mod require block - use a steampipe block instead",
				Subject:  &r.DeclRange,
			},
			)
		}
	}
	return diags
}

func (r *Require) validateSteampipeVersion(modName string) error {
	if steampipeVersionConstraint := r.SteampipeVersionConstraint(); steampipeVersionConstraint != nil {
		if !steampipeVersionConstraint.Check(version.SteampipeVersion) {
			return fmt.Errorf("steampipe version %s does not satisfy %s which requires version %s", version.SteampipeVersion.String(), modName, r.Steampipe.MinVersionString)
		}
	}
	return nil
}

// validatePluginVersions validates that for every plugin requirement there's at least one plugin installed
func (r *Require) validatePluginVersions(modName string, plugins map[string]*utils.PluginVersion) []error {
	if len(r.Plugins) == 0 {
		return nil
	}
	validationErrors := []error{}
	for _, requiredPlugin := range r.Plugins {
		if err := r.searchInstalledPluginForRequirement(modName, requiredPlugin, plugins); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}
	return validationErrors
}

func (r *Require) searchInstalledPluginForRequirement(modName string, requirement *PluginVersion, plugins map[string]*utils.PluginVersion) error {
	for installedName, installed := range plugins {
		org, name, _ := ociinstaller.NewSteampipeImageRef(installedName).GetOrgNameAndStream()
		if org != requirement.Org || name != requirement.Name {
			// no point check - different plugin
			continue
		}
		// if org and name matches but the plugin is built locally, return
		if org == requirement.Org && name == requirement.Name && installed.Version == "local" {
			return nil
		}
		if requirement.Constraint.Check(installed.Semver()) {
			// constraint is satisfied
			return nil
		}
	}
	return sperr.New("could not find plugin which satisfies requirement '%s@%s' - required by '%s'", requirement.RawName, requirement.MinVersionString, modName)
}

// AddModDependencies adds all the mod in newModVersions to our list of mods, using the following logic
// - if a mod with same name, [alias] and constraint exists, it is not added
// - if a mod with same name [and alias] and different constraint exist, it is replaced
func (r *Require) AddModDependencies(newModVersions map[string]*ModVersionConstraint) {
	// rebuild the Mods array

	// first rebuild the mod map
	for name, newVersion := range newModVersions {
		r.modMap[name] = newVersion
	}

	// now update the mod array from the map
	var newMods = make([]*ModVersionConstraint, len(r.modMap))
	idx := 0
	for _, requiredVersion := range r.modMap {
		newMods[idx] = requiredVersion
		idx++
	}
	// sort by name
	sort.Sort(ModVersionConstraintCollection(newMods))
	// write back
	r.Mods = newMods
}

func (r *Require) RemoveModDependencies(versions map[string]*ModVersionConstraint) {
	// first rebuild the mod map
	for name := range versions {
		delete(r.modMap, name)
	}
	// now update the mod array from the map
	var newMods = make([]*ModVersionConstraint, len(r.modMap))
	idx := 0
	for _, requiredVersion := range r.modMap {
		newMods[idx] = requiredVersion
		idx++
	}
	// sort by name
	sort.Sort(ModVersionConstraintCollection(newMods))
	// write back
	r.Mods = newMods
}

func (r *Require) RemoveAllModDependencies() {
	r.Mods = nil
}

func (r *Require) GetModDependency(name string /*,alias string*/) *ModVersionConstraint {
	return r.modMap[name]
}

func (r *Require) ContainsMod(requiredModVersion *ModVersionConstraint) bool {
	if c := r.GetModDependency(requiredModVersion.Name); c != nil {
		return c.Equals(requiredModVersion)
	}
	return false
}

func (r *Require) Empty() bool {
	return r.SteampipeVersionConstraint() == nil && len(r.Mods) == 0 && len(r.Plugins) == 0
}

func (r *Require) SteampipeVersionConstraint() *semver.Constraints {
	if r.Steampipe == nil {
		return nil
	}
	return r.Steampipe.Constraint

}
