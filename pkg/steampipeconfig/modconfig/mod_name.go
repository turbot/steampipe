package modconfig

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
)

func ModVersionFullName(name string, version *semver.Version) string {
	if version == nil {
		return name
	}

	return fmt.Sprintf("%s@v%s", name, version.String())
}

func ParseModFullName(fullName string) (modName string, modVersion *semver.Version, err error) {
	// we expect modLongName to be of form github.com/turbot/steampipe-mod-m2@v1.0
	// split to get the name and version
	parts := strings.Split(fullName, "@")
	if len(parts) != 2 {
		err = fmt.Errorf("invalid mod full name %s", fullName)
		return
	}
	modName = parts[0]
	versionString := parts[1]
	modVersion, err = semver.NewVersion(versionString)
	// NOTE: we expect the version to be in format 'vx.x.x', i.e. a smver with a preceding v
	if !strings.HasPrefix(versionString, "v") || err != nil {
		err = fmt.Errorf("mod file %s has invalid version", fullName)
	}
	return
}
