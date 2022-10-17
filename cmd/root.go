package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe-plugin-sdk/v4/logging"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/migrate"
	"github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
	"github.com/turbot/steampipe/pkg/statefile"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/task"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/version"
)

var exitCode int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "steampipe [--version] [--help] COMMAND [args]",
	Version: version.SteampipeVersion.String(),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LogTime("cmd.root.PersistentPreRun start")
		defer utils.LogTime("cmd.root.PersistentPreRun end")

		handleArgDeprecations()

		viper.Set(constants.ConfigKeyActiveCommand, cmd)
		viper.Set(constants.ConfigKeyActiveCommandArgs, args)
		viper.Set(constants.ConfigKeyIsTerminalTTY, isatty.IsTerminal(os.Stdout.Fd()))

		createLogger()
		handleArgDeprecations()
		initGlobalConfig()
		task.RunTasks()

		// set the max memory
		debug.SetMemoryLimit(plugin.GetMaxMemoryBytes())
	},
	Short: "Query cloud resources using SQL",
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

func handleArgDeprecations() {
	if viper.GetString(constants.ArgModLocation) == "" && viper.IsSet(constants.ArgWorkspaceChDir) {
		viper.Set(constants.ArgModLocation, viper.GetString(constants.ArgWorkspaceChDir))
	}
}

func InitCmd() {
	utils.LogTime("cmd.root.InitCmd start")
	defer utils.LogTime("cmd.root.InitCmd end")
	cwd, err := os.Getwd()
	error_helpers.FailOnError(err)

	rootCmd.PersistentFlags().String(constants.ArgInstallDir, filepaths.DefaultInstallDir, "Path to the Config Directory")
	rootCmd.PersistentFlags().String(constants.ArgWorkspaceChDir, cwd, "Path to the workspace working directory")
	rootCmd.PersistentFlags().String(constants.ArgModLocation, cwd, "Path to the workspace working directory")
	rootCmd.PersistentFlags().Bool(constants.ArgSchemaComments, true, "Include schema comments when importing connection schemas")
	// TODO elevate these to specific command? they are not used for plugin or mod commands
	// or else put validation for plugin commands or at least a warning
	rootCmd.PersistentFlags().String(constants.ArgCloudHost, constants.DefaultCloudHost, "Steampipe Cloud host")
	rootCmd.PersistentFlags().String(constants.ArgCloudToken, "", "Steampipe Cloud authentication token")
	rootCmd.PersistentFlags().String(constants.ArgWorkspaceDatabase, constants.DefaultWorkspaceDatabase, "Steampipe Cloud workspace database")
	rootCmd.PersistentFlags().String(constants.ArgWorkspaceProfile, "default", "The workspace profile to use")

	// deprecate ArgWorkspaceChDir
	workspaceChDirFlag := rootCmd.PersistentFlags().Lookup(constants.ArgWorkspaceChDir)
	workspaceChDirFlag.Deprecated = "use --mod-location"

	error_helpers.FailOnError(viper.BindPFlag(constants.ArgInstallDir, rootCmd.PersistentFlags().Lookup(constants.ArgInstallDir)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgWorkspaceChDir, workspaceChDirFlag))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgModLocation, rootCmd.PersistentFlags().Lookup(constants.ArgModLocation)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgCloudHost, rootCmd.PersistentFlags().Lookup(constants.ArgCloudHost)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgCloudToken, rootCmd.PersistentFlags().Lookup(constants.ArgCloudToken)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgWorkspaceDatabase, rootCmd.PersistentFlags().Lookup(constants.ArgWorkspaceDatabase)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgSchemaComments, rootCmd.PersistentFlags().Lookup(constants.ArgSchemaComments)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgWorkspaceProfile, rootCmd.PersistentFlags().Lookup(constants.ArgWorkspaceProfile)))

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

// initConfig reads in config file and ENV variables if set.
func initGlobalConfig() {
	utils.LogTime("cmd.root.initGlobalConfig start")
	defer utils.LogTime("cmd.root.initGlobalConfig end")

	// 1) load workspace profile
	// set install dir to the default to load the workspace profile ( we always load this out of default install dir)
	// (NOTE: _do not_ call ensureInstallDir - we do not want to create the default if it is not there)
	setInstallDir(filepaths.DefaultInstallDir)
	workspaceProfileLoader, err := steampipeconfig.NewWorkspaceProfileLoader(filepaths.WorkspaceProfileDir())
	error_helpers.FailOnError(err)
	workspaceProfile, err := workspaceProfileLoader.Get(viper.GetString(constants.ArgWorkspaceProfile))
	error_helpers.FailOnError(err)
	// set global workspace profile
	steampipeconfig.GlobalWorkspaceProfile = workspaceProfile
	// get the default workspace profile (must be there)
	defaultWorkspaceProfile, _ := workspaceProfileLoader.Get("default")

	// 2) use workspace profile to set-up viper with defaults
	err = cmdconfig.BootstrapViper(defaultWorkspaceProfile)
	error_helpers.FailOnError(err)

	// set global containing the configured install dir (create directory if needed)
	ensureInstallDir(viper.GetString(constants.ArgInstallDir))

	// 3) load the connection config and HCL options
	var cmdName = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command).Name()
	config, err := steampipeconfig.LoadSteampipeConfig(viper.GetString(constants.ArgModLocation), cmdName)
	error_helpers.FailOnError(err)
	// store global config
	steampipeconfig.GlobalConfig = config
	// set viper defaults from this config
	cmdconfig.SetDefaultsFromConfig(steampipeconfig.GlobalConfig.ConfigMap())

	// 4) if an explicit workspace profile was set, add to viper as highest precedence default
	if viper.IsSet(constants.ArgWorkspaceProfile) {
		cmdconfig.SetDefaultsFromWorkspaceProfile(workspaceProfile)
		// tildefy all paths in viper
		// (this has already been done in BootstrapViper but we may have added a path from the workspace profile)
		err = cmdconfig.TildefyPaths()
	}

	// migrate all legacy config files to use snake casing (migrated in v0.14.0)
	err = migrateLegacyFiles()
	error_helpers.FailOnErrorWithMessage(err, "failed to migrate steampipe data files")

	// now validate all config values have appropriate values
	err = validateConfig()
	error_helpers.FailOnErrorWithMessage(err, "failed to validate config")

	/*
		func (c *WorkspaceProfile) Initialise() hcl.Diagnostics {
			var diags hcl.Diagnostics
			var err error
			if c.InstallDir != "" {
				c.InstallDir, err = files.Tildefy(c.InstallDir)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{Severity: hcl.DiagError, Summary: err.Error()})
				}
			}

			if c.ModLocation != "" {
				c.ModLocation, err = files.Tildefy(c.ModLocation)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{Severity: hcl.DiagError, Summary: err.Error()})
				}
			}

			if c.snapshotLocationIsFilePath() {
				// so snapshot location _is_ file path
				// handle ~
				c.SnapshotLocation, err = files.Tildefy(c.SnapshotLocation)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{Severity: hcl.DiagError, Summary: err.Error()})
				}

				// ensure location exists
				if !files.DirectoryExists(c.SnapshotLocation) {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("SnapshotLocation %s does not exist in local file system", c.SnapshotLocation),
					})
				}
			}
			return diags
		}
	*/
}

// migrate all data files to use snake casing for property names
func migrateLegacyFiles() error {
	// skip migration for plugin manager commands because the plugin-manager will have
	// been started by some other steampipe command, which would have done the migration already
	if viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command).Name() == "plugin-manager" {
		return nil
	}
	return error_helpers.CombineErrors(
		migrate.Migrate(&statefile.State{}, filepaths.LegacyStateFilePath()),
		migrate.Migrate(&steampipeconfig.ConnectionDataMap{}, filepaths.ConnectionStatePath()),
		migrate.Migrate(&versionfile.PluginVersionFile{}, filepaths.PluginVersionFilePath()),
		migrate.Migrate(&versionfile.DatabaseVersionFile{}, filepaths.DatabaseVersionFilePath()),
	)
}

func handleArgDeprecations() {
	if viper.GetString(constants.ArgModLocation) == "" {
		viper.Set(constants.ArgModLocation, viper.GetString(constants.ArgWorkspaceChDir))
	}
}

// now validate  config values have appropriate values
// (currently validates telemetry)
func validateConfig() error {
	telemetry := viper.GetString(constants.ArgTelemetry)
	if !helpers.StringSliceContains(constants.TelemetryLevels, telemetry) {
		return fmt.Errorf(`invalid value of 'telemetry' (%s), must be one of: %s`, telemetry, strings.Join(constants.TelemetryLevels, ", "))
	}
	return nil
}

// create a hclog logger with the level specified by the SP_LOG env var
func createLogger() {
	level := logging.LogLevel()

	options := &hclog.LoggerOptions{
		Name:       "steampipe",
		Level:      hclog.LevelFromString(level),
		TimeFn:     func() time.Time { return time.Now().UTC() },
		TimeFormat: "2006-01-02 15:04:05.000 UTC",
	}
	if options.Output == nil {
		options.Output = os.Stderr
	}
	logger := hclog.New(options)
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
}

// set the top level ~/.steampipe folder (creates if it doesnt exist)
func setInstallDir(installDir string) {
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		err = os.MkdirAll(installDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("could not create installation directory: %s", installDir))
	}
	filepaths.SteampipeDir = installDir
}

func ensureInstallDir(installDir string) {
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		err = os.MkdirAll(installDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("could not create installation directory: %s", installDir))
	}

	// store as SteampipeDir
	filepaths.SteampipeDir = installDir
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
	var statusRenderer statushooks.StatusHooks = statushooks.NullHooks
	// if the client is a TTY, inject a status spinner
	if isatty.IsTerminal(os.Stdout.Fd()) {
		statusRenderer = statushooks.NewStatusSpinner()
	}

	ctx := statushooks.AddStatusHooksToContext(context.Background(), statusRenderer)
	return ctx
}
