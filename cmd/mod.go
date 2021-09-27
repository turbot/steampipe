package cmd

import (
	"fmt"

	"github.com/turbot/steampipe/cmdconfig"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/mod_installer"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/utils"
)

// modCmd :: mod management commands
func modCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "mod [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe mod management",
		Long:  `Steampipe mod management.`,
	}

	cmd.AddCommand(modInstallCmd())
	//cmd.AddCommand(modListCmd())
	//cmd.AddCommand(modUninstallCmd())
	//cmd.AddCommand(modUpdateCmd())

	return cmd
}

// modInstallCmd :: Install a mod
func modInstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "install [flags] [registry/org/]name[@version]",
		Args:  cobra.ArbitraryArgs,
		Run:   runModInstallCmd,
		Short: "Install mod dependencies",
		Long: `Install mod dependencies.
`,
	}

	cmdconfig.OnCmd(cmd)
	return cmd
}

func runModInstallCmd(*cobra.Command, []string) {
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	workspacePath := viper.GetString(constants.ArgWorkspace)

	// install workspace dependencies
	// TODO do we need to care about variables?? probably?

	if !parse.ModfileExists(workspacePath) {
		fmt.Println("No mod file found, so there are no dependencies to install")
		return
	}
	// load the modfile only
	mod, err := parse.ParseModDefinition(workspacePath)
	utils.FailOnError(err)

	installer := mod_installer.NewModInstaller(workspacePath)
	err = installer.InstallModDependencies(mod)
	fmt.Println(installer.InstallReport())
	utils.FailOnError(err)
}
