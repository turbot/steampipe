package mod_installer

import (
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

func InstallWorkspaceDependencies(opts *InstallOpts) (*InstallData, error) {
	utils.LogTime("cmd.InstallModDependencies")
	defer func() {
		utils.LogTime("cmd.InstallModDependencies end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	opts.WorkspacePath = viper.GetString(constants.ArgWorkspaceChDir)

	// install workspace dependencies
	installer, err := NewModInstaller(opts)
	if err != nil {
		return nil, err
	}

	if err := installer.InstallWorkspaceDependencies(); err != nil {
		return nil, err
	}

	return installer.installData, nil
}
