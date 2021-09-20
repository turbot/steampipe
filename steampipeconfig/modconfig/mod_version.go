package modconfig

import (
	"fmt"
	"strings"

	goVersion "github.com/hashicorp/go-version"
	typehelpers "github.com/turbot/go-kit/types"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/go-kit/helpers"
)

type ModVersion struct {
	// the fully qualified mod name, e.g. github.com/turbot/mod1
	Name          string `cty:"name" hcl:"name,label"`
	VersionString string `cty:"version" hcl:"version"`
	Version       *goVersion.Version
	Alias         *string `cty:"alias" hcl:"alias,optional"`
	DeclRange     hcl.Range
}

func (m *ModVersion) FullName() string {
	if m.HasVersion() {
		return fmt.Sprintf("%s@%s", m.Name, m.VersionString)
	}
	return m.Name
}

// HasVersion returns whether the mod has a version specified, or is the latest
// if no version is specified, or the version is "latest", this is the latest version
func (m *ModVersion) HasVersion() bool {
	return !helpers.StringSliceContains([]string{"", "latest"}, m.VersionString)
}

func (m *ModVersion) String() string {
	if alias := typehelpers.SafeString(m.Alias); alias != "" {
		return fmt.Sprintf("mod %s (%s)", m.FullName(), alias)
	}
	return fmt.Sprintf("mod %s", m.FullName())
}

// Initialise parses the version and name properties
func (m *ModVersion) Initialise() hcl.Diagnostics {
	var diags hcl.Diagnostics
	if version, err := goVersion.NewVersion(strings.TrimPrefix(m.VersionString, "v")); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid mod version %s", m.VersionString),
			Subject:  &m.DeclRange,
		})
	} else {
		m.Version = version
	}

	return diags
}
