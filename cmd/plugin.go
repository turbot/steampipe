package cmd

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/contexthelpers"
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
		AddBoolFlag(constants.ArgHelp, false, "Help for plugin install", cmdconfig.FlagOptions.WithShortHand("h"))
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
		AddBoolFlag(constants.ArgAll, false, "Update all plugins to its latest available version").
		AddBoolFlag(constants.ArgHelp, false, "Help for plugin update", cmdconfig.FlagOptions.WithShortHand("h"))

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
		AddBoolFlag("outdated", false, "Check each plugin in the list for updates").
		AddBoolFlag(constants.ArgHelp, false, "Help for plugin list", cmdconfig.FlagOptions.WithShortHand("h"))

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
		AddBoolFlag(constants.ArgHelp, false, "Help for plugin uninstall", cmdconfig.FlagOptions.WithShortHand("h"))

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
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return
	}

	// a leading blank line - since we always output multiple lines
	fmt.Println()

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

	installCount := 0
	for report := range dataChannel {
		installReports = append(installReports, report)
		if !report.Skipped {
			installCount++
		}
	}

	progressBars.Stop()

	if installCount > 0 {
		// TODO do we need to refresh connections here

		// reload the config, since an installation should have created a new config file
		var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
		config, errorsAndWarnings := steampipeconfig.LoadSteampipeConfig(viper.GetString(constants.ArgModLocation), cmd.Name())
		if errorsAndWarnings.GetError() != nil {
			error_helpers.ShowWarning(fmt.Sprintf("Failed to reload config - install report may be incomplete (%s)", errorsAndWarnings.GetError()))
		} else {
			steampipeconfig.GlobalConfig = config
		}

		statushooks.Done(ctx)
	}
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
			return helpers.Resize(constants.PluginAlreadyInstalled, 20)
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
				return helpers.Resize(constants.PluginNotFound, 20)
			} else {
				if b.Current() == 0 {
					// no install step to display yet
					return ""
				}
				return helpers.Resize(pluginInstallSteps[b.Current()-1], 20)
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
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return
	}

	if len(plugins) > 0 && !(cmdconfig.Viper().GetBool("all")) && plugins[0] == "all" {
		// improve the response to wrong argument "steampipe plugin update all"
		fmt.Println()
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		error_helpers.ShowError(ctx, fmt.Errorf("Did you mean %s?", constants.Bold("--all")))
		fmt.Println()
		return
	}

	state, err := statefile.LoadState()
	if err != nil {
		error_helpers.ShowError(ctx, fmt.Errorf("could not load state"))
		exitCode = constants.ExitCodePluginLoadingError
		return
	}

	// load up the version file data
	versionData, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		error_helpers.ShowError(ctx, fmt.Errorf("error loading current plugin data"))
		exitCode = constants.ExitCodePluginLoadingError
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
				exitCode = constants.ExitCodePluginNotFound
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

	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	statushooks.SetStatus(ctx, "Checking for available updates")
	reports := plugin.GetUpdateReport(timeoutCtx, state.InstallationID, runUpdatesFor)
	statushooks.Done(ctx)
	if len(reports) == 0 {
		// this happens if for some reason the update server could not be contacted,
		// in which case we get back an empty map
		error_helpers.ShowError(ctx, fmt.Errorf("there was an issue contacting the update server, please try later"))
		exitCode = constants.ExitCodePluginLoadingError
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
	installCount := 0

	for updateResult := range dataChannel {
		updateResults = append(updateResults, updateResult)
		if !updateResult.Skipped {
			installCount++
		}
	}
	progressBars.Stop()

	display.PrintInstallReports(updateResults, true)

	// a concluding blank line - since we always output multiple lines
	fmt.Println()
}

func doPluginUpdate(ctx context.Context, bar *uiprogress.Bar, pvr plugin.VersionCheckReport, wg *sync.WaitGroup, returnChannel chan *display.PluginInstallReport) {
	var report *display.PluginInstallReport

	if skip, skipReason := plugin.SkipUpdate(pvr); skip {
		bar.AppendFunc(func(b *uiprogress.Bar) string {
			// set the progress bar to append itself with "Already Installed"
			return helpers.Resize(skipReason, 30)
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
			return helpers.Resize(pluginInstallSteps[b.Current()-1], 20)
		})
		report = installPlugin(ctx, pvr.Plugin.Name, true, bar)
	}
	returnChannel <- report
	wg.Done()
}

func createProgressBar(plugin string, parentProgressBars *uiprogress.Progress) *uiprogress.Bar {
	bar := parentProgressBars.AddBar(len(pluginInstallSteps))
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return helpers.Resize(plugin, 20)
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
			exitCode = constants.ExitCodePluginNotFound
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

func runPluginListCmd(cmd *cobra.Command, args []string) {
	// setup a cancel context and start cancel handler
	ctx, cancel := context.WithCancel(cmd.Context())
	contexthelpers.StartCancelHandler(cancel)

	utils.LogTime("runPluginListCmd list")
	defer func() {
		utils.LogTime("runPluginListCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	pluginList, failedPluginMap, missingPluginMap, res := getPluginList(ctx)
	if res.Error != nil {
		error_helpers.ShowErrorWithMessage(ctx, res.Error, "plugin listing failed")
		exitCode = constants.ExitCodePluginListFailure
		return
	}

	// List installed plugins in a table
	if len(pluginList) != 0 {
		headers := []string{"Installed Plugin", "Version", "Connections"}
		var rows [][]string
		for _, item := range pluginList {
			rows = append(rows, []string{item.Name, item.Version, strings.Join(item.Connections, ",")})
		}
		display.ShowWrappedTable(headers, rows, &display.ShowWrappedTableOptions{AutoMerge: false})
		fmt.Printf("\n")
	}

	// List failed/missing plugins in a separate table
	if len(failedPluginMap) != 0 || len(missingPluginMap) != 0 {
		// List missing/failed plugins
		headers := []string{"Failed Plugin", "Connections", "Reason"}
		var conns = []string{}
		var missingRows [][]string
		// failed plugins
		for p, item := range failedPluginMap {
			for _, conn := range item {
				conns = append(conns, conn.Name)
			}
			missingRows = append(missingRows, []string{p, strings.Join(conns, ","), "plugin failed to start"})
			conns = []string{}
		}
		// missing plugins
		for p, item := range missingPluginMap {
			for _, conn := range item {
				conns = append(conns, conn.Name)
			}
			missingRows = append(missingRows, []string{p, strings.Join(conns, ","), "plugin not installed"})
			conns = []string{}
		}
		display.ShowWrappedTable(headers, missingRows, &display.ShowWrappedTableOptions{AutoMerge: false})
	}

	if len(res.Warnings) > 0 {
		fmt.Println()
		res.ShowWarnings()
		fmt.Printf("\n")
	}
}

func getPluginList(ctx context.Context) (pluginList []plugin.PluginListItem, failedPluginMap, missingPluginMap map[string][]*modconfig.Connection, res *modconfig.ErrorAndWarnings) {
	statushooks.Show(ctx)
	defer statushooks.Done(ctx)

	// get the maps of available and failed/missing plugins
	pluginConnectionMap, failedPluginMap, missingPluginMap, res := getPluginConnectionMap(ctx)
	if res.Error != nil {
		return nil, nil, nil, res
	}

	// TODO do we really need to look at installed plugins - can't we just use the plugin connection map
	// get a list of the installed plugins by inspecting the install location
	// pass pluginConnectionMap so we can populate the connections for each plugin
	pluginList, err := plugin.List(pluginConnectionMap)
	if err != nil {
		res.Error = err
		return nil, nil, nil, res
	}

	// remove the failed plugins from `list` since we don't want them in the installed table
	for pluginName := range failedPluginMap {
		for i := 0; i < len(pluginList); i++ {
			if pluginList[i].Name == pluginName {
				pluginList = append(pluginList[:i], pluginList[i+1:]...)
				i-- // Decrement the loop index since we just removed an element
			}
		}
	}
	return pluginList, failedPluginMap, missingPluginMap, res
}

func runPluginUninstallCmd(cmd *cobra.Command, args []string) {
	// setup a cancel context and start cancel handler
	ctx, cancel := context.WithCancel(cmd.Context())
	contexthelpers.StartCancelHandler(cancel)

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
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return
	}

	connectionMap, _, _, res := getPluginConnectionMap(ctx)
	if res.Error != nil {
		error_helpers.ShowError(ctx, res.Error)
		exitCode = constants.ExitCodePluginListFailure
		return
	}

	reports := display.PluginRemoveReports{}
	statushooks.SetStatus(ctx, fmt.Sprintf("Uninstalling %s", utils.Pluralize("plugin", len(args))))
	for _, p := range args {
		statushooks.SetStatus(ctx, fmt.Sprintf("Uninstalling %s", p))
		if report, err := plugin.Remove(ctx, p, connectionMap); err != nil {
			if strings.Contains(err.Error(), "not found") {
				exitCode = constants.ExitCodePluginNotFound
			}
			error_helpers.ShowErrorWithMessage(ctx, err, fmt.Sprintf("Failed to uninstall plugin '%s'", p))
		} else {
			report.ShortName = p
			reports = append(reports, *report)
		}
	}
	statushooks.Done(ctx)
	reports.Print()
}

func getPluginConnectionMap(ctx context.Context) (pluginConnectionMap, failedPluginMap, missingPluginMap map[string][]*modconfig.Connection, res *modconfig.ErrorAndWarnings) {
	statushooks.SetStatus(ctx, "Fetching connection map")

	res = &modconfig.ErrorAndWarnings{}
	// NOTE: start db if necessary - this will call refresh connections
	if err := db_local.EnsureDBInstalled(ctx); err != nil {
		return nil, nil, nil, modconfig.NewErrorsAndWarning(err)
	}
	startResult := db_local.StartServices(ctx, viper.GetInt(constants.ArgDatabasePort), db_local.ListenTypeLocal, constants.InvokerPlugin)
	if startResult.Error != nil {
		return nil, nil, nil, &startResult.ErrorAndWarnings
	}
	defer db_local.ShutdownService(ctx, constants.InvokerPlugin)

	// load the connection state and cache it!
	// passing nil so that we dont prune connections(connectionMap is now the connection state
	// containing all connections)
	connectionMap, _, err := steampipeconfig.GetConnectionState(nil)
	if err != nil {
		res.Error = err
		return nil, nil, nil, res
	}

	// create the map of failed/missing plugins
	failedPluginMap = map[string][]*modconfig.Connection{}
	missingPluginMap = map[string][]*modconfig.Connection{}
	for _, j := range connectionMap {
		if !j.Loaded && j.Error == "plugin failed to start" {
			if _, ok := failedPluginMap[j.Plugin]; !ok {
				failedPluginMap[j.Plugin] = []*modconfig.Connection{}
			}
			failedPluginMap[j.Plugin] = append(failedPluginMap[j.Plugin], j.Connection)
		} else if !j.Loaded && j.Error == "plugin not installed" {
			if _, ok := missingPluginMap[j.Plugin]; !ok {
				missingPluginMap[j.Plugin] = []*modconfig.Connection{}
			}
			missingPluginMap[j.Plugin] = append(missingPluginMap[j.Plugin], j.Connection)
		}
	}

	// create the map of available/loaded plugins
	pluginConnectionMap = make(map[string][]*modconfig.Connection)
	for _, v := range connectionMap {
		_, found := pluginConnectionMap[v.Plugin]
		if !found {
			pluginConnectionMap[v.Plugin] = []*modconfig.Connection{}
		}
		pluginConnectionMap[v.Plugin] = append(pluginConnectionMap[v.Plugin], v.Connection)
	}

	return pluginConnectionMap, failedPluginMap, missingPluginMap, res
}
