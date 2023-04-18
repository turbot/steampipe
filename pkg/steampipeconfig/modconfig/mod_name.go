package modconfig

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// BuildModDependencyPath converts a mod dependency name of form github.com/turbot/steampipe-mod-m2
//
//	and a version into a dependency path of form github.com/turbot/steampipe-mod-m2@v1.0.0
func BuildModDependencyPath(dependencyName string, version *semver.Version) string {
	if version == nil {
		// not expected
		return dependencyName
	}

	return fmt.Sprintf("%s@v%s", dependencyName, version.String())
}

// ParseModDependencyPath converts a mod depdency path of form github.com/turbot/steampipe-mod-m2@v1.0.0
// into the dependency name (github.com/turbot/steampipe-mod-m2) and version
func ParseModDependencyPath(fullName string) (modDependencyName string, modVersion *semver.Version, err error) {
	// split to get the name and version
	parts := strings.Split(fullName, "@")
	if len(parts) != 2 {
		err = fmt.Errorf("invalid mod full name %s", fullName)
		return
	}
	modDependencyName = parts[0]
	versionString := parts[1]
	modVersion, err = semver.NewVersion(versionString)
	// NOTE: we expect the version to be in format 'vx.x.x', i.e. a semver with a preceding v
	if !strings.HasPrefix(versionString, "v") || err != nil {
		err = fmt.Errorf("mod file %s has invalid version", fullName)
	}
	return
}
