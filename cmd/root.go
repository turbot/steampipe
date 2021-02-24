package cmd

import (
	"log"

	"github.com/turbot/steampipe/cmdconfig"

	"github.com/turbot/steampipe/version"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/version"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() error {
	log.Println("[TRACE] RootCmd Execute")
	return RootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().String(constants.ArgInstallDir, constants.DefaultInstallDir, "Path to the Config Directory")
	viper.BindPFlag(constants.ArgInstallDir, rootCmd.PersistentFlags().Lookup(constants.ArgInstallDir))

	cobra.OnInitialize(initGlobalConfig)
}

// initConfig reads in config file and ENV variables if set.
func initGlobalConfig() {
	log.Println("[TRACE] rootCmd initGlobalConfig")
	cmdconfig.InitViper()
}
