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

const filePrefix = "file:"

type VersionConstrainCollection []*ModVersionConstraint

type ModVersionConstraint struct {
	// the fully qualified mod name, e.g. github.com/turbot/mod1
	Name          string `cty:"name" hcl:"name,label"`
	VersionString string `cty:"version" hcl:"version"`
	// only one of Constraint, Branch and FilePath will be set
	Constraint *version.Constraints
	// // NOTE: aliases will be supported in the future
	//Alias string `cty:"alias" hcl:"alias"`
	// the branch to use
	Branch string
	// the local file location to use
	FilePath  string
	DeclRange hcl.Range
}

func NewModVersionConstraint(modFullName string) (*ModVersionConstraint, error) {
	var m *ModVersionConstraint
	// if name has `file:` prefix, just set the name and ignore version
	if strings.HasPrefix(modFullName, filePrefix) {
		m = &ModVersionConstraint{Name: modFullName}
	} else {
		// otherwise try to extract version from name
		segments := strings.Split(modFullName, "@")
		if len(segments) > 2 {
			return nil, fmt.Errorf("invalid mod name %s", modFullName)
		}
		m = &ModVersionConstraint{Name: segments[0]}
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
	return fmt.Sprintf("%s", m.FullName())
}

// Initialise parses the version and name properties
func (m *ModVersionConstraint) Initialise() hcl.Diagnostics {
	if strings.HasPrefix(m.Name, filePrefix) {
		m.setFilePath()
		return nil
	}

	diags := m.cleanName()
	if diags != nil {
		return diags
	}

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

func (m *ModVersionConstraint) setFilePath() {
	m.FilePath = strings.TrimPrefix(m.FilePath, filePrefix)
}

func (m *ModVersionConstraint) Equals(other *ModVersionConstraint) bool {
	// just check the hcl properties
	return m.Name == other.Name && m.VersionString == other.VersionString
}
