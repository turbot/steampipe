package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "steampipe [--version] [--help] COMMAND [args]",
	Version: version.String(),
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
	rootCmd.PersistentFlags().String(constants.ArgInstallDir, constants.DefaultInstallDir, "Path to the Config Directory")
	viper.BindPFlag(constants.ArgInstallDir, rootCmd.PersistentFlags().Lookup(constants.ArgInstallDir))

	AddCommands()
	cobra.OnInitialize(initGlobalConfig)
}

// initConfig reads in config file and ENV variables if set.
func initGlobalConfig() {
	log.Println("[TRACE] rootCmd initGlobalConfig")
	cmdconfig.InitViper()
	// set global containing install dir
	SetInstallDir()

	// load config
	config, err := steampipeconfig.Load()
	if err != nil {
		utils.ShowError(err)
		return
	}
	// todo set viper config from config
	setViperDefaults(config)
}

// SteampipeDir :: set the top level ~/.steampipe folder (creates if it doesnt exist)
func SetInstallDir() {
	installDir, err := helpers.Tildefy(viper.GetString(constants.ArgInstallDir))
	utils.FailOnErrorWithMessage(err, fmt.Sprintf("failed to sanitize install directory"))
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		err = os.MkdirAll(installDir, 0755)
		utils.FailOnErrorWithMessage(err, fmt.Sprintf("could not create installation directory: %s", installDir))
	}
	constants.SteampipeDir = installDir
}

func setViperDefaults(config *steampipeconfig.SteampipeConfig) {
	setViperDefaultsFromConfig(config)
	overrideViperDefaultsFromEnv()
}

func setViperDefaultsFromConfig(config *steampipeconfig.SteampipeConfig) {
	for k, v := range config.ConfigMap() {
		log.Println("[TRACE]", "root", "overrideViperDefaultWithConfig", fmt.Sprintf("Setting %s to %v", k, v))
		viper.SetDefault(k, v)
	}
}

func overrideViperDefaultsFromEnv() {
	// a map of environment variables to Viper Config Key
	ingest := map[string]string{}
	for k, v := range ingest {
		if val, ok := os.LookupEnv(k); ok {
			viper.SetDefault(v, val)
		}
	}
}

func AddCommands() {
	// explicitly initialise commands here rather than in init functions to allow us to handle errors from the config load
	rootCmd.AddCommand(PluginCmd())
	rootCmd.AddCommand(QueryCmd())
	rootCmd.AddCommand(ServiceCmd())
}

func Execute() {
	rootCmd.Execute()
}
