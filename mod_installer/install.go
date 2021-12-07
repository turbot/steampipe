package mod_installer

import (
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

func InstallModDependencies(opts *InstallOpts) (string, error) {
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	opts.WorkspacePath = viper.GetString(constants.ArgWorkspaceChDir)

	// install workspace dependencies
	installer, err := NewModInstaller(opts)
	if err != nil {
		return "", err
	}

	if err := installer.InstallWorkspaceDependencies(); err != nil {
		return "", err
	}

	return installer.InstallReport(), nil
}
