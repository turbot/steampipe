package modconfig

import (
	"fmt"

	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/ociinstaller"
)

type PluginVersion struct {
	// the plugin name, as specified in the mod requires block. , e.g. turbot/mod1, aws
	RawName string `cty:"name" hcl:"name,label"`
	// the version STREAM, can be either a major or minor version stream i.e. 1 or 1.1
	Version       string           `cty:"version" hcl:"version,optional"`
	ParsedVersion *version.Version `json:"-"`
	// the org and name which are parsed from the raw name
	Org       string
	Name      string
	DeclRange hcl.Range `json:"-"`
}

func (p *PluginVersion) FullName() string {
	if p.Version == "" {
		return p.ShortName()
	}
	return fmt.Sprintf("%s@%s", p.ShortName(), p.Version)
}

func (p *PluginVersion) ShortName() string {
	return fmt.Sprintf("%s/%s", p.Org, p.Name)
}

func (p *PluginVersion) String() string {
	return p.FullName()
}

// parse the version and name properties
func (p *PluginVersion) parseProperties() error {
	v, err := version.NewVersion(p.Version)

	if err != nil {
		return err
	}
	p.ParsedVersion = v
	// parse plugin name
	p.Org, p.Name, _ = ociinstaller.NewSteampipeImageRef(p.RawName).GetOrgNameAndStream()
	return nil
}
