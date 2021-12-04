package modconfig

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/version"
)

// Requires is a struct representing mod dependencies
type Requires struct {
	SteampipeVersionString string `hcl:"steampipe,optional"`
	SteampipeVersion       *semver.Version
	Plugins                []*PluginVersion `hcl:"plugin,block"`
	Mods                   []*ModVersion    `hcl:"mod,block"`
	DeclRange              hcl.Range        `json:"-"`
}

func (r *Requires) ValidateSteampipeVersion(modName string) error {
	if r.SteampipeVersion != nil {
		if version.SteampipeVersion.LessThan(r.SteampipeVersion) {
			return fmt.Errorf("steampipe version %s does not satisfy %s which requires version %s", version.SteampipeVersion.String(), modName, r.SteampipeVersion.String())
		}
	}
	return nil
}

func (r *Requires) Initialise() hcl.Diagnostics {
	var diags hcl.Diagnostics

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
	}
	return diags
}
