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
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
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
	cmd.AddCommand(modInitCmd())
	cmd.Flags().BoolP(constants.ArgHelp, "h", false, "Help for mod")

	return cmd
}

// install
func modInstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "install",
		Run:   runModInstallCmd,
		Short: "Install one or more mods and their dependencies",
		Long:  `Install one or more mods and their dependencies.`,
	}

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgPrune, "", true, "Remove unused dependencies after installation is complete").
		AddBoolFlag(constants.ArgDryRun, "", false, "Show which mods would be installed/updated/uninstalled without modifying them").
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for install")

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
	opts := newInstallOpts(cmd, args...)
	installData, err := mod_installer.InstallWorkspaceDependencies(opts)
	utils.FailOnError(err)

	fmt.Println(mod_installer.BuildInstallSummary(installData))
}

// uninstall
func modUninstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "uninstall",
		Run:   runModUninstallCmd,
		Short: "Uninstall a mod and its dependencies",
		Long:  `Uninstall a mod and its dependencies.`,
	}

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgPrune, "", true, "Remove unused dependencies after uninstallation is complete").
		AddBoolFlag(constants.ArgDryRun, "", false, "Show which mods would be uninstalled without modifying them").
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for uninstall")

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
		Short: "Update one or more mods and their dependencies",
		Long:  `Update one or more mods and their dependencies.`,
	}

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgPrune, "", true, "Remove unused dependencies after update is complete").
		AddBoolFlag(constants.ArgDryRun, "", false, "Show which mods would be updated without modifying them").
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for update")

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

	fmt.Println(mod_installer.BuildInstallSummary(installData))
}

// list
func modListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Run:   runModListCmd,
		Short: "List currently installed mods",
		Long:  `List currently installed mods.`,
	}

	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, "h", false, "Help for list")
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

// init
func modInitCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init",
		Run:   runModInitCmd,
		Short: "Initialize the current directory with a mod.sp file",
		Long:  `Initialize the current directory with a mod.sp file.`,
	}

	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, "h", false, "Help for init")
	return cmd
}

func runModInitCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runModInitCmd")
	defer func() {
		utils.LogTime("cmd.runModInitCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
			exitCode = 1
		}
	}()
	workspacePath := viper.GetString(constants.ArgWorkspaceChDir)
	if parse.ModfileExists(workspacePath) {
		fmt.Println("Working folder already contains a mod definition file")
		return
	}
	mod := modconfig.CreateDefaultMod(workspacePath)
	utils.FailOnError(mod.Save())
	fmt.Printf("Created mod definition file '%s'\n", constants.ModFilePath(workspacePath))
}

// helpers

func newInstallOpts(cmd *cobra.Command, args ...string) *mod_installer.InstallOpts {
	opts := &mod_installer.InstallOpts{
		WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir),
		DryRun:        viper.GetBool(constants.ArgDryRun),
		ModArgs:       args,
		Command:       cmd.Name(),
	}
	return opts
}
