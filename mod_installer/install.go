package mod_installer

import (
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/utils"
)

func InstallModDependencies(shouldUpdate bool) (string, error) {
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	workspacePath := viper.GetString(constants.ArgWorkspaceChDir)

	// install workspace dependencies
	// TODO do we need to care about variables?? probably?

	if !parse.ModfileExists(workspacePath) {
		return "No mod file found, so there are no dependencies to install", nil
	}
	// load the modfile only
	mod, err := parse.ParseModDefinition(workspacePath)
	utils.FailOnError(err)

	installer, err := NewModInstaller(workspacePath)
	if err != nil {
		return "", err
	}

	// set update flag
	installer.ShouldUpdate = shouldUpdate

	err = installer.InstallModDependencies(mod)
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return installer.InstallReport(), nil
}
