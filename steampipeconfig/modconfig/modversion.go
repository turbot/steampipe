package modconfig

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
)

type ModVersion struct {
	// the fully qualified mod name, e.g. github.com/turbot/mod1
	Name    string  `hcl:"name"`
	Version string  `hcl:"version"`
	Alias   *string `hcl:"alias"`
}

// FullName :: return Name@Version
func (m *ModVersion) FullName() string {
	if m.Version == "" {
		return m.Name
	}
	return fmt.Sprintf("%s@%s", m.Name, m.Version)
}

// HasVersion :: if no version is specified, or the version is "latest", this is the latest version
func (m *ModVersion) HasVersion() bool {
	return !helpers.StringSliceContains([]string{"", "latest"}, m.Version)
}

func (m *ModVersion) String() string {
	return m.FullName()
}
