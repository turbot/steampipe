package mod_installer

import (
	"os"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func (i *ModInstaller) Tidy() (modconfig.VersionListMap, error) {
	unusedMods := i.installData.getUnusedMods()
	// now delete any mod folders which are not in the lock file
	for name, versions := range unusedMods {
		for _, version := range versions {
			depPath := i.getDependencyDestPath(modVersionFullName(name, version))
			if err := os.RemoveAll(depPath); err != nil {
				return nil, err
			}
		}
	}
	// TODO remove empty folders
	return unusedMods, nil
}
