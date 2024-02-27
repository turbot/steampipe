package modconfig

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/steampipe/pkg/ociinstaller"
)

type PluginVersion struct {
	// the plugin name, as specified in the mod requires block. , e.g. turbot/mod1, aws
	RawName string `cty:"name" hcl:"name,label"`
	// the minumum version which satisfies the requirement
	MinVersionString string `cty:"min_version" hcl:"min_version,optional"`
	Constraint       *semver.Constraints
	// the org and name which are parsed from the raw name
	Org       string
	Name      string
	DeclRange hcl.Range
}

func (p *PluginVersion) FullName() string {
	if p.MinVersionString == "" {
		return p.ShortName()
	}
	return fmt.Sprintf("%s@%s", p.ShortName(), p.MinVersionString)
}

func (p *PluginVersion) ShortName() string {
	return fmt.Sprintf("%s/%s", p.Org, p.Name)
}

func (p *PluginVersion) String() string {
	return fmt.Sprintf("plugin %s", p.FullName())
}

// Initialise parses the version and name properties
func (p *PluginVersion) Initialise(block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	p.DeclRange = hclhelpers.BlockRange(block)

	// convert min version into constraint (including prereleases)
	minVersion, err := semver.NewVersion(strings.TrimPrefix(p.MinVersionString, "v"))
	if err == nil {
		p.Constraint, err = semver.NewConstraint(fmt.Sprintf(">=%s-0", minVersion))
	}
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Invalid plugin version %s", p.MinVersionString),
			Subject:  &p.DeclRange,
		})
	}
	// parse plugin name
	p.Org, p.Name, _ = ociinstaller.NewSteampipeImageRef(p.RawName).GetOrgNameAndStream()

	return diags
}
