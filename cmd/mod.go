package cmd

import (
	"fmt"
	"strings"

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
	cmd.AddCommand(modPruneCmd())
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
		AddBoolFlag(constants.ArgPrune, "", true, "Remove unreferenced mods after installation").
		AddBoolFlag(constants.ArgDryRun, "", false, "Show which mods would be installed or uninstalled without performing the installation")
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

	opts := &mod_installer.InstallOpts{
		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
		DryRun:        viper.GetBool(constants.ArgDryRun),
		ModArgs:       args,
	}

	installData, err := mod_installer.InstallWorkspaceDependencies(opts)
	utils.FailOnError(err)

	fmt.Println(mod_installer.BuildInstallSummary(installData))
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

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgPrune, "", true, "Remove unreferenced mods after uninstallation").
		AddBoolFlag(constants.ArgDryRun, "", false, "Show which mods would be uninstalled without removing them")

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

	opts := newInstallOpts(cmd, args...)
	installData, err := mod_installer.UninstallWorkspaceDependencies(opts)
	utils.FailOnError(err)

	fmt.Println(mod_installer.BuildUninstallSummary(installData))
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
		AddBoolFlag(constants.ArgPrune, "", true, "Remove unreferenced mods after installation").
		AddBoolFlag(constants.ArgDryRun, "", false, "Show which mods would be updated without updating them")

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

	opts := newInstallOpts(cmd, args...)

	installData, err := mod_installer.InstallWorkspaceDependencies(opts)
	utils.FailOnError(err)

	fmt.Println(mod_installer.BuildUpdateSummary(installData))
}

// list
func modListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Run:   runModListCmd,
		Short: "List mod dependencies",
		Long: `List mod dependencies.
`,
	}

	cmdconfig.OnCmd(cmd)
	return cmd
}

func runModListCmd(cmd *cobra.Command, _ []string) {
	utils.LogTime("cmd.runModListCmd")
	defer func() {
		utils.LogTime("cmd.runModListCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
			exitCode = 1
		}
	}()
	opts := newInstallOpts(cmd)
	installer, err := mod_installer.NewModInstaller(opts)
	utils.FailOnError(err)

	treeString := installer.GetModList()
	if len(strings.Split(treeString, "\n")) > 1 {
		fmt.Println()
	}
	fmt.Println(treeString)
}

// prune
func modPruneCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "prune",
		Run:   runModPruneCmd,
		Short: "Prune mod dependencies",
		Long: `Prune mod dependencies.
`,
	}

	cmdconfig.OnCmd(cmd)
	return cmd
}

func runModPruneCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runModPruneCmd")
	defer func() {
		utils.LogTime("cmd.runModPruneCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
			exitCode = 1
		}
	}()

	opts := &mod_installer.InstallOpts{
		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
		DryRun:        viper.GetBool(constants.ArgDryRun),
	}

	// install workspace dependencies
	installer, err := mod_installer.NewModInstaller(opts)
	utils.FailOnError(err)

	unusedMods, err := installer.Prune()
	utils.FailOnError(err)

	if count := len(unusedMods.FlatMap()); count > 0 {
		fmt.Println(mod_installer.BuildPruneSummary(unusedMods))
	}
}

func newInstallOpts(cmd *cobra.Command, args ...string) *mod_installer.InstallOpts {
	opts := &mod_installer.InstallOpts{
		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
		DryRun:        viper.GetBool(constants.ArgDryRun),
		ModArgs:       args,
		Command:       cmd.Name(),
	}
	return opts
}
