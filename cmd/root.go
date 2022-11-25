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
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/logging"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/cloud"
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
var waitForTasksChannel chan struct{}
var tasksCancel context.CancelFunc

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "steampipe [--version] [--help] COMMAND [args]",
	Version: version.SteampipeVersion.String(),
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		utils.LogTime("cmd.PersistentPostRun start")
		defer utils.LogTime("cmd.PersistentPostRun end")
		if waitForTasksChannel != nil {
			// wait for the async tasks to finish
			select {
			case <-time.After(100 * time.Millisecond):
				tasksCancel()
				return
			case <-waitForTasksChannel:
				return
			}
		}
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LogTime("cmd.root.PersistentPreRun start")
		defer utils.LogTime("cmd.root.PersistentPreRun end")

		handleArgDeprecations()

		viper.Set(constants.ConfigKeyActiveCommand, cmd)
		viper.Set(constants.ConfigKeyActiveCommandArgs, args)
		viper.Set(constants.ConfigKeyIsTerminalTTY, isatty.IsTerminal(os.Stdout.Fd()))

		createLogger()

		initGlobalConfig()

		var taskUpdateCtx context.Context
		taskUpdateCtx, tasksCancel = context.WithCancel(cmd.Context())

		waitForTasksChannel = task.RunTasks(taskUpdateCtx, cmd, args)

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

func handleArgDeprecations() {
	if !viper.IsSet(constants.ArgModLocation) && viper.IsSet(constants.ArgWorkspaceChDir) {
		viper.Set(constants.ArgModLocation, viper.GetString(constants.ArgWorkspaceChDir))
	}
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

	// load workspace profile from the configured install dir
	loader, err := loadWorkspaceProfile()
	error_helpers.FailOnError(err)

	// set global workspace profile
	steampipeconfig.GlobalWorkspaceProfile = loader.GetActiveWorkspaceProfile()

	// set-up viper with defaults from the env and default workspace profile
	err = cmdconfig.BootstrapViper(loader)
	error_helpers.FailOnError(err)

	// set global containing the configured install dir (create directory if needed)
	ensureInstallDir(viper.GetString(constants.ArgInstallDir))

	// load the connection config and HCL options
	var cmdName = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command).Name()
	config, err := steampipeconfig.LoadSteampipeConfig(viper.GetString(constants.ArgModLocation), cmdName)
	error_helpers.FailOnError(err)

	// store global config
	steampipeconfig.GlobalConfig = config

	// set viper defaults from this config
	cmdconfig.SetDefaultsFromConfig(steampipeconfig.GlobalConfig.ConfigMap())

	// if an explicit workspace profile was set, add to viper as highest precedence default
	// NOTE: if install_dir/mod_location are set these will already have been passed to viper by BootstrapViper
	if loader.ConfiguredProfile != nil {
		cmdconfig.SetDefaultsFromConfig(loader.ConfiguredProfile.ConfigMap())
	}

	// NOTE: we need to resolve the token separately
	// - that is because we need the resolved value of ArgCloudHost in order to load any saved token
	// and we cannot get this until the other config has been resolved
	err = setCloudTokenDefault(loader)
	error_helpers.FailOnError(err)

	// now validate all config values have appropriate values
	err = validateConfig()
	error_helpers.FailOnErrorWithMessage(err, "failed to validate config")

	// migrate all legacy config files to use snake casing (migrated in v0.14.0)
	err = migrateLegacyFiles()
	error_helpers.FailOnErrorWithMessage(err, "failed to migrate steampipe data files")
}

func setCloudTokenDefault(loader *steampipeconfig.WorkspaceProfileLoader) error {
	/*
	   saved cloud token
	   cloud_token in default workspace
	   explicit env var (STEAMIPE_CLOUD_TOKEN ) wins over
	   cloud_token in specific workspace
	*/
	// set viper defaults in order of increasing precedence
	// 1) saved cloud token
	savedToken, err := cloud.LoadToken()
	if err != nil {
		return err
	}
	if savedToken != "" {
		viper.SetDefault(constants.ArgCloudToken, savedToken)
	}
	// 2) default profile cloud token
	if loader.DefaultProfile.CloudToken != nil {
		viper.SetDefault(constants.ArgCloudToken, *loader.DefaultProfile.CloudToken)
	}
	// 3) env var (STEAMIPE_CLOUD_TOKEN )
	cmdconfig.SetDefaultFromEnv(constants.EnvCloudToken, constants.ArgCloudToken, "string")

	// 4) explicit workspace profile
	if p := loader.ConfiguredProfile; p != nil && p.CloudToken != nil {
		viper.SetDefault(constants.ArgCloudToken, *p.CloudToken)
	}
	return nil
}

func loadWorkspaceProfile() (*steampipeconfig.WorkspaceProfileLoader, error) {
	// set viper default for workspace profile, using STEAMPIPE_WORKSPACE env var
	cmdconfig.SetDefaultFromEnv(constants.EnvWorkspaceProfile, constants.ArgWorkspaceProfile, "string")
	// set viper default for install dir, using STEAMPIPE_INSTALL_DIR env var
	cmdconfig.SetDefaultFromEnv(constants.EnvInstallDir, constants.ArgInstallDir, "string")

	// resolve the workspace profile dir
	installDir, err := filehelpers.Tildefy(viper.GetString(constants.ArgInstallDir))
	if err != nil {
		return nil, err
	}

	workspaceProfileDir, err := filepaths.WorkspaceProfileDir(installDir)
	if err != nil {
		return nil, err
	}

	// create loader
	loader, err := steampipeconfig.NewWorkspaceProfileLoader(workspaceProfileDir)
	if err != nil {
		return nil, err
	}

	return loader, nil
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

func ensureInstallDir(installDir string) {
	log.Printf("[TRACE] ensureInstallDir %s", installDir)
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		log.Printf("[TRACE] creating install dir")
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
	var statusRenderer statushooks.StatusHooks = statushooks.NullHooks
	// if the client is a TTY, inject a status spinner
	if isatty.IsTerminal(os.Stdout.Fd()) {
		statusRenderer = statushooks.NewStatusSpinner()
	}

	ctx := statushooks.AddStatusHooksToContext(context.Background(), statusRenderer)
	return ctx
}
