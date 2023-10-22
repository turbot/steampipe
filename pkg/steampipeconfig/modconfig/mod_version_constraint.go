package modconfig

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
	Args map[string]cty.Value `cty:"args"  hcl:"args,optional"`
	// only one of Constraint, Branch and FilePath will be set
	Constraint *versionhelpers.Constraints
	// the branch to use for the mod
	Branch string
	// the local file location to use
	FilePath string
	// contains the range of the definition of the mod block
	DefRange hcl.Range
	// contains the range of the body of the mod block
	BodyRange hcl.Range
	// contains the range of the total version field
	VersionRange hcl.Range
}

// NewModVersionConstraint creates a new ModVersionConstraint - this is called when installing a mod
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
		} else {
			gitUrl, err := url.Parse(segments[0])
			if err != nil {
				return nil, fmt.Errorf("invalid mod url %s", modFullName)
			}
			// a git url with branch reference has been parsed
			branchRef := strings.Split(gitUrl.Path, "/tree/")
			// TODO: add check for commit reference
			// commitRef := strings.Split(gitUrl.Path, "/commit/")
			if len(branchRef) == 2 {
				m.Name = branchRef[0]
				// set the branch name
				m.Branch = branchRef[1]
			}
		}
	}
	// try to convert version into a semver constraint
	if err := m.Initialise(nil); err != nil {
		return nil, err
	}
	return m, nil
}

// Initialise parses the version and name properties
func (m *ModVersionConstraint) Initialise(block *hcl.Block) hcl.Diagnostics {
	if block != nil {
		// record all the ranges in the source file
		m.DefRange = block.DefRange
		m.BodyRange = block.Body.(*hclsyntax.Body).SrcRange
		// record the range of the version attribute in this structure
		if versionAttribute, ok := block.Body.(*hclsyntax.Body).Attributes["version"]; ok {
			m.VersionRange = versionAttribute.SrcRange
		}
	}

	if strings.HasPrefix(m.Name, filePrefix) {
		m.setFilePath()
		return nil
	}

	// now default the version string to latest
	if m.VersionString == "" || m.VersionString == "latest" {
		m.VersionString = "*"
	}
	if m.VersionString == "" && m.Branch != "" {
		return nil
	}
	// does the version parse as a semver version
	if c, err := versionhelpers.NewConstraint(m.VersionString); err == nil {
		// no error
		m.Constraint = c
		return nil
	}

	// so there was an error
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("invalid mod version %s", m.VersionString),
		Subject:  &m.DefRange,
	}}

}

func (m *ModVersionConstraint) DependencyPath() string {
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
	return m.DependencyPath()
}

func (m *ModVersionConstraint) setFilePath() {
	m.FilePath = strings.TrimPrefix(m.FilePath, filePrefix)
}

func (m *ModVersionConstraint) Equals(other *ModVersionConstraint) bool {
	// just check the hcl properties
	return m.Name == other.Name && m.VersionString == other.VersionString
}
