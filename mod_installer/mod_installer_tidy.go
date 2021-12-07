package mod_installer

import (
	"fmt"
	"os"
)

func (i *ModInstaller) Tidy() (InstalledModMap, error) {
	// install first if necessary
	if len(i.installData.Lock) == 0 {
		fmt.Println("no workspace lock found\nrun 'steampipe mod install'")
		return nil, nil
	}
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
	return unusedMods, nil
}
