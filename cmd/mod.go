package cmd

import (
	"fmt"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/mod_installer"
	"github.com/turbot/steampipe/utils"
)

// mod management commands
func modCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "mod [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe mod management",
		Long:  `Steampipe mod management.`,
	}

	cmd.AddCommand(modInstallCmd())
	cmd.AddCommand(modGetCmd())
	cmd.AddCommand(modListCmd())
	cmd.AddCommand(modUpdateCmd())
	cmd.AddCommand(modGetCmd())

	return cmd
}

// install
func modInstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "install",
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

	msg, err := mod_installer.InstallModDependencies(false)
	utils.FailOnError(err)
	fmt.Println(msg)
}

// get
func modGetCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get [git-provider/org/]name[@version]",
		Args:  cobra.ArbitraryArgs,
		Run:   runModGetCmd,
		Short: "Add mod dependencies",
		Long: `Add mod dependencies.
`,
	}

	cmdconfig.OnCmd(cmd)
	return cmd
}

func runModGetCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runModGetCmd")
	defer func() {
		// TODO revert to old mod file in case of error
		utils.LogTime("cmd.runModGetCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	modsArgs := append([]string{}, args...)
	// first convert the mod args into well formed mod names
	modVersions, err := getModVersions(modsArgs)
	utils.FailOnError(err)

	// load workspace mod
	workspacePath := viper.GetString(constants.ArgWorkspaceChDir)
	// TODO handle no mod file
	mod, err := parse.ParseModDefinition(workspacePath)
	utils.FailOnError(err)

	// add dependencies to mod
	mod.AddModDependencies(modVersions)
	err = mod.Save()
	utils.FailOnError(err)

	msg, err := mod_installer.InstallModDependencies(false)
	utils.FailOnError(err)
	fmt.Println(msg)

}

func getModVersions(modsArgs []string) ([]*modconfig.ModVersionConstraint, error) {
	var errors []error
	mods := make([]*modconfig.ModVersionConstraint, len(modsArgs))
	for i, modArg := range modsArgs {
		// create mod version from arg
		modVersion, err := modconfig.NewModVersionConstraint(modArg)
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

// list
func modListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Run:   runModListCmd,
		Short: "Install mod dependencies",
		Long: `Install mod dependencies.
`,
	}

	cmdconfig.OnCmd(cmd)
	return cmd
}

func runModListCmd(*cobra.Command, []string) {
	utils.LogTime("cmd.runModListCmd")
	defer func() {
		utils.LogTime("cmd.runModListCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	workspacePath := viper.GetString(constants.ArgWorkspaceChDir)
	installer, err := mod_installer.NewModInstaller(workspacePath)
	utils.FailOnError(err)

	installedMods := installer.InstalledModVersions
	utils.FailOnError(err)
	// TODO FORMAT LIST
	fmt.Println(installedMods)
}

// update
func modUpdateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update",
		Run:   runModUpdateCmd,
		Short: "Install mod dependencies",
		Long: `Install mod dependencies.
`,
	}

	cmdconfig.OnCmd(cmd)
	return cmd
}

func runModUpdateCmd(*cobra.Command, []string) {
	utils.LogTime("cmd.runModUpdateCmd")
	defer func() {
		utils.LogTime("cmd.runModUpdateCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	msg, err := mod_installer.InstallModDependencies(true)
	utils.FailOnError(err)
	fmt.Println(msg)
}
