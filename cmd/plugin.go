package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/ociinstaller/versionfile"
	"github.com/turbot/steampipe/plugin"
	"github.com/turbot/steampipe/statefile"
	"github.com/turbot/steampipe/utils"

	"github.com/spf13/cobra"
	"github.com/turbot/steampipe-plugin-sdk/logging"
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
		Use:   "install [flags] [registry/org/]name[@version]",
		Args:  cobra.ArbitraryArgs,
		Run:   runPluginInstallCmd,
		Short: "Install one or more plugins",
		Long: `Install one or more plugins.

Install a Steampipe plugin, making it available for queries and configuration.
The plugin name format is [registry/org/]name[@version]. The default
registry is hub.steampipe.io, default org is turbot and default version
is latest. The name is a required argument.

Examples:

  # Install a common plugin (turbot/aws)
  steampipe plugin install aws

  # Install a specific plugin version
  steampipe plugin install turbot/azure@0.1.0`,
	}

	cmdconfig.
		OnCmd(cmd)

	return cmd
}

// PluginUpdateCmd :: Update plugins
func PluginUpdateCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "update [flags] [registry/org/]name[@version]",
		Args:  cobra.ArbitraryArgs,
		Run:   runPluginUpdateCmd,
		Short: "Update one or more plugins",
		Long: `Update plugins.

Update one or more Steampipe plugins, making it available for queries and configuration.
The plugin name format is [registry/org/]name[@version]. The default
registry is hub.steampipe.io, default org is turbot and default version
is latest. The name is a required argument.

Examples:

  # Update all plugins to their latest available version 
  steampipe plugin update --all

  # Update a common plugin (turbot/aws)
  steampipe plugin update aws`,
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
		Use:   "uninstall [flags] [registry/org/]name",
		Args:  cobra.ArbitraryArgs,
		Run:   runPluginUninstallCmd,
		Short: "Uninstall a plugin",
		Long: `Uninstall a plugin.

Uninstall a Steampipe plugin, removing it from use. The plugin name format is
[registry/org/]name. (Version is not relevant in uninstall, since only one
version of a plugin can be installed at a time.)

Example:

  # Uninstall a common plugin (turbot/aws)
  steampipe plugin uninstall aws

`,
	}

	cmdconfig.OnCmd(cmd)

	return cmd
}

func runPluginInstallCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runPluginInstallCmd install")
	defer func() {
		logging.LogTime("runPluginInstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	// args to 'plugin install' -- one or more plugins to install
	// These can be simple names ('aws') for "standard" plugins, or
	// full refs to the OCI image (us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0)
	plugins := append([]string{}, args...)
	installSkipped := []string{}

	if len(plugins) == 0 {
		fmt.Println()
		utils.ShowError(fmt.Errorf("you need to provide at least one plugin to install"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		return
	}

	if len(plugins) > 1 {
		fmt.Println()
	}

	for idx, p := range plugins {
		isPluginExists, _ := plugin.Exists(p)
		if isPluginExists {
			installSkipped = append(installSkipped, p)
			continue
		}
		if idx > 0 {
			fmt.Println()
		}
		spinner := utils.ShowSpinner(fmt.Sprintf("Installing plugin %s...", p))
		image, err := plugin.Install(p)
		utils.StopSpinner(spinner)
		if err != nil {
			msg := fmt.Sprintf("install failed for plugin '%s'", p)

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
		fmt.Printf("Installed plugin: %s%s\n", constants.Bold(p), versionString)
		org := image.Config.Plugin.Organization
		if org == "turbot" {
			fmt.Printf("Documentation:    https://hub.steampipe.io/plugins/%s/%s\n", org, p)
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

type updateSkip struct {
	plugin string
	reason string
}

func (u *updateSkip) String() string {
	return fmt.Sprintf("Plugin:   %s\nReason:   %s", u.plugin, u.reason)
}

func runPluginUpdateCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runPluginUpdateCmd install")
	defer func() {
		logging.LogTime("runPluginUpdateCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	// args to 'plugin update' -- one or more plugins to install
	// These can be simple names ('aws') for "standard" plugins, or
	// full refs to the OCI image (us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0)
	plugins := append([]string{}, args...)

	if len(plugins) == 0 && !(cmdconfig.Viper().GetBool("all")) {
		fmt.Println()
		utils.ShowError(fmt.Errorf("you need to provide at least one plugin to update or use the %s flag", constants.Bold("--all")))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		return
	}

	if len(plugins) > 0 && cmdconfig.Viper().GetBool("all") {
		// we can't allow update and install at the same time
		fmt.Println()
		utils.ShowError(fmt.Errorf("%s cannot be used when updating specific plugins", constants.Bold("`--all`")))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		return
	}

	state, err := statefile.LoadState()
	if err != nil {
		utils.ShowError(fmt.Errorf("could not load state"))
		return
	}

	// load up the version file data
	versionData, err := versionfile.Load()
	if err != nil {
		utils.ShowError(fmt.Errorf("error loading current plugin data"))
		return
	}
	if versionData.Plugins == nil {
		versionData.Plugins = map[string]*versionfile.InstalledVersion{}
	}

	var runUpdatesFor []*versionfile.InstalledVersion
	var updateSkipped []updateSkip

	if cmdconfig.Viper().GetBool("all") {
		for k, v := range versionData.Plugins {
			ref := ociinstaller.NewSteampipeImageRef(k)
			org, name, stream := ref.GetOrgNameAndStream()
			key := fmt.Sprintf("%s/%s@%s", org, name, stream)

			plugins = append(plugins, key)
			runUpdatesFor = append(runUpdatesFor, v)
		}
	} else {
		// get the args and retrieve the installed versions
		for _, p := range plugins {
			ref := ociinstaller.NewSteampipeImageRef(p)
			isExists, _ := plugin.Exists(p)
			if isExists {
				runUpdatesFor = append(runUpdatesFor, versionData.Plugins[ref.DisplayImageRef()])
			} else {
				updateSkipped = append(updateSkipped, updateSkip{p, "Not Installed"})
			}
		}
	}

	if len(plugins) > 0 {
		// add a blank line at the top since this is going to be
		// a multi output
		fmt.Println()
	}

	spinner := utils.ShowSpinner("Checking for available updates")
	reports := plugin.GetUpdateReport(state.InstallationID, runUpdatesFor)
	utils.StopSpinner(spinner)

	if len(reports) < len(runUpdatesFor) {
		// this happens if for some reason the update server could not be contacted,
		// in which case we get back an empty map
		utils.ShowError(fmt.Errorf("there was an issue contacting the update server. Please try later"))
		return
	}

	for _, report := range reports {
		if report.Plugin.ImageDigest == report.CheckResponse.Digest {
			updateSkipped = append(updateSkipped, updateSkip{
				fmt.Sprintf("%s@%s", report.CheckResponse.Name, report.CheckResponse.Stream),
				"Latest already installed",
			})
			continue
		}

		spinner := utils.ShowSpinner(fmt.Sprintf("Updating plugin %s...", report.CheckResponse.Name))
		image, err := plugin.Install(report.Plugin.Name)
		utils.StopSpinner(spinner)
		if err != nil {
			msg := fmt.Sprintf("update failed for plugin '%s'", report.Plugin.Name)

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
		fmt.Printf("Updated plugin: %s%s\n", constants.Bold(report.Plugin.Name), versionString)
		org := image.Config.Plugin.Organization
		name := image.Config.Plugin.Name
		if org == "turbot" {
			fmt.Printf("Documentation:  https://hub.steampipe.io/plugins/%s/%s\n", org, name)
		}
		fmt.Println()
	}

	if len(updateSkipped) > 0 {
		skipReasons := []string{}
		for _, s := range updateSkipped {
			skipReasons = append(skipReasons, s.String())
		}

		fmt.Printf(
			"Skipped the following %s:\n\n%s",
			utils.Pluralize("plugin", len(updateSkipped)),
			strings.Join(skipReasons, "\n\n"),
		)
		if len(updateSkipped) > 1 {
			fmt.Println()
		}
	}

	if len(plugins) > 1 {
		fmt.Println()
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
	defer func() {
		logging.LogTime("runPluginListCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	connectionMap, err := getPluginConnectionMap()
	if err != nil {
		utils.ShowErrorWithMessage(err,
			fmt.Sprintf("Plugin Listing failed"))
		return
	}

	list, err := plugin.List(connectionMap)
	if err != nil {
		utils.ShowErrorWithMessage(err,
			fmt.Sprintf("Plugin Listing failed"))
	}
	headers := []string{"Name", "Version", "Connections"}
	rows := [][]string{}
	for _, item := range list {
		rows = append(rows, []string{item.Name, item.Version, strings.Join(item.Connections, ",")})
	}
	display.ShowWrappedTable(headers, rows, false)
}

func runPluginUninstallCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runPluginUninstallCmd uninstall")

	defer func() {
		logging.LogTime("runPluginUninstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	if len(args) == 0 {
		fmt.Println()
		utils.ShowError(fmt.Errorf("you need to provide at least one plugin to uninstall"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		return
	}

	connectionMap, err := getPluginConnectionMap()
	if err != nil {
		utils.ShowError(err)
		return
	}

	for _, p := range args {
		if err := plugin.Remove(p, connectionMap); err != nil {
			utils.ShowErrorWithMessage(err, fmt.Sprintf("Failed to uninstall plugin '%s'", p))
		} else {
			fmt.Println("Uninstalled plugin", p)
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
