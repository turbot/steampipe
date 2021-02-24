package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/statefile"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/pluginmanager"
	"github.com/turbot/steampipe/utils"
)

func init() {
	rootCmd.AddCommand(PluginCmd())
}

// PluginCmd :: Plugin management commands
func PluginCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "plugin [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe plugin management",
		Long: `Steampipe plugin management.

Plugins extend Steampipe to work with many different services and providers.
Find plugins using the public registry at https://registry.steampipe.io.

Examples:

  # Install or update a plugin
  steampipe plugin install aws

  # List installed plugins
  steampipe plugin list

  # Uninstall a plugin
  steampipe plugin uninstall aws`,
	}

	cmd.AddCommand(PluginInstallCmd())
	cmd.AddCommand(PluginListCmd())
	cmd.AddCommand(PluginUninstallCmd())
	cmd.AddCommand(PluginUpdateCmd())

	return cmd
}

// PluginInstallCmd :: Install a plugin
func PluginInstallCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "install [flags] [registry/[org/]]name[:version]",
		Args:  cobra.ArbitraryArgs,
		Run:   runPluginInstallCmd,
		Short: "Install one or more plugins",
		Long: `Install one or more plugins.

Install a Steampipe plugin, making it available for queries and configuration.
The plugin name format is [registry/[org/]]name[:version]. The default
registry is hub.steampipe.io, default org is turbot and default version
is latest. The name is a required argument.

Examples:

  # Install a common plugin (turbot/aws)
  steampipe plugin install aws

  # Install a plugin published by DMI to the public registry
  steampipe plugin install dmi/paper

  # Install a plugin from a private registry
  steampipe plugin install my-registry.dmi.com/dmi/internal

  # Install a specific plugin version
  steampipe plugin install turbot/azure:0.1.0

  # Install multiple plugins at once
  steampipe plugin install aws dmi/paper`,
	}

	cmdconfig.
		OnCmd(cmd)

	return cmd
}

// PluginUpdateCmd :: Update plugins
func PluginUpdateCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "update [flags] [registry/[org/]]name[:version]",
		Args:  cobra.ArbitraryArgs,
		Run:   runPluginUpdateCmd,
		Short: "Update one or more plugins",
		Long: `Update plugins.

Update one or more Steampipe plugins, making it available for queries and configuration.
The plugin name format is [registry/[org/]]name[:version]. The default
registry is hub.steampipe.io, default org is turbot and default version
is latest. The name is a required argument.

Examples:

  # Update a common plugin (turbot/aws)
  steampipe plugin update aws

  # Update a plugin published by DMI to the public registry
  steampipe plugin update dmi/paper

  # Update a plugin from a private registry
  steampipe plugin update my-registry.dmi.com/dmi/internal

  # Update a specific plugin version
  steampipe plugin update turbot/azure:0.1.0

  # Update all plugins to their latest available version (only works for plugins installed from hub.steampipe.io)
  steampipe plugin update --all

  # Update multiple plugins at once
  steampipe plugin update aws dmi/paper`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag("all", "", false, "Update all plugins to its latest available version")

	return cmd
}

// PluginListCmd :: List plugins
func PluginListCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Run:   runPluginListCmd,
		Short: "List currently installed plugins",
		Long: `List currently installed plugins.

List all Steampipe plugins installed for this user.

Examples:

  # List installed plugins
  steampipe plugin list

  # List plugins that have updates available
  steampipe plugin list --outdated`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag("outdated", "", false, "Check each plugin in the list for updates")

	return cmd
}

// PluginUninstallCmd :: Uninstall a plugin
func PluginUninstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "uninstall [flags] [registry/[org/]]name",
		Args:  cobra.ArbitraryArgs,
		Run:   runPluginUninstallCmd,
		Short: "Uninstall a plugin",
		Long: `Uninstall a plugin.

Uninstall a Steampipe plugin, removing it from use. The plugin name format is
[registry/[org/]]name. (Version is not relevant in uninstall, since only one
version of a plugin can be installed at a time.)

Examples:

  # Uninstall a common plugin (turbot/aws)
  steampipe plugin uninstall aws

  # Uninstall a plugin published by DMI to the public registry
  steampipe plugin uninstall dmi/paper

  # Uninstall a plugin from a private registry
  steampipe plugin uninstall my-registry.dmi.com/dmi/internal`,
	}

	cmdconfig.OnCmd(cmd)

	return cmd
}

func runPluginInstallCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runPluginInstallCmd install")
	defer logging.LogTime("runPluginInstallCmd end")

	// args to 'plugin install' -- one or more plugins to install
	// These can be simple names ('aws') for "standard" plugins, or
	// full refs to the OCI image (us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0)
	plugins := append([]string{}, args...)
	installSkipped := []string{}

	if len(plugins) == 0 {
		utils.ShowError(fmt.Errorf("you need to provide at least one plugin to install"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		return
	}

	if len(plugins) > 1 {
		fmt.Println()
	}

	for idx, plugin := range plugins {
		isPluginExists, _ := pluginmanager.IsPluginExists(plugin)
		if isPluginExists {
			installSkipped = append(installSkipped, plugin)
			continue
		}
		if idx > 0 {
			fmt.Println()
		}
		spinner := utils.ShowSpinner(fmt.Sprintf("Installing plugin %s...", plugin))
		image, err := pluginmanager.Install(plugin)
		utils.StopSpinner(spinner)
		if err != nil {
			msg := fmt.Sprintf("install failed for plugin '%s'", plugin)

			if strings.HasSuffix(err.Error(), "not found") {
				msg += ": not found"
			} else {
				log.Printf("[DEBUG] %s", err.Error())
			}
			utils.ShowError(fmt.Errorf(msg))
			continue
		}
		versionString := ""
		if image.Config.Plugin.Version != "" {
			versionString = " v" + image.Config.Plugin.Version
		}
		fmt.Printf("Installed plugin: %s%s\n", constants.Bold(plugin), versionString)
		org := image.Config.Plugin.Organization
		if org == "turbot" {
			fmt.Printf("Documentation:    https://hub.steampipe.io/plugins/%s/%s\n", org, plugin)
		}
	}

	if len(installSkipped) > 0 {
		fmt.Printf(
			"Skipped the following %s, since %s already installed:\n",
			utils.Pluralize("plugin", len(installSkipped)),
			utils.Pluralize("it is", len(installSkipped)),
		)
		for _, s := range installSkipped {
			fmt.Printf("    > %s\n", constants.Bold(s))
		}
		fmt.Printf(
			"\nTo update %s, please run: %s\n",
			utils.Pluralize("this plugin", len(installSkipped)),
			constants.Bold(fmt.Sprintf(
				"steampipe plugin update %s",
				strings.Join(installSkipped, " "),
			)),
		)
	}

	if len(plugins) > 1 {
		fmt.Println("")
	}

	// refresh connections - we do this to validate the plugins
	// ignore errors - if we get this far we have successfully installed
	// reporting an error in the validation may be confusing
	// - we will retry next time query is run and report any errors then
	if len(plugins) > len(installSkipped) {
		refreshConnections()
	}
}

func runPluginUpdateCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runPluginUpdateCmd install")
	defer logging.LogTime("runPluginUpdateCmd end")

	// args to 'plugin update' -- one or more plugins to install
	// These can be simple names ('aws') for "standard" plugins, or
	// full refs to the OCI image (us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0)
	plugins := append([]string{}, args...)

	if len(plugins) == 0 && !cmdconfig.Viper().GetBool("all") {
		utils.ShowError(fmt.Errorf("you need to provide at least one plugin to update or use the %s flag", constants.Bold("--all")))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		return
	}

	// we can't allow update and install at the same time
	if cmdconfig.Viper().GetBool("all") {
		if len(plugins) > 0 {
			utils.ShowError(fmt.Errorf("%s cannot be used when updating specific plugins", constants.Bold("`--all`")))
			return
		}

		// get the update report
		plugins = []string{}
		state, _ := statefile.LoadState()
		reports := pluginmanager.GetPluginUpdateReport(state.InstallationID)
		for _, report := range reports {
			if report.Plugin.ImageDigest != report.CheckResponse.Digest {
				plugins = append(plugins, fmt.Sprintf("%s/%s@%s", report.CheckResponse.Org, report.CheckResponse.Name, report.CheckResponse.Stream))
			}
		}
	}

	if len(plugins) == 0 {
		fmt.Println("All plugins are up to date.")
		return
	}

	updateSkipped := []string{}

	for _, plugin := range plugins {
		isPluginExists, _ := pluginmanager.IsPluginExists(plugin)
		if !isPluginExists {
			updateSkipped = append(updateSkipped, plugin)
			continue
		}
		if len(args) > 1 {
			fmt.Println()
		}

		spinner := utils.ShowSpinner(fmt.Sprintf("Updating plugin %s...", plugin))
		image, err := pluginmanager.Install(plugin)
		utils.StopSpinner(spinner)
		if err != nil {
			if err.Error() == constants.ENOTEXISTS {
				updateSkipped = append(updateSkipped, plugin)
				continue
			}

			msg := fmt.Sprintf("update failed for plugin '%s'", plugin)

			if strings.HasSuffix(err.Error(), "not found") {
				msg += ": not found"
			} else {
				log.Printf("[DEBUG] %s", err.Error())
			}
			utils.ShowError(fmt.Errorf(msg))
			continue
		}
		versionString := ""
		if image.Config.Plugin.Version != "" {
			versionString = " v" + image.Config.Plugin.Version
		}
		fmt.Printf("Updated plugin: %s%s\n", constants.Bold(plugin), versionString)
		org := image.Config.Plugin.Organization
		if org == "turbot" {
			fmt.Printf("Documentation:  https://hub.steampipe.io/plugins/%s/%s\n", org, plugin)
		}
	}

	if len(updateSkipped) > 0 {
		fmt.Printf(
			"\nSkipped the following %s, since %s not installed:\n",
			utils.Pluralize("plugin", len(updateSkipped)),
			utils.Pluralize("it is", len(updateSkipped)),
		)
		for _, s := range updateSkipped {
			fmt.Printf("    > %s\n", constants.Bold(s))
		}
		fmt.Printf(
			"\nTo install %s, please run: %s\n",
			utils.Pluralize("this plugin", len(updateSkipped)),
			constants.Bold(fmt.Sprintf(
				"steampipe plugin install %s",
				strings.Join(updateSkipped, " "),
			)),
		)
	}
	if len(plugins) > 1 {
		fmt.Println("")
	}

	// refresh connections - we do this to validate the plugins
	// ignore errors - if we get this far we have successfully installed
	// reporting an error in the validation may be confusing
	// - we will retry next time query is run and report any errors then
	if len(plugins) > len(updateSkipped) {
		refreshConnections()
	}
}

// start service if necessary and refresh connections
func refreshConnections() error {
	// todo move this into db package
	db.EnsureDBInstalled()
	status, err := db.GetStatus()
	if err != nil {
		return errors.New("could not retrieve service status")
	}

	var client *db.Client
	if status == nil {
		// the db service is not started - start it
		db.StartService(db.InvokerInstaller)
		defer db.Shutdown(client, db.InvokerInstaller)
	}

	client, err = db.GetClient(false)
	if err != nil {
		return err
	}

	// refresh connections
	if err = db.RefreshConnections(client); err != nil {
		return err
	}

	return nil
}

func runPluginListCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runPluginListCmd list")
	defer logging.LogTime("runPluginListCmd end")

	connectionMap, err := getPluginConnectionMap()
	if err != nil {
		utils.ShowErrorWithMessage(err,
			fmt.Sprintf("Plugin Listing failed"))
		return
	}

	list, err := pluginmanager.List(connectionMap)
	if err != nil {
		utils.ShowErrorWithMessage(err,
			fmt.Sprintf("Plugin Listing failed"))
	}
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Version", "Connections"})
	for _, item := range list {
		t.AppendRow(table.Row{item.Name, item.Version, strings.Join(item.Connections, ",")})
	}
	t.Render()
}

func runPluginUninstallCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runPluginUninstallCmd uninstall")
	defer logging.LogTime("runPluginUninstallCmd end")

	connectionMap, err := getPluginConnectionMap()
	if err != nil {
		utils.ShowError(err)
		return
	}

	for _, plugin := range args {
		if err := pluginmanager.Remove(plugin, connectionMap); err != nil {
			utils.ShowErrorWithMessage(err, fmt.Sprintf("Failed to uninstall plugin '%s'", plugin))
		} else {
			fmt.Println("Uninstalled plugin", plugin)
		}
	}
}

// returns a map of pluginFullName -> []{connections using pluginFullName}
func getPluginConnectionMap() (map[string][]string, error) {
	status, err := db.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("Could not start steampipe service")
	}

	if status == nil {
		// the db service is not started - start it
		db.StartService(db.InvokerPlugin)
		defer func() {
			status, _ := db.GetStatus()
			if status.Invoker == db.InvokerPlugin {
				db.StopDB(true)
			}
		}()
	}

	client, err := db.GetClient(true)
	if err != nil {
		return nil, fmt.Errorf("Could not connect with steampipe service")
	}

	pluginConnectionMap := map[string][]string{}

	for k, v := range *client.ConnectionMap() {
		_, found := pluginConnectionMap[v.Plugin]
		if !found {
			pluginConnectionMap[v.Plugin] = []string{}
		}
		pluginConnectionMap[v.Plugin] = append(pluginConnectionMap[v.Plugin], k)
	}
	return pluginConnectionMap, nil
}
