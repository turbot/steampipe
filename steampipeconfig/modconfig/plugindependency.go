package modconfig

import (
	"fmt"
)

type PluginDependency struct {
	// the fully qualified plugin name, e.g. github.com/turbot/mod1
	Name string `hcl:"name"`
	// the version STREAM, can be either a major or minor version stream i.e. 1 or 1.1
	Version string `hcl:"version"`
}

func (p *PluginDependency) FullName() string {
	if p.Version == "" {
		return p.Name
	}
	return fmt.Sprintf("%s@%s", p.Name, p.Version)
}

func (p *PluginDependency) String() string {
	return p.FullName()
}
