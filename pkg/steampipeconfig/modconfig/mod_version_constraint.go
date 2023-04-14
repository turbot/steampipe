package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/versionhelpers"
	"github.com/zclconf/go-cty/cty"
)

const filePrefix = "file:"

type VersionConstrainCollection []*ModVersionConstraint

type ModVersionConstraint struct {
	// the fully qualified mod name, e.g. github.com/turbot/mod1
	Name string `cty:"name" hcl:"name,label"`
	VersionString    string `cty:"version" hcl:"version,optional"`
	MinVersionString string `cty:"min_version" hcl:"min_version,optional"`
	// variable values to be set on the dependency mod
	Args map[string]cty.Value `cty:"args"  hcl:"args,optional"`
	// only one of Constraint, Branch and FilePath will be set
	Constraint *versionhelpers.Constraints
	// the branch to use
	Branch string
	// the local file location to use
	FilePath  string
	DeclRange hcl.Range
}

func NewModVersionConstraint(modFullName string) (*ModVersionConstraint, error) {
	m := &ModVersionConstraint{
		Name: modFullName,
		Args: make(map[string]cty.Value),
	}

	// try to convert version into a semver constraint
	if err := m.Initialise(); err != nil {
		return nil, err
	}
	return m, nil
}

// Initialise parses the version and name properties
func (m *ModVersionConstraint) Initialise() hcl.Diagnostics {
	// if name has `file:` prefix, just set the name and ignore version
	if strings.HasPrefix(m.Name, filePrefix) {
		m.setFilePath()
		return nil
	}

	// otherwise try to extract version from name
	segments := strings.Split(m.Name, "@")
	if len(segments) > 2 {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid mod name %s", m.Name),
			Subject:  &m.DeclRange,
		}}

	}
	m.Name = segments[0]
	if len(segments) == 2 {
		// if MinVersionString is already set, error
		if m.MinVersionString != "" {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("both 'min_version' and a version constraint in the mod name are set"),
				Subject:  &m.DeclRange,
			}}
		}
		m.MinVersionString = segments[1]
	}

	if m.VersionString != "" {
		if m.MinVersionString != "" {
			var msg string
			// is a version specified in the mod name?
			if len(segments) == 2 {
				msg = "both 'version' and a version constraint in the mod name are set"
			} else {
				msg = "both 'version' and 'min_version' are set"
			}
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  msg,
				Subject:  &m.DeclRange,
			}}
		}

		// otherwise just copy over to MinVersionString
		m.MinVersionString = m.VersionString
	}

	// now default the version string to latest
	if m.MinVersionString == "" {
		m.MinVersionString = "latest"
	}

	if m.MinVersionString == "latest" {
		m.Constraint, _ = versionhelpers.NewConstraint("*")
		return nil
	}
	// does the version parse as a semver version
	if c, err := versionhelpers.NewConstraint(m.MinVersionString); err == nil {
		// no error
		m.Constraint = c
		return nil
	}

	// todo handle branch and commit hash

	// so there was an error
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("invalid mod version %s", m.MinVersionString),
		Subject:  &m.DeclRange,
	}}

}

func (m *ModVersionConstraint) FullName() string {
	if m.HasVersion() {
		return fmt.Sprintf("%s@%s", m.Name, m.MinVersionString)
	}
	return m.Name
}

// HasVersion returns whether the mod has a version specified, or is the latest
// if no version is specified, or the version is "latest", this is the latest version
func (m *ModVersionConstraint) HasVersion() bool {
	return !helpers.StringSliceContains([]string{"", "latest", "*"}, m.MinVersionString)
}

func (m *ModVersionConstraint) String() string {
	return m.FullName()
}

func (m *ModVersionConstraint) setFilePath() {
	m.FilePath = strings.TrimPrefix(m.FilePath, filePrefix)
}

func (m *ModVersionConstraint) Equals(other *ModVersionConstraint) bool {
	// just check the hcl properties
	return m.Name == other.Name && m.MinVersionString == other.MinVersionString
}
