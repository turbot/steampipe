package cmd

import (
	"fmt"

	"github.com/turbot/steampipe/steampipeconfig/version_map"

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
	cmd.AddCommand(modUninstallCmd())
	cmd.AddCommand(modUpdateCmd())
	cmd.AddCommand(modListCmd())

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

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgPrune, "", true, "Remove unreferenced mods after installation")
	return cmd
}

func runModInstallCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
			exitCode = 1
		}
	}()

	// if any mod names were passed as args, convert into formed mod names
	modArgs, err := getRequiredModVersionsFromArgs(args)
	// TODO validate only 1 version of each mod
	utils.FailOnError(err)

	// if any mod args are specied, set the update flag
	// - if a mod is specified as an arg which is already installed with a different constraint,
	// it should be installed with th elkatest available version
	// (if constraint is the same nothing should be done unless update flag is set)
	/*

		latest m1@1.5

		1:
		current dep: m1@*
		installed: m1@1.1
		args: m1@* --update
		result: m1@1.5

		2:
		current dep: m1@1.1
		installed: m1@1.1
		args: m1@1.1 --update
		result: m1@1.1

		3:
		current dep: m1@1.1
		installed: m1@1.1
		args: m1@1.2
		result: m1@1.2

		4:
		current dep: m1@1.1
		installed: m1@1.1
		args: m1@1.*
		result: m1@1.5

		5:
		current dep: m1@1.*
		installed: m1@1.1
		args: m1@1.*
		result: m1@1.1 [NO UPDATE]

		6:
		current dep: m1@1.*
		installed: m1@1.1
		args: m1@1.0
		result: m1@1.0



	*/

	opts := &mod_installer.InstallOpts{
		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
		ModArgs:       modArgs,
	}

	installData, err := mod_installer.InstallWorkspaceDependencies(opts)
	utils.FailOnError(err)

	fmt.Printf(mod_installer.BuildInstallSummary(installData))
}

// uninstall
func modUninstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "uninstall",
		Run:   runModUninstallCmd,
		Short: "Uninstall mod dependencies",
		Long: `Uninstall mod dependencies.
`,
	}

	cmdconfig.OnCmd(cmd)
	return cmd
}

func runModUninstallCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
			exitCode = 1
		}
	}()

	// if any mod names were passed as args, convert into formed mod names
	modArgs, err := getRequiredModVersionsFromArgs(args)
	// TODO validate only 1 version of each mod
	utils.FailOnError(err)

	opts := &mod_installer.InstallOpts{
		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
		ModArgs:       modArgs,
	}

	installData, err := mod_installer.UninstallWorkspaceDependencies(opts)
	utils.FailOnError(err)

	fmt.Printf(mod_installer.BuildUninstallSummary(installData))
}

// update
func modUpdateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update",
		Run:   runModUpdateCmd,
		Short: "Update workspace dependencies",
		Long: `Update workspace dependencies.
`,
	}

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgPrune, "", true, "Remove unreferenced mods after installation")

	return cmd
}

func runModUpdateCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runModUpdateCmd")
	defer func() {
		utils.LogTime("cmd.runModUpdateCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
			exitCode = 1
		}
	}()

	// if any mod names were passed as args, convert into formed mod names
	modArgs, err := getRequiredModVersionsFromArgs(args)
	// TODO validate only 1 version of each mod
	utils.FailOnError(err)

	opts := &mod_installer.InstallOpts{
		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
		ModArgs:       modArgs,
		Updating:      true,
	}

	installData, err := mod_installer.InstallWorkspaceDependencies(opts)
	utils.FailOnError(err)

	fmt.Printf(mod_installer.BuildInstallSummary(installData))
}

func getRequiredModVersionsFromArgs(modsArgs []string) (version_map.VersionConstraintMap, error) {
	var errors []error
	mods := make(version_map.VersionConstraintMap, len(modsArgs))
	for _, modArg := range modsArgs {
		// create mod version from arg
		modVersion, err := modconfig.NewModVersionConstraint(modArg)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		// TODO include alias in key
		mods[modVersion.Name] = modVersion
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
			exitCode = 1
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

//func doUpdates(updateMods version_map.VersionConstraintMap) {
//	opts := &mod_installer.InstallOpts{
//		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
//		Updating:      true,
//		UpdateMods:    updateMods,
//	}
//
//	installData, err := mod_installer.InstallWorkspaceDependencies(opts)
//	utils.FailOnError(err)
//	fmt.Printf(mod_installer.BuildUpdateSummary(installData))
//}
//
//
//// tidy
//func modTidyCmd() *cobra.Command {
//	var cmd = &cobra.Command{
//		Use:   "tidy",
//		Run:   runModTidyCmd,
//		Short: "Tidy mod dependencies",
//		Long: `Tidy mod dependencies.
//`,
//	}
//
//	cmdconfig.OnCmd(cmd)
//	return cmd
//}
//
//func runModTidyCmd(cmd *cobra.Command, args []string) {
//	utils.LogTime("cmd.runModTidyCmd")
//	defer func() {
//		utils.LogTime("cmd.runModTidyCmd end")
//		if r := recover(); r != nil {
//			utils.ShowError(helpers.ToError(r))
//			exitCode = 1
//		}
//	}()
//
//	opts := &mod_installer.InstallOpts{WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir)}
//
//	// install workspace dependencies
//	installer, err := mod_installer.NewModInstaller(opts)
//	utils.FailOnError(err)
//
//	unusedMods, err := installer.Tidy()
//	utils.FailOnError(err)
//
//	if count := len(unusedMods.FlatMap()); count > 0 {
//		fmt.Printf("Removed %d unused %s\n", count, utils.Pluralize("mod", count))
//	}
//}

//func showUpdates() {
//	opts := &mod_installer.InstallOpts{
//		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
//	}
//	current, updates, err := mod_installer.GetAvailableUpdates(opts)
//	utils.FailOnError(err)
//	fmt.Println(mod_installer.BuildAvailableUpdateSummary(current, updates))
//}
