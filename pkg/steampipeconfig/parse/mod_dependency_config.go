package parse

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type ModDependencyConfig struct {
	Version        *semver.Version
	DependencyPath *string
	DependencyName string
}

func (c ModDependencyConfig) SetModProperties(mod *modconfig.Mod) {
	mod.Version = c.Version
	mod.DependencyPath = c.DependencyPath
	mod.DependencyName = c.DependencyName
}

func NewDependencyConfig(modDependency *modconfig.ModVersionConstraint, version *semver.Version) *ModDependencyConfig {
	d := fmt.Sprintf("%s@v%s", modDependency.Name, version.String())
	return &ModDependencyConfig{Version: version,
		DependencyPath: &d,
		DependencyName: modDependency.Name,
	}
}
