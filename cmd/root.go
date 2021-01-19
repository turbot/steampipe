package cmd

import (
	"log"
	"path/filepath"

	"github.com/turbot/steampipe/cmdconfig"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/version"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var requiredColor = color.New(color.Bold).SprintfFunc()

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
  steampipe help query`,
}

//var viper *v.Viper

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	log.Println("[TRACE] rootCmd Execute")
	return rootCmd.Execute()
}

func init() {
	var cfgFile string

	defaultCfgFile := filepath.Join(constants.ConfigDir(), constants.DefaultConfigFileName)

	// TODO(nw) replace --config with --config-dir, it's a directory of settings files
	rootCmd.PersistentFlags().StringVar(&cfgFile, constants.ArgConfig, defaultCfgFile, "Name of the config file")
	viper.BindPFlag(constants.ArgConfig, rootCmd.PersistentFlags().Lookup(constants.ArgConfig))

	// TODO(nw) - Add color bool flag, default true, description "Use colors in output", persistent through levels
	cobra.OnInitialize(initGlobalConfig)
}

// initConfig reads in config file and ENV variables if set.
func initGlobalConfig() {
	log.Println("[TRACE] rootCmd initGlobalConfig")
	cmdconfig.InitViper(viper.GetViper())
}
