package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
	"github.com/turbot/steampipe/pkg/plugin"
	"github.com/turbot/steampipe/pkg/statefile"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

// Plugin management commands
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

var pluginInstallSteps = []string{
	"Downloading",
	"Installing Plugin",
	"Installing Docs",
	"Installing Config",
	"Updating Steampipe",
	"Done",
}

func runPluginInstallCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("runPluginInstallCmd install")
	defer func() {
		utils.LogTime("runPluginInstallCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	// args to 'plugin install' -- one or more plugins to install
	// plugin names can be simple names ('aws') for "standard" plugins,
	// or full refs to the OCI image (us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0)
	plugins := append([]string{}, args...)
	installReports := make(display.PluginInstallReports, 0, len(plugins))

	if len(plugins) == 0 {
		fmt.Println()
		error_helpers.ShowError(ctx, fmt.Errorf("you need to provide at least one plugin to install"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		exitCode = constants.ExitCodeInsufficientOrWrongArguments
		return
	}

	// a leading blank line - since we always output multiple lines
	fmt.Println()

	statusSpinner := statushooks.NewStatusSpinner()
	progressBars := uiprogress.New()
	installWaitGroup := &sync.WaitGroup{}
	dataChannel := make(chan *display.PluginInstallReport, len(plugins))

	progressBars.Start()

	for _, pluginName := range plugins {
		installWaitGroup.Add(1)
		bar := createProgressBar(pluginName, progressBars)
		go doPluginInstall(ctx, bar, pluginName, installWaitGroup, dataChannel)
	}
	go func() {
		installWaitGroup.Wait()
		close(dataChannel)
	}()
	for report := range dataChannel {
		installReports = append(installReports, report)
	}

	progressBars.Stop()
	statusSpinner.UpdateSpinnerMessage("Refreshing connections...")
	refreshConnectionsIfNecessary(ctx, installReports, true)
	statusSpinner.Done()
	display.PrintInstallReports(installReports, false)

	// a concluding blank line - since we always output multiple lines
	fmt.Println()
}

func doPluginInstall(ctx context.Context, bar *uiprogress.Bar, pluginName string, wg *sync.WaitGroup, returnChannel chan *display.PluginInstallReport) {
	var report *display.PluginInstallReport

	pluginAlreadyInstalled, _ := plugin.Exists(pluginName)
	if pluginAlreadyInstalled {
		// set the bar to MAX
		bar.Set(len(pluginInstallSteps))
		// let the bar append itself with "Already Installed"
		bar.AppendFunc(func(b *uiprogress.Bar) string {
			return utils.Resize(constants.PluginAlreadyInstalled, 20)
		})
		report = &display.PluginInstallReport{
			Plugin:         pluginName,
			Skipped:        true,
			SkipReason:     constants.PluginAlreadyInstalled,
			IsUpdateReport: false,
		}
	} else {
		// let the bar append itself with the current installation step
		bar.AppendFunc(func(b *uiprogress.Bar) string {
			if report != nil && report.SkipReason == constants.PluginNotFound {
				return utils.Resize(constants.PluginNotFound, 20)
			} else {
				if b.Current() == 0 {
					// no install step to display yet
					return ""
				}
				return utils.Resize(pluginInstallSteps[b.Current()-1], 20)
			}
		})
		report = installPlugin(ctx, pluginName, false, bar)
	}
	returnChannel <- report
	wg.Done()
}

func runPluginUpdateCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("runPluginUpdateCmd start")
	defer func() {
		utils.LogTime("runPluginUpdateCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	// args to 'plugin update' -- one or more plugins to update
	// These can be simple names ('aws') for "standard" plugins,
	// or full refs to the OCI image (us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0)
	plugins, err := resolveUpdatePluginsFromArgs(args)
	if err != nil {
		fmt.Println()
		error_helpers.ShowError(ctx, err)
		fmt.Println()
		cmd.Help()
		fmt.Println()
		exitCode = constants.ExitCodeInsufficientOrWrongArguments
		return
	}

	if len(plugins) > 0 && !(cmdconfig.Viper().GetBool("all")) && plugins[0] == "all" {
		// improve the response to wrong argument "steampipe plugin update all"
		fmt.Println()
		error_helpers.ShowError(ctx, fmt.Errorf("Did you mean %s?", constants.Bold("--all")))
		fmt.Println()
		return
	}

	state, err := statefile.LoadState()
	if err != nil {
		error_helpers.ShowError(ctx, fmt.Errorf("could not load state"))
		exitCode = constants.ExitCodeLoadingError
		return
	}

	// load up the version file data
	versionData, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		error_helpers.ShowError(ctx, fmt.Errorf("error loading current plugin data"))
		exitCode = constants.ExitCodeLoadingError
		return
	}

	var runUpdatesFor []*versionfile.InstalledVersion
	updateResults := make(display.PluginInstallReports, 0, len(plugins))

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
				updateResults = append(updateResults, &display.PluginInstallReport{
					Skipped:        true,
					Plugin:         p,
					SkipReason:     constants.PluginNotInstalled,
					IsUpdateReport: true,
				})
			}
		}
	}

	if len(plugins) == len(updateResults) {
		// we have report for all
		// this may happen if all given plugins are
		// not installed
		display.PrintInstallReports(updateResults, true)
		fmt.Println()
		return
	}
	statusSpinner := statushooks.NewStatusSpinner(statushooks.WithMessage("Checking for available updates"))
	// long timeout - we are happy to wait
	timeout := 30 * time.Second
	reports := plugin.GetUpdateReport(state.InstallationID, runUpdatesFor, timeout)
	statusSpinner.Done()

	if len(reports) == 0 {
		// this happens if for some reason the update server could not be contacted,
		// in which case we get back an empty map
		error_helpers.ShowError(ctx, fmt.Errorf("there was an issue contacting the update server. Please try later."))
		exitCode = constants.ExitCodeLoadingError
		return
	}

	updateWaitGroup := &sync.WaitGroup{}
	dataChannel := make(chan *display.PluginInstallReport, len(reports))
	progressBars := uiprogress.New()
	progressBars.Start()

	sorted := utils.SortedMapKeys(reports)
	for _, key := range sorted {
		report := reports[key]
		updateWaitGroup.Add(1)
		bar := createProgressBar(report.ShortName(), progressBars)
		go doPluginUpdate(ctx, bar, report, updateWaitGroup, dataChannel)
	}
	go func() {
		updateWaitGroup.Wait()
		close(dataChannel)
	}()
	for updateResult := range dataChannel {
		updateResults = append(updateResults, updateResult)
	}
	// prevent fdw update message
	nullStatusHookCtx := statushooks.DisableStatusHooks(ctx)
	refreshConnectionsIfNecessary(nullStatusHookCtx, updateResults, false)
	progressBars.Stop()
	fmt.Println()
	display.PrintInstallReports(updateResults, true)

	// a concluding blank line - since we always output multiple lines
	fmt.Println()
}

func doPluginUpdate(ctx context.Context, bar *uiprogress.Bar, pvr plugin.VersionCheckReport, wg *sync.WaitGroup, returnChannel chan *display.PluginInstallReport) {
	var report *display.PluginInstallReport

	if skip, skipReason := plugin.SkipUpdate(pvr); skip {
		bar.AppendFunc(func(b *uiprogress.Bar) string {
			// set the progress bar to append itself with "Already Installed"
			return utils.Resize(skipReason, 30)
		})
		// set the progress bar to the maximum
		bar.Set(len(pluginInstallSteps))
		report = &display.PluginInstallReport{
			Plugin:         fmt.Sprintf("%s@%s", pvr.CheckResponse.Name, pvr.CheckResponse.Stream),
			Skipped:        true,
			SkipReason:     skipReason,
			IsUpdateReport: true,
		}
	} else {
		bar.AppendFunc(func(b *uiprogress.Bar) string {
			// set the progress bar to append itself  with the step underway
			if b.Current() == 0 {
				// no install step to display yet
				return ""
			}
			return utils.Resize(pluginInstallSteps[b.Current()-1], 20)
		})
		report = installPlugin(ctx, pvr.Plugin.Name, true, bar)
	}
	returnChannel <- report
	wg.Done()
}

func createProgressBar(plugin string, parentProgressBars *uiprogress.Progress) *uiprogress.Bar {
	bar := parentProgressBars.AddBar(len(pluginInstallSteps))
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return utils.Resize(plugin, 20)
	})
	return bar
}

func installPlugin(ctx context.Context, pluginName string, isUpdate bool, bar *uiprogress.Bar) *display.PluginInstallReport {
	// start a channel for progress publications from plugin.Install
	progress := make(chan struct{}, 5)
	defer func() {
		// close the progress channel
		close(progress)
	}()
	go func() {
		for {
			// wait for a message on the progress channel
			<-progress
			// increment the progress bar
			bar.Incr()
		}
	}()

	image, err := plugin.Install(ctx, pluginName, progress)
	if err != nil {
		msg := ""
		_, name, stream := ociinstaller.NewSteampipeImageRef(pluginName).GetOrgNameAndStream()
		if isPluginNotFoundErr(err) {
			msg = constants.PluginNotFound
		} else {
			msg = err.Error()
		}
		return &display.PluginInstallReport{
			Plugin:         fmt.Sprintf("%s@%s", name, stream),
			Skipped:        true,
			SkipReason:     msg,
			IsUpdateReport: isUpdate,
		}
	}

	org, name, stream := image.ImageRef.GetOrgNameAndStream()
	versionString := ""
	if image.Config.Plugin.Version != "" {
		versionString = " v" + image.Config.Plugin.Version
	}
	docURL := fmt.Sprintf("https://hub.steampipe.io/plugins/%s/%s", org, name)
	return &display.PluginInstallReport{
		Plugin:         fmt.Sprintf("%s@%s", name, stream),
		Skipped:        false,
		Version:        versionString,
		DocURL:         docURL,
		IsUpdateReport: isUpdate,
	}
}

func isPluginNotFoundErr(err error) bool {
	return strings.HasSuffix(err.Error(), "not found")
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
func refreshConnectionsIfNecessary(ctx context.Context, reports display.PluginInstallReports, shouldReload bool) error {
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

	client, err := db_local.GetLocalClient(ctx, constants.InvokerPlugin, nil)
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
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	pluginConnectionMap, missingPluginMap, err := getPluginConnectionMap(cmd.Context())
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "Plugin Listing failed")
		exitCode = constants.ExitCodePluginListFailure
		return
	}
	log.Printf("[TRACE] missing plugins: %v", missingPluginMap)

	list, err := plugin.List(pluginConnectionMap)
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "Plugin Listing failed")
		exitCode = constants.ExitCodePluginListFailure
	}

	// If there are missing plugins which have connections left over, list them
	// along with installed plugins
	if len(missingPluginMap) != 0 {
		// List installed plugins
		if len(list) != 0 {
			headers := []string{"Installed Plugin", "Version", "Connections"}
			rows := [][]string{}
			for _, item := range list {
				rows = append(rows, []string{item.Name, item.Version, strings.Join(item.Connections, ",")})
			}
			display.ShowWrappedTable(headers, rows, false)
			fmt.Printf("\n")
		}

		// List missing plugins
		headers := []string{"Missing Plugin", "Connections"}
		conns := []string{}
		missingRows := [][]string{}
		for p, item := range missingPluginMap {
			for _, conn := range item {
				conns = append(conns, conn.Name)
			}
			missingRows = append(missingRows, []string{p, strings.Join(conns, ",")})
			conns = []string{}
		}
		display.ShowWrappedTable(headers, missingRows, false)
	} else {
		headers := []string{"Installed Plugin", "Version", "Connections"}
		rows := [][]string{}
		for _, item := range list {
			rows = append(rows, []string{item.Name, item.Version, strings.Join(item.Connections, ",")})
		}
		display.ShowWrappedTable(headers, rows, false)
	}
}

func runPluginUninstallCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("runPluginUninstallCmd uninstall")

	defer func() {
		utils.LogTime("runPluginUninstallCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	if len(args) == 0 {
		fmt.Println()
		error_helpers.ShowError(ctx, fmt.Errorf("you need to provide at least one plugin to uninstall"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		exitCode = constants.ExitCodeInsufficientOrWrongArguments
		return
	}

	connectionMap, _, err := getPluginConnectionMap(ctx)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		exitCode = constants.ExitCodePluginListFailure
		return
	}

	reports := display.PluginRemoveReports{}
	spinner := statushooks.NewStatusSpinner(statushooks.WithMessage(fmt.Sprintf("Uninstalling %s", utils.Pluralize("plugin", len(args)))))
	for _, p := range args {
		spinner.SetStatus(fmt.Sprintf("Uninstalling %s", p))
		if report, err := plugin.Remove(ctx, p, connectionMap); err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, fmt.Sprintf("Failed to uninstall plugin '%s'", p))
		} else {
			report.ShortName = p
			reports = append(reports, *report)
		}
	}
	spinner.Done()
	reports.Print()
}

// returns a map of pluginFullName -> []{connections using pluginFullName}
func getPluginConnectionMap(ctx context.Context) (map[string][]modconfig.Connection, map[string][]modconfig.Connection, error) {
	client, err := db_local.GetLocalClient(ctx, constants.InvokerPlugin, nil)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close(ctx)
	res := client.RefreshConnectionAndSearchPaths(ctx)
	if res.Error != nil {
		return nil, nil, res.Error
	}
	// display any initialisation warnings
	if len(res.Warnings) > 0 {
		res.ShowWarnings()
		fmt.Printf("\n")
	}

	missingPlugins := res.Updates.MissingPlugins

	pluginConnectionMap := make(map[string][]modconfig.Connection)
	// missingPluginConnectionMap := make(map[string][]modconfig.Connection)

	for _, v := range *client.ConnectionMap() {
		_, found := pluginConnectionMap[v.Plugin]
		if !found {
			pluginConnectionMap[v.Plugin] = []modconfig.Connection{}
		}
		pluginConnectionMap[v.Plugin] = append(pluginConnectionMap[v.Plugin], *v.Connection)
	}

	if missingPlugins != nil {
		return pluginConnectionMap, missingPlugins, nil
	}
	return pluginConnectionMap, nil, nil
}
