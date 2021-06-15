package modconfig

import (
	"fmt"

	"github.com/blang/semver"

	"github.com/hashicorp/hcl/v2"
)

type PluginVersion struct {
	// the fully qualified plugin name, e.g. github.com/turbot/mod1
	Name string `cty:"name" hcl:"name,label"`
	// the version STREAM, can be either a major or minor version stream i.e. 1 or 1.1
	Version       string `cty:"version" hcl:"version,optional"`
	DeclRange     hcl.Range
	parsedVersion *semver.Version
}

func (p *PluginVersion) FullName() string {
	if p.Version == "" {
		return p.Name
	}
	return fmt.Sprintf("%s@%s", p.Name, p.Version)
}

func (p *PluginVersion) String() string {
	return p.FullName()
}

func (p *PluginVersion) Validate() error {
	v, err := semver.New(p.Version)
	if err != nil {
		return err
	}
	p.parsedVersion = v
	return nil
}
