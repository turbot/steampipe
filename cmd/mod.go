package cmd

import (
	"fmt"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

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
	cmd.AddCommand(modTidyCmd())

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

	msg, err := mod_installer.InstallWorkspaceDependencies(&mod_installer.InstallOpts{ShouldUpdate: false})
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
		utils.LogTime("cmd.runModGetCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	modsArgs := append([]string{}, args...)
	// first convert the mod args into well formed mod names
	requiredModVersions, err := getRequiredModVersionsFromArgs(modsArgs)
	utils.FailOnError(err)
	if len(requiredModVersions) == 0 {
		fmt.Println("No mods to add")
		return
	}
	// just call install, passing an AdditionalMods option
	installData, err := mod_installer.InstallWorkspaceDependencies(&mod_installer.InstallOpts{
		GetMods: requiredModVersions,
	})
	utils.FailOnError(err)

	fmt.Printf(mod_installer.BuildInstallSummary(installData))
}

func getRequiredModVersionsFromArgs(modsArgs []string) ([]*modconfig.ModVersionConstraint, error) {
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

	//workspacePath := viper.GetString(constants.ArgWorkspaceChDir)
	//installer, err := mod_installer.NewModInstaller(&mod_installer.InstallOpts{WorkspacePath: workspacePath})
	//utils.FailOnError(err)

	//installedMods := installer.GetModList()
	//utils.FailOnError(err)
	//// TODO FORMAT LIST
	//fmt.Println(installedMods)
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

func runModUpdateCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runModUpdateCmd")
	defer func() {
		utils.LogTime("cmd.runModUpdateCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	modsArgs := append([]string{}, args...)
	// first convert the mod args into well formed mod names
	updateMods, err := getRequiredModVersionsFromArgs(modsArgs)
	utils.FailOnError(err)

	installData, err := mod_installer.InstallWorkspaceDependencies(&mod_installer.InstallOpts{ShouldUpdate: true, UpdateMods: updateMods})
	utils.FailOnError(err)
	fmt.Printf(mod_installer.BuildUpdateSummary(installData))
}

// tidy
func modTidyCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "tidy",
		Run:   runModTidyCmd,
		Short: "Tidy mod dependencies",
		Long: `Tidy mod dependencies.
`,
	}

	cmdconfig.OnCmd(cmd)
	return cmd
}

func runModTidyCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runModTidyCmd")
	defer func() {
		utils.LogTime("cmd.runModTidyCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	opts := &mod_installer.InstallOpts{WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir)}

	// install workspace dependencies
	installer, err := mod_installer.NewModInstaller(opts)
	utils.FailOnError(err)

	unusedMods, err := installer.Tidy()
	utils.FailOnError(err)

	if count := len(unusedMods.FlatMap()); count > 0 {
		fmt.Printf("Removed %d unused %s\n", count, utils.Pluralize("mod", count))
	}
}
