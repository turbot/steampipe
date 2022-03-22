package modconfig

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
)

func ModVersionFullName(name string, version *semver.Version) string {
	if version == nil {
		return name
	}
	versionString := GetMonotonicVersionString(version)
	return fmt.Sprintf("%s@v%s", name, versionString)
}

func GetMonotonicVersionString(v *semver.Version) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d.%d", v.Major(), v.Minor())
	if v.Metadata() != "" {
		fmt.Fprintf(&buf, "+%s", v.Metadata())
	}
	return buf.String()
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
