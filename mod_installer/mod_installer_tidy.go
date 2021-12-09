package mod_installer

import (
	"github.com/turbot/steampipe/steampipeconfig/version_map"
)

func (i *ModInstaller) Tidy() (version_map.VersionListMap, error) {
	// TODO
	//unusedMods := i.installData.getUnusedMods()
	//// now delete any mod folders which are not in the lock file
	//for name, versions := range unusedMods {
	//	for _, version := range versions {
	//		depPath := i.getDependencyDestPath(modVersionFullName(name, version))
	//		if err := os.RemoveAll(depPath); err != nil {
	//			return nil, err
	//		}
	//	}
	//}
	//// TODO remove empty folders
	//return unusedMods, nil
	return nil, nil
}
