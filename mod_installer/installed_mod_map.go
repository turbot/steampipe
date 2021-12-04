package mod_installer

import (
	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type InstalledModMap map[string][]*semver.Version

func (i InstalledModMap) GetVersionSatisfyingRequirement(requiredVersion *modconfig.ModVersion) *semver.Version {
	// is this mod installed
	modVersions, ok := i[requiredVersion.Name]
	if !ok {
		return nil
	}
	for _, v := range modVersions {
		if requiredVersion.VersionConstraint.Check(v) {
			return v
		}
	}
	return nil
}

func (i InstalledModMap) Add(modName string, modVersion *semver.Version) {
	i[modName] = append(i[modName], modVersion)
}
