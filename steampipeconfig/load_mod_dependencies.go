package steampipeconfig

import (
	"fmt"
	"os"
	"path"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// if deep is false only load single level of dependencies - if true load full tree (tbd if this is needed)
func LoadModDependencies(m *modconfig.Mod, modsFolder string, modMap modconfig.ModMap, deep bool) error {
	for _, dep := range m.ModDepends {
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
		opts := &LoadModOptions{
			ListOptions: &filehelpers.ListOptions{
				// recursively load mod files
				Flags: filehelpers.FilesRecursive,
			},
		}
		mod, err := LoadMod(modPath, opts)
		if err != nil {
			return err
		}

		modMap[dependencyName] = mod

		if deep {
			if err := LoadModDependencies(mod, modsFolder, modMap, deep); err != nil {
				return err
			}
		}
	}

	return nil
}
