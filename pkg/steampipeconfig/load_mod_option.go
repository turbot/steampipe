package steampipeconfig

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type LoadModOption = func(mod *modconfig.Mod)

func WithDependencyConfig(modDependencyName string, version *semver.Version) LoadModOption {
	return func(mod *modconfig.Mod) {
		mod.Version = version
		// build the ModDependencyPath from the modDependencyName and the version
		dependencyPath := fmt.Sprintf("%s@v%s", modDependencyName, version.String())
		mod.DependencyPath = &dependencyPath
		mod.DependencyName = modDependencyName
	}
}
