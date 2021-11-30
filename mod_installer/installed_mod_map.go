package mod_installer

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// InstalledModMap is is map keyed by mod name storing a list of all the mod version installed for each mod
type InstalledModMap map[string]semver.Collection

func (i InstalledModMap) GetVersionSatisfyingRequirement(requiredVersion *modconfig.ModVersionConstraint) *semver.Version {
	// is this mod installed
	modVersions, ok := i[requiredVersion.Name]
	if !ok {
		return nil
	}
	for _, v := range modVersions {
		if requiredVersion.Constraint.Check(v) {
			return v
		}
	}
	return nil
}

func (i InstalledModMap) Add(modName string, modVersion *semver.Version) {
	versions := append(i[modName], modVersion)
	// reverse sort the versions
	sort.Sort(sort.Reverse(versions))
	i[modName] = versions

}

// FlatMap converts the InstalledModMap map into a bool map keyed by qualified mod name
func (m InstalledModMap) FlatMap() map[string]bool {
	var res = make(map[string]bool)
	for name, versions := range m {
		for _, version := range versions {
			key := fmt.Sprintf("%s@%s", name, version)
			res[key] = true
		}
	}
	return res
}
