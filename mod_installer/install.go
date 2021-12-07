package mod_installer

import (
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/utils"
)

type InstallOpts struct {
	ShouldUpdate bool
	GetMods      []*modconfig.ModVersionConstraint
	UpdateMods   []*modconfig.ModVersionConstraint
}

func InstallModDependencies(opts *InstallOpts) (string, error) {
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	workspacePath := viper.GetString(constants.ArgWorkspaceChDir)

	// install workspace dependencies
	var mod *modconfig.Mod
	if !parse.ModfileExists(workspacePath) {
		if len(opts.GetMods) == 0 {
			return "No mod file found, so there are no dependencies to install", nil
		}
		// so there is no mod file, but we are adding mod dependencies - create a default mod
		mod = modconfig.CreateDefaultMod(workspacePath)
	} else {
		// load the modfile only
		var err error
		mod, err = parse.ParseModDefinition(workspacePath)
		if err != nil {
			return "", err
		}
	}

	installer, err := NewModInstaller(workspacePath)
	if err != nil {
		return "", err
	}

	// set update flag
	installer.ShouldUpdate = opts.ShouldUpdate
	installer.GetMods = opts.GetMods
	installer.UpdateMods = opts.UpdateMods

	err = installer.InstallModDependencies(mod)
	if err != nil {
		return "", err
	}

	return installer.InstallReport(), nil
}
