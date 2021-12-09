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
	cmd.AddCommand(modListCmd())
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

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgTidy, "", true, "Remove unreferenced mods after installation").
		AddBoolFlag(constants.ArgUpdate, "", true, "Update all dependent mods to the latest available version").
		AddBoolFlag(constants.ArgShowUpdates, "", true, "Update all dependent mods to the latest available version")
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

	if viper.GetBool(constants.ArgShowUpdates) {
		showUpdates()
		return
	}

	// if any mod names were passed as args, convert into formed mod names
	modArgs, err := getRequiredModVersionsFromArgs(args)
	// TODO validate only 1 version of each mod
	utils.FailOnError(err)

	opts := &mod_installer.InstallOpts{
		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
		// TODO handle show
		Updating: viper.GetBool(constants.ArgUpdate),
		ModArgs:  modArgs,
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

func showUpdates() {
	opts := &mod_installer.InstallOpts{
		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
	}
	current, updates, err := mod_installer.GetAvailableUpdates(opts)
	utils.FailOnError(err)
	fmt.Println(mod_installer.BuildAvailableUpdateSummary(current, updates))
}

//func runModGetCmd(cmd *cobra.Command, args []string) {
//	utils.LogTime("cmd.runModGetCmd")
//	defer func() {
//		utils.LogTime("cmd.runModGetCmd end")
//		if r := recover(); r != nil {
//			utils.ShowError(helpers.ToError(r))
//			exitCode = 1
//		}
//	}()
//
//	modsArgs := append([]string{}, args...)
//	// first convert the mod args into well formed mod names
//	requiredModVersions, err := getRequiredModVersionsFromArgs(modsArgs)
//	// TODO validate only 1 version of each mod
//	utils.FailOnError(err)
//	if len(requiredModVersions) == 0 {
//		fmt.Println("No mods to add")
//		return
//	}
//	// just call install, passing GetMods option
//	opts := &mod_installer.InstallOpts{
//		AddMods:       requiredModVersions,
//		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
//	}
//	installData, err := mod_installer.InstallWorkspaceDependencies(opts)
//	utils.FailOnError(err)
//
//	getSummary := mod_installer.BuildGetSummary(installData, requiredModVersions)
//	fmt.Printf(getSummary)
//}

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

// update
//func modUpdateCmd() *cobra.Command {
//	var cmd = &cobra.Command{
//		Use:   "update",
//		Run:   runModUpdateCmd,
//		Short: "Update workspace dependencies",
//		Long: `Update workspace dependencies.
//`,
//	}
//
//	cmdconfig.OnCmd(cmd).
//		AddBoolFlag(constants.ArgAll, "", false, "Update all mods to its latest available version").
//		AddBoolFlag(constants.ArgShow, "", false, "Just display the updates which would be performed")
//
//	return cmd
//}

//func runModUpdateCmd(cmd *cobra.Command, args []string) {
//	utils.LogTime("cmd.runModUpdateCmd")
//	defer func() {
//		utils.LogTime("cmd.runModUpdateCmd end")
//		if r := recover(); r != nil {
//			utils.ShowError(helpers.ToError(r))
//			exitCode = 1
//		}
//	}()
//
//	// args to 'mod update' -- one or more mods to update
//	mods, err := validateArgs(args)
//	if err != nil {
//		fmt.Println()
//		utils.ShowError(err)
//		fmt.Println()
//		cmd.Help()
//		fmt.Println()
//		exitCode = 2
//		return
//	}
//
//	// first convert the mod args into well-formed mod names
//	updateMods, err := getRequiredModVersionsFromArgs(mods)
//	utils.FailOnError(err)
//
//	// if show flag is set, just display the potential updates
//	if cmdconfig.Viper().GetBool(constants.ArgShow) {
//		showUpdates()
//	} else {
//		doUpdates(updateMods)
//	}
//}
//

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
			exitCode = 1
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
