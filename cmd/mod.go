package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/modinstaller"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/utils"
)

// mod management commands
func modCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "mod [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe mod management",
		Long: `Steampipe mod management.

Mods enable you to run, build, and share dashboards, benchmarks and other resources.

Find pre-built mods in the public registry at https://hub.steampipe.io.

Examples:

    # Create a new mod in the current directory
    steampipe mod init

    # Install a mod
    steampipe mod install github.com/turbot/steampipe-mod-aws-compliance
    
    # Update a mod
    steampipe mod update github.com/turbot/steampipe-mod-aws-compliance
    
    # List installed mods
    steampipe mod list
    
    # Uninstall a mod
    steampipe mod uninstall github.com/turbot/steampipe-mod-aws-compliance
	`,
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
		AddBoolFlag(constants.ArgPrune, true, "Remove unused dependencies after installation is complete").
		AddBoolFlag(constants.ArgDryRun, false, "Show which mods would be installed/updated/uninstalled without modifying them").
		AddBoolFlag(constants.ArgHelp, false, "Help for install", cmdconfig.FlagOptions.WithShortHand("h"))

	return cmd
}

func runModInstallCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	// if any mod names were passed as args, convert into formed mod names
	opts := newInstallOpts(cmd, args...)
	installData, err := modinstaller.InstallWorkspaceDependencies(opts)
	error_helpers.FailOnError(err)

	fmt.Println(modinstaller.BuildInstallSummary(installData))
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
		AddBoolFlag(constants.ArgPrune, true, "Remove unused dependencies after uninstallation is complete").
		AddBoolFlag(constants.ArgDryRun, false, "Show which mods would be uninstalled without modifying them").
		AddBoolFlag(constants.ArgHelp, false, "Help for uninstall", cmdconfig.FlagOptions.WithShortHand("h"))

	return cmd
}

func runModUninstallCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	opts := newInstallOpts(cmd, args...)
	installData, err := modinstaller.UninstallWorkspaceDependencies(ctx, opts)
	error_helpers.FailOnError(err)

	fmt.Println(modinstaller.BuildUninstallSummary(installData))
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
		AddBoolFlag(constants.ArgPrune, true, "Remove unused dependencies after update is complete").
		AddBoolFlag(constants.ArgDryRun, false, "Show which mods would be updated without modifying them").
		AddBoolFlag(constants.ArgHelp, false, "Help for update", cmdconfig.FlagOptions.WithShortHand("h"))

	return cmd
}

func runModUpdateCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runModUpdateCmd")
	defer func() {
		utils.LogTime("cmd.runModUpdateCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	opts := newInstallOpts(cmd, args...)

	installData, err := modinstaller.InstallWorkspaceDependencies(opts)
	error_helpers.FailOnError(err)

	fmt.Println(modinstaller.BuildInstallSummary(installData))
}

// list
func modListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Run:   runModListCmd,
		Short: "List currently installed mods",
		Long:  `List currently installed mods.`,
	}

	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, false, "Help for list", cmdconfig.FlagOptions.WithShortHand("h"))
	return cmd
}

func runModListCmd(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runModListCmd")
	defer func() {
		utils.LogTime("cmd.runModListCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()
	opts := newInstallOpts(cmd)
	installer, err := modinstaller.NewModInstaller(opts)
	error_helpers.FailOnError(err)

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

	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, false, "Help for init", cmdconfig.FlagOptions.WithShortHand("h"))
	return cmd
}

func runModInitCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runModInitCmd")
	defer func() {
		utils.LogTime("cmd.runModInitCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()
	workspacePath := viper.GetString(constants.ArgModLocation)
	if parse.ModfileExists(workspacePath) {
		fmt.Println("Working folder already contains a mod definition file")
		return
	}
	mod := modconfig.CreateDefaultMod(workspacePath)
	err := mod.Save()
	error_helpers.FailOnError(err)
	fmt.Printf("Created mod definition file '%s'\n", filepaths.ModFilePath(workspacePath))
}

// helpers

func newInstallOpts(cmd *cobra.Command, args ...string) *modinstaller.InstallOpts {
	opts := &modinstaller.InstallOpts{
		WorkspacePath: viper.GetString(constants.ArgModLocation),
		DryRun:        viper.GetBool(constants.ArgDryRun),
		ModArgs:       args,
		Command:       cmd.Name(),
	}
	return opts
}
