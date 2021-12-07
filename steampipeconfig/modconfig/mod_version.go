package modconfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/version"
)

type ModVersionConstraint struct {
	// the fully qualified mod name, e.g. github.com/turbot/mod1
	Name          string `cty:"name" hcl:"name,label"`
	VersionString string `cty:"version" hcl:"version"`
	//Alias         *string `cty:"alias" hcl:"alias,optional"`

	// only one of Constraint, Branch and FilePath will be set
	Constraint *version.Constraints
	// the branch to use
	Branch string
	// the local file location to use
	FilePath string

	DeclRange hcl.Range
}

func NewModVersionConstraint(modFullName string) (*ModVersionConstraint, error) {
	segments := strings.Split(modFullName, "@")
	if len(segments) > 2 {
		return nil, fmt.Errorf("invalid mod name %s", modFullName)
	}
	v := &ModVersionConstraint{Name: segments[0]}
	if len(segments) == 2 {
		v.VersionString = segments[1]
	}
	if err := v.Initialise(); err != nil {
		return nil, err
	}
	return v, nil
}

func (m *ModVersionConstraint) FullName() string {
	if m.HasVersion() {
		return fmt.Sprintf("%s@%s", m.Name, m.VersionString)
	}
	return m.Name
}

// HasVersion returns whether the mod has a version specified, or is the latest
// if no version is specified, or the version is "latest", this is the latest version
func (m *ModVersionConstraint) HasVersion() bool {
	return !helpers.StringSliceContains([]string{"", "latest"}, m.VersionString)
}

func (m *ModVersionConstraint) String() string {
	//if alias := typehelpers.SafeString(m.Alias); alias != "" {
	//	return fmt.Sprintf("%s (%s)", m.FullName(), alias)
	//}
	return fmt.Sprintf("%s", m.FullName())
}

// Initialise parses the version and name properties
func (m *ModVersionConstraint) Initialise() hcl.Diagnostics {
	diags := m.cleanName()
	if diags != nil {
		return diags
	}

	if strings.HasPrefix(m.VersionString, "file:") {
		m.FilePath = m.VersionString
		return diags
	}

	// TODO ignore prerelease
	if m.VersionString == "" {
		m.Constraint, _ = version.NewConstraint("*")
		m.VersionString = "latest"
		return diags
	}
	if m.VersionString == "latest" {
		m.Constraint, _ = version.NewConstraint("*")
		return diags
	}
	// does the version parse as a semver version
	if c, err := version.NewConstraint(m.VersionString); err == nil {
		m.Constraint = c
		v, _ := semver.NewVersion("1.1")
		a := c.Check(v)
		log.Println(a)

		return diags
	}

	// todo handle branch and commit hash
	diags = append(diags, &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("invalid mod version %s", m.VersionString),
		Subject:  &m.DeclRange,
	})
	return diags
}

func (m *ModVersionConstraint) cleanName() hcl.Diagnostics {
	segments := strings.Split(m.Name, "/")
	l := len(segments)
	if l == 3 {
		// leave as is
		return nil
	}
	if l == 1 {
		turbotGithubPrefix := "github.com/turbot/"
		modNamePrefix := "steampipe-mod-"
		if !strings.HasPrefix(m.Name, modNamePrefix) {
			m.Name = fmt.Sprintf("%s%s%s", turbotGithubPrefix, modNamePrefix, m.Name)
		} else {
			m.Name = fmt.Sprintf("%s%s", turbotGithubPrefix, m.Name)
		}
		return nil
	}

	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("invalid mod name %s", m.Name),
	}}
}
