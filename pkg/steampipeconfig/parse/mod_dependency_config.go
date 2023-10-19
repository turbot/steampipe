package parse

import (
	"fmt"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
)

type ModDependencyConfig struct {
	ModDependency  *versionmap.ResolvedVersionConstraint
	DependencyPath *string
}

func (c ModDependencyConfig) SetModProperties(mod *modconfig.Mod) {
	mod.Version = c.ModDependency.Version
	mod.DependencyPath = c.DependencyPath
	mod.DependencyName = c.ModDependency.Name
}

func NewDependencyConfig(modDependency *versionmap.ResolvedVersionConstraint) *ModDependencyConfig {
	d := fmt.Sprintf("%s@v%s", modDependency.Name, modDependency.Version.String())
	return &ModDependencyConfig{
		DependencyPath: &d,
		ModDependency:  modDependency,
	}
}
