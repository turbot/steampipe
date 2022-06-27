package modconfig

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/ociinstaller"
)

type PluginVersion struct {
	// the plugin name, as specified in the mod requires block. , e.g. turbot/mod1, aws
	RawName string `cty:"name" hcl:"name,label"`
	// the version STREAM, can be either a major or minor version stream i.e. 1 or 1.1
	VersionString string `cty:"version" hcl:"version,optional"`
	Version       *semver.Version
	// the org and name which are parsed from the raw name
	Org       string
	Name      string
	DeclRange hcl.Range `json:"-"`
}

func (p *PluginVersion) FullName() string {
	if p.VersionString == "" {
		return p.ShortName()
	}
	return fmt.Sprintf("%s@%s", p.ShortName(), p.VersionString)
}

func (p *PluginVersion) ShortName() string {
	return fmt.Sprintf("%s/%s", p.Org, p.Name)
}

func (p *PluginVersion) String() string {
	return fmt.Sprintf("plugin %s", p.FullName())
}

// Initialise parses the version and name properties
func (p *PluginVersion) Initialise() hcl.Diagnostics {
	var diags hcl.Diagnostics
	if version, err := semver.NewVersion(strings.TrimPrefix(p.VersionString, "v")); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid plugin version %s", p.VersionString),
			Subject:  &p.DeclRange,
		})
	} else {
		p.Version = version
	}

	// parse plugin name
	p.Org, p.Name, _ = ociinstaller.NewSteampipeImageRef(p.RawName).GetOrgNameAndStream()

	return diags
}
