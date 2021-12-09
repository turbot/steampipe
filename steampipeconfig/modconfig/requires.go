package modconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/version"
)

// Requires is a struct representing mod dependencies
type Requires struct {
	SteampipeVersionString string `hcl:"steampipe,optional"`
	SteampipeVersion       *semver.Version
	Plugins                []*PluginVersion        `hcl:"plugin,block"`
	Mods                   []*ModVersionConstraint `hcl:"mod,block"`
	DeclRange              hcl.Range               `json:"-"`
	// map keyed by name [and alias]
	modMap map[string]*ModVersionConstraint
}

func newRequires() *Requires {
	r := &Requires{}
	r.initialise()
	return r
}

func (r *Requires) initialise() hcl.Diagnostics {
	var diags hcl.Diagnostics
	r.modMap = make(map[string]*ModVersionConstraint)

	if r.SteampipeVersionString != "" {
		steampipeVersion, err := semver.NewVersion(strings.TrimPrefix(r.SteampipeVersionString, "v"))
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid required steampipe version %s", r.SteampipeVersionString),
				Subject:  &r.DeclRange,
			})
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
	return diags
}

func (r *Requires) ValidateSteampipeVersion(modName string) error {
	if r.SteampipeVersion != nil {
		if version.SteampipeVersion.LessThan(r.SteampipeVersion) {
			return fmt.Errorf("steampipe version %s does not satisfy %s which requires version %s", version.SteampipeVersion.String(), modName, r.SteampipeVersion.String())
		}
	}
	return nil
}

// AddModDependencies adds all the mod in newModVersions to our list of mods, using the following logic
// - if a mod with same name, [alias] and constraint exists, it is not added
// - if a mod with same name [and alias] and different constraint exist, it is replaced
func (r *Requires) AddModDependencies(newModVersions map[string]*ModVersionConstraint) {
	// rebuild the Mods array

	for name, newVersion := range newModVersions {
		// todo take alias into account

		// if this existing mod is being replaced (i.e. is is in newModVersions), skip
		if existingVersion, ok := r.modMap[name]; ok {
			if existingVersion.Constraint.Equals(newVersion.Constraint) {
				continue
			}
			// so the contraints are different - fall through to update the stored version
		}
		r.modMap[name] = newVersion
	}

	// now update the mod array from teh map
	var newMods = make([]*ModVersionConstraint, len(r.modMap))
	idx := 0
	for _, requiredVersion := range r.modMap {
		newMods[idx] = requiredVersion
	}
	// sort by name
	sort.Sort(ModVersionConstraintCollection(newMods))
	// write back
	r.Mods = newMods
}

func (r *Requires) GetModDependency(name string /*,alias string*/) *ModVersionConstraint {
	return r.modMap[name]
}

func (r *Requires) ContainsMod(requiredModVersion *ModVersionConstraint) bool {
	if c := r.GetModDependency(requiredModVersion.Name); c != nil {
		return c.Equals(requiredModVersion)
	}
	return false
}
