package modconfig

import (
	"fmt"

	typehelpers "github.com/turbot/go-kit/types"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/go-kit/helpers"
)

type ModVersion struct {
	// the fully qualified mod name, e.g. github.com/turbot/mod1
	Name      string  `cty:"name" hcl:"name,label"`
	Version   string  `cty:"version" hcl:"version"`
	Alias     *string `cty:"alias" hcl:"alias,optional"`
	DeclRange hcl.Range
}

func (m *ModVersion) FullName() string {
	if m.HasVersion() {
		return fmt.Sprintf("%s@%s", m.Name, m.Version)
	}
	return m.Name
}

// HasVersion returns whether the mod has a version specified, or is the latest
// if no version is specified, or the version is "latest", this is the latest version
func (m *ModVersion) HasVersion() bool {
	return !helpers.StringSliceContains([]string{"", "latest"}, m.Version)
}

func (m *ModVersion) String() string {
	if alias := typehelpers.SafeString(m.Alias); alias != "" {
		return fmt.Sprintf("mod %s (%s)", m.FullName(), alias)
	}
	return fmt.Sprintf("mod %s", m.FullName())
}
