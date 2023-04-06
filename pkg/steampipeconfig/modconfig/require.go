package modconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
	"github.com/turbot/steampipe/pkg/version"
	"github.com/turbot/steampipe/sperr"
)

// Require is a struct representing mod dependencies
type Require struct {
	SteampipeVersion       *semver.Version
	SteampipeVersionString string           `hcl:"steampipe,optional"`
	Plugins                []*PluginVersion `hcl:"plugin,block"`
	Mods                   []*ModVersionConstraint
	DeclRange              hcl.Range
	// map keyed by name [and alias]
	modMap map[string]*ModVersionConstraint
}

func NewRequire() *Require {
	return &Require{
		modMap: make(map[string]*ModVersionConstraint),
	}
}

func (r *Require) initialise() error {
	var diags hcl.Diagnostics
	r.modMap = make(map[string]*ModVersionConstraint)

	if r.SteampipeVersionString != "" {
		steampipeVersion, err := semver.NewVersion(strings.TrimPrefix(r.SteampipeVersionString, "v"))
		if err != nil {
			return fmt.Errorf("invalid required steampipe version %s", r.SteampipeVersionString)
		}

		r.SteampipeVersion = steampipeVersion
	}

	for _, p := range r.Plugins {
		moreDiags := p.Initialise()
		diags = append(diags, moreDiags...)
	}
	for _, m := range r.Mods {
		moreDiags := m.Initialise()
		diags = append(diags, moreDiags...)
		if !diags.HasErrors() {
			// key map entry by name [and alias]
			r.modMap[m.Name] = m
		}
	}
	return plugin.DiagsToError("failed to initialise Require struct", diags)
}

func (r *Require) ValidateSteampipeVersion(modName string) error {
	if r.SteampipeVersion != nil {
		if version.SteampipeVersion.LessThan(r.SteampipeVersion) {
			return fmt.Errorf("steampipe version %s does not satisfy %s which requires version %s", version.SteampipeVersion.String(), modName, r.SteampipeVersion.String())
		}
	}
	return nil
}

// validates that for every plugin requirement there's at least one plugin installed
func (r *Require) ValidatePluginVersions(modName string, plugins *versionfile.PluginVersionFile) error {
	if len(r.Plugins) == 0 {
		return nil
	}
	errors := []error{}
	for _, requiredPlugin := range r.Plugins {
		if err := r.searchInstalledPluginForRequirement(modName, requiredPlugin, plugins); err != nil {
			errors = append(errors, err)
		}
	}
	return error_helpers.CombineErrors(errors...)
}

func (r *Require) searchInstalledPluginForRequirement(modName string, requirement *PluginVersion, plugins *versionfile.PluginVersionFile) error {
	for _, installed := range plugins.Plugins {
		if satisfied, err := r.isRequirementSatisfiedByInstalled(requirement, installed); satisfied && err != nil {
			return nil
		}
	}
	return sperr.New("could not find plugin which satisfies requirement '%s' in '%s'", requirement.RawName, modName)
}

func (r *Require) isRequirementSatisfiedByInstalled(requirement *PluginVersion, installed *versionfile.InstalledVersion) (bool, error) {
	org, name, _ := ociinstaller.NewSteampipeImageRef(installed.Name).GetOrgNameAndStream()
	if org == requirement.Org && name == requirement.Name {
		requiredVersion := requirement.Version
		installedVersion, err := semver.NewVersion(installed.Version)
		if err != nil {
			return false, err
		}
		// constraint specifies a minimum version
		return requiredVersion.LessThan(installedVersion), nil
	}
	return false, nil
}

// AddModDependencies adds all the mod in newModVersions to our list of mods, using the following logic
// - if a mod with same name, [alias] and constraint exists, it is not added
// - if a mod with same name [and alias] and different constraint exist, it is replaced
func (r *Require) AddModDependencies(newModVersions map[string]*ModVersionConstraint) {
	// rebuild the Mods array

	// first rebuild the mod map
	for name, newVersion := range newModVersions {
		// todo take alias into account
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
		// todo take alias into account
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
	return r.SteampipeVersion == nil && len(r.Mods) == 0 && len(r.Plugins) == 0
}
