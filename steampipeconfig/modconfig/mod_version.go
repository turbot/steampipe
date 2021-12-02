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
	Name          string  `cty:"name" hcl:"name,label"`
	VersionString string  `cty:"version" hcl:"version"`
	Alias         *string `cty:"alias" hcl:"alias,optional"`

	// only one of VersionConstraint, Branch and FilePath will be set
	VersionConstraint goVersion.Constraints
	// the branch to use
	Branch string
	// the local file location to use
	FilePath string

	DeclRange hcl.Range
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

	if strings.HasPrefix(m.VersionString, "file:") {
		m.FilePath = m.VersionString
		return diags
	}
	// does the version parse as a semver version
	if v, err := goVersion.NewConstraint(m.VersionString); err == nil {
		m.VersionConstraint = v
		return diags
	}

	// otherwise assume it is a branch
	m.Branch = m.VersionString

	return diags
}
