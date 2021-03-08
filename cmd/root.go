package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	log.Println("[TRACE] rootCmd Execute")
	return rootCmd.Execute()
}

func init() {
	// TODO(nw) - Add color bool flag, default true, description "Use colors in output", persistent through levels

	defaultCfgFile := "~/.steampipe"

	// TODO(nw) replace --config with --config-dir, it's a directory of settings files
	rootCmd.PersistentFlags().String(constants.ArgConfigDir, defaultCfgFile, "Path to the Config Directory")
	viper.BindPFlag(constants.ArgConfigDir, rootCmd.PersistentFlags().Lookup(constants.ArgConfigDir))

	cobra.OnInitialize(initGlobalConfig)
}

// initConfig reads in config file and ENV variables if set.
func initGlobalConfig() {
	log.Println("[TRACE] rootCmd initGlobalConfig")
	cmdconfig.InitViper()
}
