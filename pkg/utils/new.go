package utils

import (
	"github.com/Masterminds/semver/v3"
)

type PluginVersion struct {
	Version string
}

func (p *PluginVersion) IsLocal() bool {
	return p.Version == "local"
}

func (p *PluginVersion) IsSemver() bool {
	_, err := semver.NewVersion(p.Version)
	return err == nil
}

func (p *PluginVersion) Semver() *semver.Version {
	if smv, err := semver.NewVersion(p.Version); err == nil {
		return smv
	}
	return nil
}

func (p *PluginVersion) String() string {
	return p.Version
}
