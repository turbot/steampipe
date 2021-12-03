package mod_installer

import (
	goVersion "github.com/hashicorp/go-version"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type InstalledModMap map[string][]*goVersion.Version

func (i InstalledModMap) GetVersionSatisfyingRequirement(requiredVersion *modconfig.ModVersion) *goVersion.Version {
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

func (i InstalledModMap) Add(modName string, modVersion *goVersion.Version) {
	i[modName] = append(i[modName], modVersion)
}
