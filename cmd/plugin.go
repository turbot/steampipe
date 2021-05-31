package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/ociinstaller/versionfile"
	"github.com/turbot/steampipe/plugin"
	"github.com/turbot/steampipe/statefile"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

// PluginCmd :: Plugin management commands
func PluginCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "plugin [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe plugin management",
		Long: `Steampipe plugin management.

Plugins extend Steampipe to work with many different services and providers.
Find plugins using the public registry at https://hub.steampipe.io.

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
	utils.LogTime("runPluginInstallCmd install")
	defer func() {
		utils.LogTime("runPluginInstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	// args to 'plugin install' -- one or more plugins to install
	// plugin names can be simple names ('aws') for "standard" plugins,
	// or full refs to the OCI image (us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0)
	plugins := append([]string{}, args...)
	installReports := make([]display.InstallReport, 0, len(plugins))

	if len(plugins) == 0 {
		fmt.Println()
		utils.ShowError(fmt.Errorf("you need to provide at least one plugin to install"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		return
	}

	// a leading blank line - since we always output multiple lines
	fmt.Println()

	spinner := display.ShowSpinner("")

	for _, p := range plugins {
		isPluginExists, _ := plugin.Exists(p)
		if isPluginExists {
			installReports = append(installReports, display.InstallReport{
				Plugin:         p,
				Skipped:        true,
				SkipReason:     display.ALREADY_INSTALLED,
				IsUpdateReport: false,
			})
			continue
		}
		display.UpdateSpinnerMessage(spinner, fmt.Sprintf("Installing plugin: %s", p))
		image, err := plugin.Install(p)
		if err != nil {
			msg := ""
			if strings.HasSuffix(err.Error(), "not found") {
				msg = "Not found"
			} else {
				msg = err.Error()
			}
			installReports = append(installReports, display.InstallReport{
				Skipped:        true,
				Plugin:         p,
				SkipReason:     msg,
				IsUpdateReport: false,
			})
			continue
		}
		versionString := ""
		if image.Config.Plugin.Version != "" {
			versionString = " v" + image.Config.Plugin.Version
		}
		org := image.Config.Plugin.Organization
		name := image.Config.Plugin.Name
		docURL := ""
		if org == "turbot" {
			docURL = fmt.Sprintf("https://hub.steampipe.io/plugins/%s/%s", org, name)
		}
		installReports = append(installReports, display.InstallReport{
			Skipped:        false,
			Plugin:         p,
			DocURL:         docURL,
			Version:        versionString,
			IsUpdateReport: false,
		})
	}

	display.StopSpinner(spinner)

	refreshConnectionsIfNecessary(installReports, false)
	display.PrintInstallReports(installReports, false)

	// a concluding blank line - since we always output multiple lines
	fmt.Println()
}

func runPluginUpdateCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runPluginUpdateCmd install")
	defer func() {
		utils.LogTime("runPluginUpdateCmd end")
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

	var runUpdatesFor []*versionfile.InstalledVersion
	updateReports := make([]display.InstallReport, 0, len(plugins))

	// a leading blank line - since we always output multiple lines
	fmt.Println()

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
				updateReports = append(updateReports, display.InstallReport{
					Skipped:        true,
					Plugin:         p,
					SkipReason:     display.NOT_INSTALLED,
					IsUpdateReport: true,
				})
			}
		}
	}

	if len(plugins) == len(updateReports) {
		// we have report for all
		// this may happen if all given plugins are
		// not installed
		display.PrintInstallReports(updateReports, true)
		fmt.Println()
		return
	}

	spinner := display.ShowSpinner("Checking for available updates")
	reports := plugin.GetUpdateReport(state.InstallationID, runUpdatesFor)
	display.StopSpinner(spinner)

	if len(reports) == 0 {
		// this happens if for some reason the update server could not be contacted,
		// in which case we get back an empty map
		utils.ShowError(fmt.Errorf("there was an issue contacting the update server. Please try later"))
		return
	}

	for _, report := range reports {
		if report.Plugin.ImageDigest == report.CheckResponse.Digest {
			updateReports = append(updateReports, display.InstallReport{
				Plugin:         fmt.Sprintf("%s@%s", report.CheckResponse.Name, report.CheckResponse.Stream),
				Skipped:        true,
				SkipReason:     display.LATEST_ALREADY_INSTALLED,
				IsUpdateReport: true,
			})
			continue
		}

		spinner := display.ShowSpinner(fmt.Sprintf("Updating plugin %s...", report.CheckResponse.Name))
		image, err := plugin.Install(report.Plugin.Name)
		display.StopSpinner(spinner)
		if err != nil {
			msg := ""
			if strings.HasSuffix(err.Error(), "not found") {
				msg = "Not found"
			} else {
				msg = err.Error()
			}
			updateReports = append(updateReports, display.InstallReport{
				Plugin:         fmt.Sprintf("%s@%s", report.CheckResponse.Name, report.CheckResponse.Stream),
				Skipped:        true,
				SkipReason:     msg,
				IsUpdateReport: true,
			})
			continue
		}

		versionString := ""
		if image.Config.Plugin.Version != "" {
			versionString = " v" + image.Config.Plugin.Version
		}
		org := image.Config.Plugin.Organization
		name := image.Config.Plugin.Name
		docURL := ""
		if org == "turbot" {
			docURL = fmt.Sprintf("https://hub.steampipe.io/plugins/%s/%s", org, name)
		}
		updateReports = append(updateReports, display.InstallReport{
			Plugin:         fmt.Sprintf("%s@%s", report.CheckResponse.Name, report.CheckResponse.Stream),
			Skipped:        false,
			Version:        versionString,
			DocURL:         docURL,
			IsUpdateReport: true,
		})
	}

	refreshConnectionsIfNecessary(updateReports, true)
	display.PrintInstallReports(updateReports, true)

	// a concluding blank line - since we always output multiple lines
	fmt.Println()
}

// start service if necessary and refresh connections
func refreshConnectionsIfNecessary(reports []display.InstallReport, isUpdate bool) error {
	// get count of skipped reports
	skipped := 0
	for _, report := range reports {
		if report.Skipped {
			skipped++
		}
	}
	if skipped == len(reports) {
		// if all were skipped,
		// no point continuing
		return nil
	}

	// reload the config, since an installation MUST have created a new config file
	if !isUpdate {
		var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
		config, err := steampipeconfig.LoadSteampipeConfig(viper.GetString(constants.ArgWorkspace), cmd.Name())
		if err != nil {
			return err
		}
		steampipeconfig.Config = config
	}

	// todo move this into db package
	db.EnsureDBInstalled()
	status, err := db.GetStatus()
	if err != nil {
		return errors.New("could not retrieve service status")
	}

	var client *db.Client
	if status == nil {
		// the db service is not started - start it
		db.StartService(db.InvokerPlugin)
		defer func() { db.Shutdown(client, db.InvokerPlugin) }()
	}

	// TODO i think we can pass true here and not refresh below
	client, err = db.NewClient(false)
	if err != nil {
		return err
	}

	// refresh connections
	if _, err = client.RefreshConnections(); err != nil {
		return err
	}

	return nil
}

func runPluginListCmd(*cobra.Command, []string) {
	utils.LogTime("runPluginListCmd list")
	defer func() {
		utils.LogTime("runPluginListCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	connectionMap, err := getPluginConnectionMap()
	if err != nil {
		utils.ShowErrorWithMessage(err, "Plugin Listing failed")
		return
	}

	list, err := plugin.List(connectionMap)
	if err != nil {
		utils.ShowErrorWithMessage(err, "Plugin Listing failed")
	}
	headers := []string{"Name", "Version", "Connections"}
	rows := [][]string{}
	for _, item := range list {
		rows = append(rows, []string{item.Name, item.Version, strings.Join(item.Connections, ",")})
	}
	display.ShowWrappedTable(headers, rows, false)
}

func runPluginUninstallCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runPluginUninstallCmd uninstall")

	defer func() {
		utils.LogTime("runPluginUninstallCmd end")
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

	var client *db.Client
	if status == nil {
		// the db service is not started - start it
		db.StartService(db.InvokerPlugin)
		defer func() { db.Shutdown(client, db.InvokerPlugin) }()
	}

	client, err = db.NewClient(true)
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
