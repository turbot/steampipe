package utils

import (
	"github.com/Masterminds/semver/v3"
	"github.com/turbot/steampipe/sperr"
)

type PluginVersion struct {
	version string
	semver  *semver.Version
}

func NewPluginVersion(version string) (*PluginVersion, error) {
	if smv, err := semver.NewVersion(version); err == nil {
		pluginVersion := &PluginVersion{
			version: version,
			semver:  smv,
		}
		return pluginVersion, nil
	}
	if version == "local" {
		return LocalPluginVersion(), nil
	}
	return nil, sperr.New("version must be a valid semver or 'local'; got: %s", version)
}

func LocalPluginVersion() *PluginVersion {
	return &PluginVersion{
		version: "local",
	}
}

func (p *PluginVersion) IsLocal() bool {
	return p.semver == nil
}

func (p *PluginVersion) IsSemver() bool {
	return p.semver != nil
}

func (p *PluginVersion) Semver() *semver.Version {
	return p.semver
}

func (p *PluginVersion) String() string {
	return p.version
}
