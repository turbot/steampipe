package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"

"github.com/turbot/pipe-fittings/error_helpers"
"github.com/turbot/pipe-fittings/statushooks"
"github.com/turbot/pipe-fittings/utils"
"github.com/turbot/pipe-fittings/version"
"github.com/turbot/steampipe/pkg/filepaths_steampipe"
)

var exitCode int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "steampipe [--version] [--help] COMMAND [args]",
	Version: version.SteampipeVersion.String(),
	Short:   "Query cloud resources using SQL",
	Long: `Query cloud resources using SQL.

The available commands for execution are listed below.
The most common, useful commands are shown first, followed by
less common or more advanced commands. If you're just getting
started with Steampipe, stick with the common commands. For the
other commands, please read the help and docs before usage.

Getting started:

  # Interactive SQL query console
  steampipe query

  # Execute a defined SQL query
  steampipe query "select * from aws_s3_bucket"

  # Install a plugin
  steampipe plugin install azure

  # Get help for a command
  steampipe help query

  Documentation available at https://steampipe.io/docs
 `,
}

func InitCmd() {
	utils.LogTime("cmd.root.InitCmd start")
	defer utils.LogTime("cmd.root.InitCmd end")

	rootCmd.SetVersionTemplate(fmt.Sprintf("Steampipe v%s\n", version.SteampipeVersion.String()))

	// global flags
	rootCmd.PersistentFlags().String(constants.ArgWorkspaceProfile, "default", "The workspace profile to use") // workspace profile profile is a global flag since install-dir(global) can be set through the workspace profile
	rootCmd.PersistentFlags().String(constants.ArgInstallDir, filepaths_steampipe.DefaultInstallDir, "Path to the Config Directory")
	rootCmd.PersistentFlags().Bool(constants.ArgSchemaComments, true, "Include schema comments when importing connection schemas")

	error_helpers.FailOnError(viper.BindPFlag(constants.ArgInstallDir, rootCmd.PersistentFlags().Lookup(constants.ArgInstallDir)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgWorkspaceProfile, rootCmd.PersistentFlags().Lookup(constants.ArgWorkspaceProfile)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgSchemaComments, rootCmd.PersistentFlags().Lookup(constants.ArgSchemaComments)))

	AddCommands()

	// disable auto completion generation, since we don't want to support
	// powershell yet - and there's no way to disable powershell in the default generator
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().BoolP(constants.ArgHelp, "h", false, "Help for steampipe")
	rootCmd.Flags().BoolP(constants.ArgVersion, "v", false, "Version for steampipe")

	hideRootFlags(constants.ArgSchemaComments)

	// tell OS to reclaim memory immediately
	os.Setenv("GODEBUG", "madvdontneed=1")

}

func hideRootFlags(flags ...string) {
	for _, flag := range flags {
		rootCmd.Flag(flag).Hidden = true
	}
}

func AddCommands() {
	// explicitly initialise commands here rather than in init functions to allow us to handle errors from the config load
	rootCmd.AddCommand(
		pluginCmd(),
		queryCmd(),
		checkCmd(),
		serviceCmd(),
		modCmd(),
		generateCompletionScriptsCmd(),
		pluginManagerCmd(),
		dashboardCmd(),
		variableCmd(),
		loginCmd(),
	)
}

func Execute() int {
	utils.LogTime("cmd.root.Execute start")
	defer utils.LogTime("cmd.root.Execute end")

	ctx := createRootContext()

	rootCmd.ExecuteContext(ctx)
	return exitCode
}

// create the root context - add a status renderer
func createRootContext() context.Context {
	statusRenderer := statushooks.NullHooks
	// if the client is a TTY, inject a status spinner
	if isatty.IsTerminal(os.Stdout.Fd()) {
		statusRenderer = statushooks.NewStatusSpinnerHook()
	}

	ctx := statushooks.AddStatusHooksToContext(context.Background(), statusRenderer)
	return ctx
}
