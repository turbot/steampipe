package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/ociinstaller/versionfile"
	"github.com/turbot/steampipe/plugin"
	"github.com/turbot/steampipe/statefile"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

//  Plugin management commands
func pluginCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "plugin [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe plugin management",
		Long: `Steampipe plugin management.

Plugins extend Steampipe to work with many different services and providers.
Find plugins using the public registry at https://hub.steampipe.io.

Examples:

  # Install a plugin
  steampipe plugin install aws

  # Update a plugin
  steampipe plugin update aws

  # List installed plugins
  steampipe plugin list

  # Uninstall a plugin
  steampipe plugin uninstall aws`,
	}

	cmd.AddCommand(pluginInstallCmd())
	cmd.AddCommand(pluginListCmd())
	cmd.AddCommand(pluginUninstallCmd())
	cmd.AddCommand(pluginUpdateCmd())
	cmd.Flags().BoolP(constants.ArgHelp, "h", false, "Help for plugin")

	return cmd
}

// Install a plugin
func pluginInstallCmd() *cobra.Command {
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
		OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for plugin install")
	return cmd
}

// Update plugins
func pluginUpdateCmd() *cobra.Command {
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
		AddBoolFlag(constants.ArgAll, "", false, "Update all plugins to its latest available version").
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for plugin update")

	return cmd
}

// List plugins
func pluginListCmd() *cobra.Command {
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
		AddBoolFlag("outdated", "", false, "Check each plugin in the list for updates").
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for plugin list")

	return cmd
}

// Uninstall a plugin
func pluginUninstallCmd() *cobra.Command {
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

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for plugin uninstall")

	return cmd
}

// exitCode=1 For unknown errors resulting in panics
// exitCode=2 For insufficient/wrong arguments passed in the command
// exitCode=3 For errors related to loading state, loading version data or an issue contacting the update server.
// exitCode=4 For plugin listing failures
func runPluginInstallCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("runPluginInstallCmd install")
	defer func() {
		utils.LogTime("runPluginInstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
			exitCode = 1
		}
	}()

	// args to 'plugin install' -- one or more plugins to install
	// plugin names can be simple names ('aws') for "standard" plugins,
	// or full refs to the OCI image (us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0)
	plugins := append([]string{}, args...)
	installReports := make([]display.InstallReport, 0, len(plugins))

	if len(plugins) == 0 {
		fmt.Println()
		utils.ShowError(ctx, fmt.Errorf("you need to provide at least one plugin to install"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		exitCode = 2
		return
	}

	// a leading blank line - since we always output multiple lines
	fmt.Println()

	statusSpinner := statushooks.NewStatusSpinner()

	for _, p := range plugins {
		isPluginExists, _ := plugin.Exists(p)
		if isPluginExists {
			installReports = append(installReports, display.InstallReport{
				Plugin:         p,
				Skipped:        true,
				SkipReason:     constants.PluginAlreadyInstalled,
				IsUpdateReport: false,
			})
			continue
		}
		statusSpinner.SetStatus(fmt.Sprintf("Installing plugin: %s", p))
		image, err := plugin.Install(cmd.Context(), p)
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
		docURL := fmt.Sprintf("https://hub.steampipe.io/plugins/%s/%s", org, name)
		installReports = append(installReports, display.InstallReport{
			Skipped:        false,
			Plugin:         p,
			DocURL:         docURL,
			Version:        versionString,
			IsUpdateReport: false,
		})
	}

	statusSpinner.Done()

	refreshConnectionsIfNecessary(cmd.Context(), installReports, true)
	display.PrintInstallReports(installReports, false)

	// a concluding blank line - since we always output multiple lines
	fmt.Println()
}

func runPluginUpdateCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("runPluginUpdateCmd install")
	defer func() {
		utils.LogTime("runPluginUpdateCmd end")
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
			exitCode = 1
		}
	}()

	// args to 'plugin update' -- one or more plugins to update
	// These can be simple names ('aws') for "standard" plugins,
	// or full refs to the OCI image (us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0)
	plugins, err := resolveUpdatePluginsFromArgs(args)
	if err != nil {
		fmt.Println()
		utils.ShowError(ctx, err)
		fmt.Println()
		cmd.Help()
		fmt.Println()
		exitCode = 2
		return
	}

	state, err := statefile.LoadState()
	if err != nil {
		utils.ShowError(ctx, fmt.Errorf("could not load state"))
		exitCode = 3
		return
	}

	// load up the version file data
	versionData, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		utils.ShowError(ctx, fmt.Errorf("error loading current plugin data"))
		exitCode = 3
		return
	}

	var runUpdatesFor []*versionfile.InstalledVersion
	updateReports := make([]display.InstallReport, 0, len(plugins))

	// a leading blank line - since we always output multiple lines
	fmt.Println()

	if cmdconfig.Viper().GetBool(constants.ArgAll) {
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
					SkipReason:     constants.PluginNotInstalled,
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

	statusSpinner := statushooks.NewStatusSpinner(statushooks.WithMessage("Checking for available updates"))
	reports := plugin.GetUpdateReport(state.InstallationID, runUpdatesFor)
	statusSpinner.Done()

	if len(reports) == 0 {
		// this happens if for some reason the update server could not be contacted,
		// in which case we get back an empty map
		utils.ShowError(ctx, fmt.Errorf("there was an issue contacting the update server. Please try later"))
		exitCode = 3
		return
	}

	for _, report := range reports {
		if report.Plugin.ImageDigest == report.CheckResponse.Digest {
			updateReports = append(updateReports, display.InstallReport{
				Plugin:         fmt.Sprintf("%s@%s", report.CheckResponse.Name, report.CheckResponse.Stream),
				Skipped:        true,
				SkipReason:     constants.PluginLatestAlreadyInstalled,
				IsUpdateReport: true,
			})
			continue
		}

		statusSpinner.SetStatus(fmt.Sprintf("Updating plugin %s...", report.CheckResponse.Name))
		image, err := plugin.Install(cmd.Context(), report.Plugin.Name)
		statusSpinner.Done()
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
		docURL := fmt.Sprintf("https://hub.steampipe.io/plugins/%s/%s", org, name)
		updateReports = append(updateReports, display.InstallReport{
			Plugin:         fmt.Sprintf("%s@%s", report.CheckResponse.Name, report.CheckResponse.Stream),
			Skipped:        false,
			Version:        versionString,
			DocURL:         docURL,
			IsUpdateReport: true,
		})
	}

	refreshConnectionsIfNecessary(cmd.Context(), updateReports, false)
	display.PrintInstallReports(updateReports, true)

	// a concluding blank line - since we always output multiple lines
	fmt.Println()
}

func resolveUpdatePluginsFromArgs(args []string) ([]string, error) {
	plugins := append([]string{}, args...)

	if len(plugins) == 0 && !(cmdconfig.Viper().GetBool("all")) {
		// either plugin name(s) or "all" must be provided
		return nil, fmt.Errorf("you need to provide at least one plugin to update or use the %s flag", constants.Bold("--all"))
	}

	if len(plugins) > 0 && cmdconfig.Viper().GetBool(constants.ArgAll) {
		// we can't allow update and install at the same time
		return nil, fmt.Errorf("%s cannot be used when updating specific plugins", constants.Bold("`--all`"))
	}
	return plugins, nil
}

// start service if necessary and refresh connections
func refreshConnectionsIfNecessary(ctx context.Context, reports []display.InstallReport, shouldReload bool) error {
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
	if shouldReload {
		var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
		config, err := steampipeconfig.LoadSteampipeConfig(viper.GetString(constants.ArgWorkspaceChDir), cmd.Name())
		if err != nil {
			return err
		}
		steampipeconfig.GlobalConfig = config
	}

	client, err := db_local.GetLocalClient(ctx, constants.InvokerPlugin)
	if err != nil {
		return err
	}
	defer client.Close(ctx)
	res := client.RefreshConnectionAndSearchPaths(ctx)
	if res.Error != nil {
		return res.Error
	}
	// display any initialisation warnings
	res.ShowWarnings()
	return nil
}

func runPluginListCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("runPluginListCmd list")
	defer func() {
		utils.LogTime("runPluginListCmd end")
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
			exitCode = 1
		}
	}()

	pluginConnectionMap, err := getPluginConnectionMap(cmd.Context())
	if err != nil {
		utils.ShowErrorWithMessage(ctx, err, "Plugin Listing failed")
		exitCode = 4
		return
	}

	list, err := plugin.List(pluginConnectionMap)
	if err != nil {
		utils.ShowErrorWithMessage(ctx, err, "Plugin Listing failed")
		exitCode = 4
	}
	headers := []string{"Name", "Version", "Connections"}
	rows := [][]string{}
	for _, item := range list {
		rows = append(rows, []string{item.Name, item.Version, strings.Join(item.Connections, ",")})
	}
	display.ShowWrappedTable(headers, rows, false)
}

func runPluginUninstallCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("runPluginUninstallCmd uninstall")

	defer func() {
		utils.LogTime("runPluginUninstallCmd end")
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
			exitCode = 1
		}
	}()

	if len(args) == 0 {
		fmt.Println()
		utils.ShowError(ctx, fmt.Errorf("you need to provide at least one plugin to uninstall"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		exitCode = 2
		return
	}

	connectionMap, err := getPluginConnectionMap(ctx)
	if err != nil {
		utils.ShowError(ctx, err)
		exitCode = 4
		return
	}

	for _, p := range args {
		if err := plugin.Remove(ctx, p, connectionMap); err != nil {
			utils.ShowErrorWithMessage(ctx, err, fmt.Sprintf("Failed to uninstall plugin '%s'", p))
		}
	}
}

// returns a map of pluginFullName -> []{connections using pluginFullName}
func getPluginConnectionMap(ctx context.Context) (map[string][]modconfig.Connection, error) {
	client, err := db_local.GetLocalClient(ctx, constants.InvokerPlugin)
	if err != nil {
		return nil, err
	}
	defer client.Close(ctx)
	res := client.RefreshConnectionAndSearchPaths(ctx)
	if res.Error != nil {
		return nil, res.Error
	}
	// display any initialisation warnings
	res.ShowWarnings()

	pluginConnectionMap := make(map[string][]modconfig.Connection)

	for _, v := range *client.ConnectionMap() {
		_, found := pluginConnectionMap[v.Plugin]
		if !found {
			pluginConnectionMap[v.Plugin] = []modconfig.Connection{}
		}
		pluginConnectionMap[v.Plugin] = append(pluginConnectionMap[v.Plugin], *v.Connection)
	}
	return pluginConnectionMap, nil
}
