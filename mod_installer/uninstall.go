package mod_installer

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/utils"
)

func UninstallWorkspaceDependencies(opts *InstallOpts) (*InstallData, error) {
	utils.LogTime("cmd.UninstallWorkspaceDependencies")
	defer func() {
		utils.LogTime("cmd.UninstallWorkspaceDependencies end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	// uninstall workspace dependencies
	installer, err := NewModInstaller(opts)
	if err != nil {
		return nil, err
	}

	if err := installer.UninstallWorkspaceDependencies(); err != nil {
		return nil, err
	}

	return installer.installData, nil

}
