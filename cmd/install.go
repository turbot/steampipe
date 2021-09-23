package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/mod_installer"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/utils"
)

func installCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "install",
		TraverseChildren: true,
		Run:              runInstallCmd,
		Short:            "Install dependencies for the current workspace",
		Long:             "Install dependencies for the current workspace",
	}

	// Notes:
	// * In the future we may add --csv and --json flags as shortcuts for --output
	cmdconfig.OnCmd(cmd)
	return cmd
}

func runInstallCmd(*cobra.Command, []string) {
	utils.LogTime("cmd.runQueryCmd start")

	defer func() {
		utils.LogTime("cmd.runQueryCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	workspacePath := viper.GetString(constants.ArgWorkspace)

	// install workspace dependencies
	// TODO do we need to care about variables?? probably?

	// load the modfile only
	mod, err := parse.ParseModDefinition(workspacePath)
	utils.FailOnError(err)
	installer := mod_installer.NewModInstaller(workspacePath)
	err = installer.InstallModDependencies(mod)
	utils.FailOnError(err)
}
