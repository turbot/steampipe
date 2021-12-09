package modconfig

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
)

func ModVersionFullName(name string, version *semver.Version) string {
	return fmt.Sprintf("%s@%s", name, version.Original())
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
	modVersion, err = semver.NewVersion(parts[1])
	if err != nil {
		err = fmt.Errorf("mod file %s has invalid version", fullName)
	}
	return
}
