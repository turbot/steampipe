package modconfig

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/go-kit/helpers"
)

type ModVersion struct {
	// the fully qualified mod name, e.g. github.com/turbot/mod1
	ShortName string
	FullName  string `hcl:"name"`

	Version   string  `hcl:"version"`
	Alias     *string `hcl:"alias"`
	DeclRange hcl.Range
}

func NewModVersion(block *hcl.Block) *ModVersion {
	return &ModVersion{
		ShortName: block.Labels[0],
		Version:   block.Labels[1],
		FullName:  fmt.Sprintf("mod.%s", block.Labels[0]),
		DeclRange: block.DefRange,
	}
}

// Schema :: hcl schema for control
func (m *ModVersion) Schema() *hcl.BodySchema {
	return buildAttributeSchema(m)
}

func (m *ModVersion) CtyValue() (cty.Value, error) {
	return getCtyValue(m)
}

// Name :: return Name@Version
func (m *ModVersion) Name() string {
	// TODO what about mod version in name?
	return m.FullName
}

// HasVersion :: if no version is specified, or the version is "latest", this is the latest version
func (m *ModVersion) HasVersion() bool {
	return !helpers.StringSliceContains([]string{"", "latest"}, m.Version)
}

func (m *ModVersion) String() string {
	return m.Name()
}
