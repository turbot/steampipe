package mod

import (
	"fmt"
	"os"
	"path"

	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// LoadModDependencies :: load all dependencies of given mod
// used to load all mods in a workspace, using the workspace manifest mod

func LoadModDependencies(parentMod *modconfig.Mod, modsFolder string) (ModMap, error) {
	var res = ModMap{}
	if err := loadModDependencies(parentMod, modsFolder, res, false); err != nil {
		return nil, err
	}
	return res, nil
}

// if deep is false only load single level of dependencies - if true load full tree (tbd if this is needed)
func loadModDependencies(parentMod *modconfig.Mod, modsFolder string, modMap ModMap, deep bool) error {

	for _, dep := range parentMod.ModDepends {

		dependencyName := dep.FullName()
		// have we already loaded this dependency?
		if _, ok := modMap[dependencyName]; ok {
			continue
		}
		// convert mod Name into a path
		modPath := path.Join(modsFolder, dependencyName)
		if _, err := os.Stat(modPath); os.IsNotExist(err) {
			return fmt.Errorf("mod %s not found in mod folder %s", dependencyName, modsFolder)
		}

		// now try to parse the mod
		mod, err := steampipeconfig.LoadMod(modPath)
		if err != nil {
			return err
		}

		modMap[dependencyName] = mod

		if deep {
			if err := loadModDependencies(mod, modsFolder, modMap, deep); err != nil {
				return err
			}
		}
	}

	return nil
}
