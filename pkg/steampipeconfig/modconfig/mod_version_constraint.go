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
	Name          string `cty:"name" hcl:"name,label"`
	VersionString string `cty:"version" hcl:"version"`
	// variable values to be set on the dependency mod
	Args map[string]cty.Value `cty:"args"`
	// only one of Constraint, Branch and FilePath will be set
	Constraint *versionhelpers.Constraints
	// // NOTE: aliases will be supported in the future
	//Alias string `cty:"alias" hcl:"alias"`
	// the branch to use
	Branch string
	// the local file location to use
	FilePath  string
	DeclRange hcl.Range
}

func NewModVersionConstraint(modFullName string) (*ModVersionConstraint, error) {
	m := &ModVersionConstraint{
		Args: make(map[string]cty.Value),
	}

	// if name has `file:` prefix, just set the name and ignore version
	if strings.HasPrefix(modFullName, filePrefix) {
		m.Name = modFullName
	} else {
		// otherwise try to extract version from name
		segments := strings.Split(modFullName, "@")
		if len(segments) > 2 {
			return nil, fmt.Errorf("invalid mod name %s", modFullName)
		}
		m.Name = segments[0]
		if len(segments) == 2 {
			m.VersionString = segments[1]
		}
	}

	// try to convert version into a semver constraint
	if err := m.Initialise(); err != nil {
		return nil, err
	}
	return m, nil
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
	return !helpers.StringSliceContains([]string{"", "latest", "*"}, m.VersionString)
}

func (m *ModVersionConstraint) String() string {
	return m.FullName()
}

// Initialise parses the version and name properties
func (m *ModVersionConstraint) Initialise() hcl.Diagnostics {
	if strings.HasPrefix(m.Name, filePrefix) {
		m.setFilePath()
		return nil
	}
	var diags hcl.Diagnostics

	if m.VersionString == "" {
		m.Constraint, _ = versionhelpers.NewConstraint("*")
		m.VersionString = "latest"
		return diags
	}
	if m.VersionString == "latest" {
		m.Constraint, _ = versionhelpers.NewConstraint("*")
		return diags
	}
	// does the version parse as a semver version
	if c, err := versionhelpers.NewConstraint(m.VersionString); err == nil {
		// no error
		m.Constraint = c
		return diags
	}

	// todo handle branch and commit hash

	// so there was an error
	diags = append(diags, &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("invalid mod version %s", m.VersionString),
		Subject:  &m.DeclRange,
	})
	return diags
}

func (m *ModVersionConstraint) setFilePath() {
	m.FilePath = strings.TrimPrefix(m.FilePath, filePrefix)
}

func (m *ModVersionConstraint) Equals(other *ModVersionConstraint) bool {
	// just check the hcl properties
	return m.Name == other.Name && m.VersionString == other.VersionString
}
