package steampipeconfig

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
)

type LoadModOption = func(mod *LoadModConfig)
type LoadModConfig struct {
	Version        *semver.Version
	DependencyPath *string
	DependencyName string
}

func WithDependencyConfig(modDependencyName string, version *semver.Version) LoadModOption {
	return func(mod *LoadModConfig) {
		mod.Version = version
		// build the ModDependencyPath from the modDependencyName and the version
		dependencyPath := fmt.Sprintf("%s@v%s", modDependencyName, version.String())
		mod.DependencyPath = &dependencyPath
		mod.DependencyName = modDependencyName
	}
}
