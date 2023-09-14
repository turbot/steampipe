package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-hclog"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/logging"
	sdklogging "github.com/turbot/steampipe-plugin-sdk/v5/logging"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/cloud"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/constants/runtime"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/task"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/version"
	"golang.org/x/exp/maps"
)

var exitCode int
var waitForTasksChannel chan struct{}
var tasksCancelFn context.CancelFunc

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "steampipe [--version] [--help] COMMAND [args]",
	Version: version.SteampipeVersion.String(),
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		utils.LogTime("cmd.PersistentPostRun start")
		defer utils.LogTime("cmd.PersistentPostRun end")
		if waitForTasksChannel != nil {
			// wait for the async tasks to finish
			select {
			case <-time.After(100 * time.Millisecond):
				tasksCancelFn()
				return
			case <-waitForTasksChannel:
				return
			}
		}
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LogTime("cmd.root.PersistentPreRun start")
		defer utils.LogTime("cmd.root.PersistentPreRun end")

		viper.Set(constants.ConfigKeyActiveCommand, cmd)
		viper.Set(constants.ConfigKeyActiveCommandArgs, args)
		viper.Set(constants.ConfigKeyIsTerminalTTY, isatty.IsTerminal(os.Stdout.Fd()))

		// steampipe completion should not create INSTALL DIR or seup/init global config
		if cmd.Name() == "completion" {
			return
		}

		// create a buffer which can be used as a sink for log writes
		// till INSTALL_DIR is setup in initGlobalConfig
		logBuffer := bytes.NewBuffer([]byte{})

		// create a logger before initGlobalConfig - we may need to reinitialize the logger
		// depending on the value of the log_level value in global general options
		createLogger(logBuffer, cmd)

		// set up the global viper config with default values from
		// config files and ENV variables
		ew := initGlobalConfig()

		// if the log level was set in the general config
		if logLevelNeedsReset() {
			logLevel := viper.GetString(constants.ArgLogLevel)
			// set my environment to the desired log level
			// so that this gets inherited by any other process
			// started by this process (postgres/plugin-manager)
			error_helpers.FailOnErrorWithMessage(
				os.Setenv(sdklogging.EnvLogLevel, logLevel),
				"Failed to setup logging",
			)
		}

		// recreate the logger
		// this will put the new log level (if any) to effect as well as start streaming to the
		// log file.
		createLogger(logBuffer, cmd)

		// runScheduledTasks skips running tasks if this instance is the plugin manager
		waitForTasksChannel = runScheduledTasks(cmd.Context(), cmd, args, ew)

		// ensure all plugin installation directories have a version.json file
		// (this is to handle the case of migrating an existing installation from v0.20.x)
		// no point doing this for the plugin-manager since that would have been done by the initiating CLI process
		if !task.IsPluginManagerCmd(cmd) {
			versionfile.EnsureVersionFilesInPluginDirectories()
		}

		// set the max memory if specified
		setMemoryLimit()
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

func setMemoryLimit() {
	maxMemoryBytes := viper.GetInt64(constants.ArgMemoryMaxMb) * 1024 * 1024
	if maxMemoryBytes > 0 {
		// set the max memory
		debug.SetMemoryLimit(maxMemoryBytes)
	}
}

// runScheduledTasks runs the task runner and returns a channel which is closed when
// task run is complete
//
// runScheduledTasks skips running tasks if this instance is the plugin manager
func runScheduledTasks(ctx context.Context, cmd *cobra.Command, args []string, ew *error_helpers.ErrorAndWarnings) chan struct{} {
	// skip running the task runner if this is the plugin manager
	// since it's supposed to be a daemon
	if task.IsPluginManagerCmd(cmd) {
		return nil
	}

	taskUpdateCtx, cancelFn := context.WithCancel(ctx)
	tasksCancelFn = cancelFn

	return task.RunTasks(
		taskUpdateCtx,
		cmd,
		args,
		// pass the config value in rather than runRasks querying viper directly - to avoid concurrent map access issues
		// (we can use the update-check viper config here, since initGlobalConfig has already set it up
		// with values from the config files and ENV settings - update-check cannot be set from the command line)
		task.WithUpdateCheck(viper.GetBool(constants.ArgUpdateCheck)),
		// show deprecation warnings
		task.WithPreHook(func(_ context.Context) {
			displayDeprecationWarnings(ew)
		}),
	)

}

// the log level will need resetting if
//
//	this process does not have a log level set in it's environment
//	the GlobalConfig has a loglevel set
func logLevelNeedsReset() bool {
	envLogLevelIsSet := envLogLevelSet()
	generalOptionsSet := (steampipeconfig.GlobalConfig.GeneralOptions != nil && steampipeconfig.GlobalConfig.GeneralOptions.LogLevel != nil)

	return !envLogLevelIsSet && generalOptionsSet
}

// envLogLevelSet checks whether any of the current or legacy log level env vars are set
func envLogLevelSet() bool {
	_, ok := os.LookupEnv(sdklogging.EnvLogLevel)
	if ok {
		return ok
	}
	// handle legacy env vars
	for _, e := range sdklogging.LegacyLogLevelEnvVars {
		_, ok = os.LookupEnv(e)
		if ok {
			return ok
		}
	}
	return false
}

func InitCmd() {
	utils.LogTime("cmd.root.InitCmd start")
	defer utils.LogTime("cmd.root.InitCmd end")

	rootCmd.SetVersionTemplate(fmt.Sprintf("Steampipe v%s\n", version.SteampipeVersion.String()))

	// global flags
	rootCmd.PersistentFlags().String(constants.ArgWorkspaceProfile, "default", "The workspace profile to use") // workspace profile profile is a global flag since install-dir(global) can be set through the workspace profile
	rootCmd.PersistentFlags().String(constants.ArgInstallDir, filepaths.DefaultInstallDir, "Path to the Config Directory")
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

// initConfig reads in config file and ENV variables if set.
func initGlobalConfig() *error_helpers.ErrorAndWarnings {
	utils.LogTime("cmd.root.initGlobalConfig start")
	defer utils.LogTime("cmd.root.initGlobalConfig end")

	// load workspace profile from the configured install dir
	loader, err := getWorkspaceProfileLoader()
	error_helpers.FailOnError(err)

	// set global workspace profile
	steampipeconfig.GlobalWorkspaceProfile = loader.GetActiveWorkspaceProfile()

	var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	// set-up viper with defaults from the env and default workspace profile
	err = cmdconfig.BootstrapViper(loader, cmd)
	error_helpers.FailOnError(err)

	// set global containing the configured install dir (create directory if needed)
	ensureInstallDir(viper.GetString(constants.ArgInstallDir))

	// load the connection config and HCL options
	config, loadConfigErrorsAndWarnings := steampipeconfig.LoadSteampipeConfig(viper.GetString(constants.ArgModLocation), cmd.Name())
	error_helpers.FailOnError(loadConfigErrorsAndWarnings.GetError())

	// store global config
	steampipeconfig.GlobalConfig = config

	// set viper defaults from this config
	cmdconfig.SetDefaultsFromConfig(steampipeconfig.GlobalConfig.ConfigMap())

	// set the rest of the defaults from ENV
	// ENV takes precedence over any default configuration
	cmdconfig.SetDefaultsFromEnv()

	// if an explicit workspace profile was set, add to viper as highest precedence default
	// NOTE: if install_dir/mod_location are set these will already have been passed to viper by BootstrapViper
	// since the "ConfiguredProfile" is passed in through a cmdline flag, it will always take precedence
	if loader.ConfiguredProfile != nil {
		cmdconfig.SetDefaultsFromConfig(loader.ConfiguredProfile.ConfigMap(cmd))
	}

	// NOTE: we need to resolve the token separately
	// - that is because we need the resolved value of ArgCloudHost in order to load any saved token
	// and we cannot get this until the other config has been resolved
	err = setCloudTokenDefault(loader)
	error_helpers.FailOnError(err)

	// now validate all config values have appropriate values
	err = validateConfig()
	error_helpers.FailOnErrorWithMessage(err, "failed to validate config")

	return loadConfigErrorsAndWarnings
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
	cmdconfig.SetDefaultFromEnv(constants.EnvCloudToken, constants.ArgCloudToken, cmdconfig.String)

	// 4) explicit workspace profile
	if p := loader.ConfiguredProfile; p != nil && p.CloudToken != nil {
		viper.SetDefault(constants.ArgCloudToken, *p.CloudToken)
	}
	return nil
}

func getWorkspaceProfileLoader() (*steampipeconfig.WorkspaceProfileLoader, error) {
	// set viper default for workspace profile, using STEAMPIPE_WORKSPACE env var
	cmdconfig.SetDefaultFromEnv(constants.EnvWorkspaceProfile, constants.ArgWorkspaceProfile, cmdconfig.String)
	// set viper default for install dir, using STEAMPIPE_INSTALL_DIR env var
	cmdconfig.SetDefaultFromEnv(constants.EnvInstallDir, constants.ArgInstallDir, cmdconfig.String)

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

// now validate  config values have appropriate values
// (currently validates telemetry)
func validateConfig() error {
	telemetry := viper.GetString(constants.ArgTelemetry)
	if !helpers.StringSliceContains(constants.TelemetryLevels, telemetry) {
		return fmt.Errorf(`invalid value of 'telemetry' (%s), must be one of: %s`, telemetry, strings.Join(constants.TelemetryLevels, ", "))
	}
	diagnostics, ok := os.LookupEnv(plugin.EnvDiagnosticsLevel)
	if ok {

		if _, isValid := plugin.ValidDiagnosticsLevels[strings.ToUpper(diagnostics)]; !isValid {
			return fmt.Errorf(`invalid value of '%s' (%s), must be one of: %s`, plugin.EnvDiagnosticsLevel, diagnostics, strings.Join(maps.Keys(plugin.ValidDiagnosticsLevels), ", "))
		}
	}
	return nil
}

// create a hclog logger with the level specified by the SP_LOG env var
func createLogger(logBuffer *bytes.Buffer, cmd *cobra.Command) {
	if task.IsPluginManagerCmd(cmd) {
		// nothing to do here - plugin manager sets up it's own logger
		// refer https://github.com/turbot/steampipe/blob/710a96d45fd77294de8d63d77bf78db65133e5ca/cmd/plugin_manager.go#L102
		return
	}

	level := sdklogging.LogLevel()
	var logDestination io.Writer
	if len(filepaths.SteampipeDir) == 0 {
		// write to the buffer - this is to make sure that we don't lose logs
		// till the time we get the log directory
		logDestination = logBuffer
	} else {
		logDestination = logging.NewRotatingLogWriter(filepaths.EnsureLogDir(), "steampipe")

		// write out the buffered contents
		_, _ = logDestination.Write(logBuffer.Bytes())
	}

	hcLevel := hclog.LevelFromString(level)

	options := &hclog.LoggerOptions{
		// make the name unique so that logs from this instance can be filtered
		Name:       fmt.Sprintf("steampipe [%s]", runtime.ExecutionID),
		Level:      hcLevel,
		Output:     logDestination,
		TimeFn:     func() time.Time { return time.Now().UTC() },
		TimeFormat: "2006-01-02 15:04:05.000 UTC",
	}
	logger := sdklogging.NewLogger(options)
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)

	// if the buffer is empty then this is the first time the logger is getting setup
	// write out a banner
	if logBuffer.Len() == 0 {
		// pump in the initial set of logs
		// this will also write out the Execution ID - enabling easy filtering of logs for a single execution
		// we need to do this since all instances will log to a single file and logs will be interleaved
		log.Printf("[INFO] ********************************************************\n")
		log.Printf("[INFO] **%16s%20s%16s**\n", " ", fmt.Sprintf("Steampipe [%s]", runtime.ExecutionID), " ")
		log.Printf("[INFO] ********************************************************\n")
		log.Printf("[INFO] Version:   v%s\n", version.VersionString)
		log.Printf("[INFO] Log level: %s\n", sdklogging.LogLevel())
		log.Printf("[INFO] Log date: %s\n", time.Now().Format("2006-01-02"))
		//
	}
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
	statusRenderer := statushooks.NullHooks
	// if the client is a TTY, inject a status spinner
	if isatty.IsTerminal(os.Stdout.Fd()) {
		statusRenderer = statushooks.NewStatusSpinnerHook()
	}

	ctx := statushooks.AddStatusHooksToContext(context.Background(), statusRenderer)
	return ctx
}

// displayDeprecationWarnings shows the deprecated warnings in a formatted way
func displayDeprecationWarnings(errorsAndWarnings *error_helpers.ErrorAndWarnings) {
	if len(errorsAndWarnings.Warnings) > 0 {
		fmt.Println(color.YellowString(fmt.Sprintf("\nDeprecation %s:", utils.Pluralize("warning", len(errorsAndWarnings.Warnings)))))
		for _, warning := range errorsAndWarnings.Warnings {
			fmt.Printf("%s\n\n", warning)
		}
		fmt.Println("For more details, see https://steampipe.io/docs/reference/config-files/workspace")
		fmt.Println()
	}
}
