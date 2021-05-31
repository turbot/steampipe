package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/task"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/version"
)

var exitCode int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "steampipe [--version] [--help] COMMAND [args]",
	Version: version.String(),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		viper.Set(constants.ConfigKeyActiveCommand, cmd)
		viper.Set(constants.ConfigKeyActiveCommandArgs, args)
		initGlobalConfig()
		createLogger()
		task.RunTasks()
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

	rootCmd.PersistentFlags().String(constants.ArgInstallDir, constants.DefaultInstallDir, "Path to the Config Directory")
	rootCmd.PersistentFlags().String(constants.ArgWorkspace, "", "Path to the workspace (default to current working directory) ")

	viper.BindPFlag(constants.ArgInstallDir, rootCmd.PersistentFlags().Lookup(constants.ArgInstallDir))
	viper.BindPFlag(constants.ArgWorkspace, rootCmd.PersistentFlags().Lookup(constants.ArgWorkspace))

	AddCommands()

}

// initConfig reads in config file and ENV variables if set.
func initGlobalConfig() {
	// set global containing install dir
	setInstallDir()

	workspace := viper.GetString(constants.ArgWorkspace)
	if workspace == "" {
		// default to working directory
		workingDir, err := os.Getwd()
		utils.FailOnErrorWithMessage(err, "could not read current directory")
		workspace = workingDir
		viper.Set(constants.ArgWorkspace, workspace)
	}

	// load config (this sets the global config steampipeconfig.Config)
	var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	config, err := steampipeconfig.LoadSteampipeConfig(workspace, cmd.Name())
	utils.FailOnError(err)

	steampipeconfig.Config = config

	// set viper config defaults from config and env vars
	cmdconfig.SetViperDefaults(steampipeconfig.Config)
}

// CreateLogger :: create a hclog logger with the level specified by the SP_LOG env var
func createLogger() {
	// TODO GET FROM VIPER
	level := logging.LogLevel()

	options := &hclog.LoggerOptions{Name: "steampipe", Level: hclog.LevelFromString(level)}
	if options.Output == nil {
		options.Output = os.Stderr
	}
	logger := hclog.New(options)
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
}

// SteampipeDir :: set the top level ~/.steampipe folder (creates if it doesnt exist)
func setInstallDir() {
	installDir, err := helpers.Tildefy(viper.GetString(constants.ArgInstallDir))
	utils.FailOnErrorWithMessage(err, fmt.Sprintf("failed to sanitize install directory"))
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		err = os.MkdirAll(installDir, 0755)
		utils.FailOnErrorWithMessage(err, fmt.Sprintf("could not create installation directory: %s", installDir))
	}
	constants.SteampipeDir = installDir
}

func AddCommands() {
	// explicitly initialise commands here rather than in init functions to allow us to handle errors from the config load
	rootCmd.AddCommand(PluginCmd())
	rootCmd.AddCommand(QueryCmd())
	rootCmd.AddCommand(CheckCmd())
	rootCmd.AddCommand(ServiceCmd())
}

func Execute() int {
	rootCmd.Execute()
	return exitCode
}
