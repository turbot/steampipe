package modinstaller

import (
	"context"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/utils"
)

func UninstallWorkspaceDependencies(ctx context.Context, opts *InstallOpts) (*InstallData, error) {
	utils.LogTime("cmd.UninstallWorkspaceDependencies")
	defer func() {
		utils.LogTime("cmd.UninstallWorkspaceDependencies end")
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
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
