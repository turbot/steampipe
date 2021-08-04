package modconfig

import (
	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/go-kit/helpers"
)

type ModVersion struct {
	// the fully qualified mod name, e.g. github.com/turbot/mod1
	// TODO think about names
	ShortName string `hcl:"name,label"`
	FullName  string `cty:"name"`

	Version   string    `cty:"version" hcl:"version"`
	Alias     *string   `cty:"alias" hcl:"alias,optional"`
	DeclRange hcl.Range `json:"-"`
}

// Name returns Name@Version
func (m *ModVersion) Name() string {
	// TODO what about mod version in name?
	return m.FullName
}

// HasVersion returns whether the mod has a version specified, or is the latest
// if no version is specified, or the version is "latest", this is the latest version
func (m *ModVersion) HasVersion() bool {
	return !helpers.StringSliceContains([]string{"", "latest"}, m.Version)
}

func (m *ModVersion) String() string {
	return m.Name()
}
