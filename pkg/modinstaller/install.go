package modinstaller

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/utils"
)

func InstallWorkspaceDependencies(opts *InstallOpts) (_ *InstallData, err error) {
	utils.LogTime("cmd.InstallWorkspaceDependencies")
	defer func() {
		utils.LogTime("cmd.InstallWorkspaceDependencies end")
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

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
