package modconfig

import (
	"fmt"
	"github.com/turbot/go-kit/hcl_helpers"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/ociinstaller"
)

type PluginVersion struct {
	// the plugin name, as specified in the mod requires block. , e.g. turbot/mod1, aws
	RawName string `cty:"name" hcl:"name,label"`
	// deprecated: use MinVersionString
	VersionString string `cty:"version" hcl:"version,optional"`
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
	p.DeclRange = hcl_helpers.BlockRange(block)
	// handle deprecation warnings/errors
	if p.VersionString != "" {
		if p.MinVersionString != "" {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Both 'min_version' and deprecated 'version' property are set",
				Subject:  &p.DeclRange,
			})
			return diags
		}
		// raise deprecation warning
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  fmt.Sprintf("Property 'version' is deprecated - use 'min_version' instead, in plugin '%s' require block", p.RawName),
			Subject:  &p.DeclRange,
		})
		// copy into new property
		p.MinVersionString = p.VersionString
	}

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
