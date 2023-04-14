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
	"github.com/turbot/steampipe/pkg/version"
	"github.com/turbot/steampipe/sperr"
)

// TODO KAI SET RANGE

type SteampipeRequire struct {
	MinVersionString string `hcl:"min_version,optional"`
	Constraint       *semver.Constraints
	DeclRange        hcl.Range
}

func (r SteampipeRequire) initialise() hcl.Diagnostics {
	constraint, err := semver.NewConstraint(fmt.Sprintf(">=%s", strings.TrimPrefix(r.MinVersionString, "v")))
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid required steampipe version %s", r.MinVersionString),
				Subject:  &r.DeclRange,
			}}
	}

	r.Constraint = constraint
	return nil
}

// Require is a struct representing mod dependencies
type Require struct {
	Plugins   []*PluginVersion        `hcl:"plugin,block"`
	Steampipe *SteampipeRequire       `hcl:"steampipe,block"`
	SteampipeVersionString string `hcl:"steampipe,optional"`
	Mods                   []*ModVersionConstraint
	modMap                 map[string]*ModVersionConstraint
	DeclRange              hcl.Range
}

func NewRequire() *Require {
	return &Require{
		modMap: make(map[string]*ModVersionConstraint),
	}
}

func (r *Require) initialise() error {
	var diags hcl.Diagnostics
	r.modMap = make(map[string]*ModVersionConstraint)

	if r.Steampipe != nil && r.Steampipe.MinVersionString != "" {
		moreDiags := r.Steampipe.initialise()
		diags = append(diags, moreDiags...)
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
	if steampipeVersionConstraint := r.SteampipeVersionConstraint(); steampipeVersionConstraint != nil {
		if !steampipeVersionConstraint.Check(version.SteampipeVersion) {
			return fmt.Errorf("steampipe version %s does not satisfy %s which requires version %s", version.SteampipeVersion.String(), modName, r.Steampipe.MinVersionString)
		}
	}
	return nil
}

// validates that for every plugin requirement there's at least one plugin installed
func (r *Require) ValidatePluginVersions(modName string, plugins map[string]*semver.Version) error {
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

func (r *Require) searchInstalledPluginForRequirement(modName string, requirement *PluginVersion, plugins map[string]*semver.Version) error {
	for installedName, installed := range plugins {
		org, name, _ := ociinstaller.NewSteampipeImageRef(installedName).GetOrgNameAndStream()
		if org != requirement.Org || name != requirement.Name {
			// no point check - different plugin
			continue
		}
		if requirement.Version.LessThan(installed) || requirement.Version.Equal(installed) {
			return nil
		}
	}
	return sperr.New("could not find plugin which satisfies requirement '%s@%s' in '%s'", requirement.RawName, requirement.VersionString, modName)
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
	return r.SteampipeVersionConstraint == nil && len(r.Mods) == 0 && len(r.Plugins) == 0
}

func (r *Require) SteampipeVersionConstraint() *semver.Constraints {
	if r.Steampipe == nil {
		return nil
	}
	return r.Steampipe.Constraint

}
