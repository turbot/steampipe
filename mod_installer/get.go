package mod_installer

import (
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

func GetMods(modsArgs []string) (string, error) {
	// first convert the mod args into well formed mod names
	modVersions, err := getModVersions(modsArgs)
	if err != nil {
		return "", err
	}
	installer, err := NewModInstaller(viper.GetString(constants.ArgWorkspaceChDir))
	if err != nil {
		return "", err
	}

	err = installer.installModDependenciesRecursively(modVersions)
	if err != nil {
		return "", err
	}
	// update mod file

	return "", nil
}

func getModVersions(modsArgs []string) ([]*modconfig.ModVersion, error) {
	var errors []error
	mods := make([]*modconfig.ModVersion, len(modsArgs))
	for i, modArg := range modsArgs {
		// create mod version from arg
		modVersion, err := modconfig.NewModVersion(modArg)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		mods[i] = modVersion
	}
	if len(errors) > 0 {
		return nil, utils.CombineErrors(errors...)
	}
	return mods, nil
}
