package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/contexthelpers"
	perror_helpers "github.com/turbot/pipe-fittings/v2/error_helpers"
	putils "github.com/turbot/pipe-fittings/v2/ociinstaller"
	pplugin "github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/querydisplay"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/pipe-fittings/v2/versionfile"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/installationstate"
	"github.com/turbot/steampipe/v2/pkg/ociinstaller"
	"github.com/turbot/steampipe/v2/pkg/plugin"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

type installedPlugin struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Connections []string `json:"connections"`
}

type failedPlugin struct {
	Name        string   `json:"name"`
	Reason      string   `json:"reason"`
	Connections []string `json:"connections"`
}

type pluginJsonOutput struct {
	Installed []installedPlugin `json:"installed"`
	Failed    []failedPlugin    `json:"failed"`
	Warnings  []string          `json:"warnings"`
}

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
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			utils.LogTime("cmd.plugin.PersistentPostRun start")
			defer utils.LogTime("cmd.plugin.PersistentPostRun end")
			pplugin.CleanupOldTmpDirs(cmd.Context())
		},
	}
	cmd.AddCommand(pluginInstallCmd())
	cmd.AddCommand(pluginListCmd())
	cmd.AddCommand(pluginUninstallCmd())
	cmd.AddCommand(pluginUpdateCmd())
	cmd.Flags().BoolP(pconstants.ArgHelp, "h", false, "Help for plugin")

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

  # Install all missing plugins that are specified in configuration files
  steampipe plugin install

  # Install a common plugin (turbot/aws)
  steampipe plugin install aws

  # Install a specific plugin version
  steampipe plugin install turbot/azure@0.1.0

  # Hide progress bars during installation
  steampipe plugin install --progress=false aws

  # Skip creation of default plugin config file
  steampipe plugin install --skip-config aws`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(pconstants.ArgProgress, true, "Display installation progress").
		AddBoolFlag(pconstants.ArgSkipConfig, false, "Skip creating the default config file for plugin").
		AddBoolFlag(pconstants.ArgHelp, false, "Help for plugin install", cmdconfig.FlagOptions.WithShortHand("h"))
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
  steampipe plugin update aws

  # Hide progress bars during update
  steampipe plugin update --progress=false aws`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(pconstants.ArgAll, false, "Update all plugins to its latest available version").
		AddBoolFlag(pconstants.ArgProgress, true, "Display installation progress").
		AddBoolFlag(pconstants.ArgHelp, false, "Help for plugin update", cmdconfig.FlagOptions.WithShortHand("h"))
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
  steampipe plugin list --outdated

  # List plugins output in json
  steampipe plugin list --output json`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag("outdated", false, "Check each plugin in the list for updates").
		AddStringFlag(pconstants.ArgOutput, "table", "Output format: table or json").
		AddBoolFlag(pconstants.ArgHelp, false, "Help for plugin list", cmdconfig.FlagOptions.WithShortHand("h"))
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
		AddBoolFlag(pconstants.ArgHelp, false, "Help for plugin uninstall", cmdconfig.FlagOptions.WithShortHand("h"))

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
	// plugin names can be simple names for "standard" plugins, constraint suffixed names
	// or full refs to the OCI image
	// - aws
	// - aws@0.118.0
	// - aws@^0.118
	// - ghcr.io/turbot/steampipe/plugins/turbot/aws:1.0.0
	plugins := append([]string{}, args...)
	showProgress := viper.GetBool(pconstants.ArgProgress)
	installReports := make(pplugin.PluginInstallReports, 0, len(plugins))

	if len(plugins) == 0 {
		if len(steampipeconfig.GlobalConfig.Plugins) == 0 {
			error_helpers.ShowError(ctx, sperr.New("No connections or plugins configured"))
			exitCode = constants.ExitCodeInsufficientOrWrongInputs
			return
		}

		// get the list of plugins to install
		for imageRef := range steampipeconfig.GlobalConfig.Plugins {
			ref := putils.NewImageRef(imageRef)
			plugins = append(plugins, ref.GetFriendlyName())
		}
	}

	state, err := installationstate.Load()
	if err != nil {
		error_helpers.ShowError(ctx, fmt.Errorf("could not load state"))
		exitCode = constants.ExitCodePluginLoadingError
		return
	}

	// a leading blank line - since we always output multiple lines
	fmt.Println()
	progressBars := uiprogress.New()
	installWaitGroup := &sync.WaitGroup{}
	reportChannel := make(chan *pplugin.PluginInstallReport, len(plugins))

	if showProgress {
		progressBars.Start()
	}
	for _, pluginName := range plugins {
		installWaitGroup.Add(1)
		bar := createProgressBar(pluginName, progressBars)

		ref := putils.NewImageRef(pluginName)
		org, name, constraint := ref.GetOrgNameAndStream()
		orgAndName := fmt.Sprintf("%s/%s", org, name)
		var resolved pplugin.ResolvedPluginVersion
		if ref.IsFromTurbotHub() {
			rpv, err := pplugin.GetLatestPluginVersionByConstraint(ctx, state.InstallationID, org, name, constraint)
			if err != nil || rpv == nil {
				report := &pplugin.PluginInstallReport{
					Plugin:         pluginName,
					Skipped:        true,
					SkipReason:     pconstants.InstallMessagePluginNotFound,
					IsUpdateReport: false,
				}
				reportChannel <- report
				installWaitGroup.Done()
				continue
			}
			resolved = *rpv
		} else {
			resolved = pplugin.NewResolvedPluginVersion(orgAndName, constraint, constraint)
		}

		go doPluginInstall(ctx, bar, pluginName, resolved, installWaitGroup, reportChannel)
	}
	go func() {
		installWaitGroup.Wait()
		close(reportChannel)
	}()
	installCount := 0
	for report := range reportChannel {
		installReports = append(installReports, report)
		if !report.Skipped {
			installCount++
		} else if !(report.Skipped && report.SkipReason == "Already installed") {
			exitCode = constants.ExitCodePluginInstallFailure
		}
	}
	if showProgress {
		progressBars.Stop()
	}

	if installCount > 0 {
		// TODO do we need to refresh connections here

		// reload the config, since an installation should have created a new config file
		var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
		config, errorsAndWarnings := steampipeconfig.LoadSteampipeConfig(ctx, viper.GetString(pconstants.ArgModLocation), cmd.Name())
		if errorsAndWarnings.GetError() != nil {
			error_helpers.ShowWarning(fmt.Sprintf("Failed to reload config - install report may be incomplete (%s)", errorsAndWarnings.GetError()))
		} else {
			steampipeconfig.GlobalConfig = config
		}

		statushooks.Done(ctx)
	}
	pplugin.PrintInstallReports(installReports, false)

	// a concluding blank line - since we always output multiple lines
	fmt.Println()
}

func doPluginInstall(ctx context.Context, bar *uiprogress.Bar, pluginName string, resolvedPlugin pplugin.ResolvedPluginVersion, wg *sync.WaitGroup, returnChannel chan *pplugin.PluginInstallReport) {
	var report *pplugin.PluginInstallReport

	pluginAlreadyInstalled, _ := pplugin.Exists(ctx, pluginName)
	if pluginAlreadyInstalled {
		// set the bar to MAX
		//nolint:golint,errcheck // the error happens if we set this over the max value
		bar.Set(len(pluginInstallSteps))
		// let the bar append itself with "Already Installed"
		bar.AppendFunc(func(b *uiprogress.Bar) string {
			return helpers.Resize(pconstants.InstallMessagePluginAlreadyInstalled, 20)
		})
		report = &pplugin.PluginInstallReport{
			Plugin:         pluginName,
			Skipped:        true,
			SkipReason:     pconstants.InstallMessagePluginAlreadyInstalled,
			IsUpdateReport: false,
		}
	} else {
		// let the bar append itself with the current installation step
		bar.AppendFunc(func(b *uiprogress.Bar) string {
			if report != nil && report.SkipReason == pconstants.InstallMessagePluginNotFound {
				return helpers.Resize(pconstants.InstallMessagePluginNotFound, 20)
			} else {
				if b.Current() == 0 {
					// no install step to display yet
					return ""
				}
				return helpers.Resize(pluginInstallSteps[b.Current()-1], 20)
			}
		})

		report = installPlugin(ctx, resolvedPlugin, false, bar)
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
	// These can be simple names for "standard" plugins, constraint suffixed names
	// or full refs to the OCI image
	// - aws
	// - aws@0.118.0
	// - aws@^0.118
	// - ghcr.io/turbot/steampipe/plugins/turbot/aws:1.0.0
	plugins, err := resolveUpdatePluginsFromArgs(args)
	showProgress := viper.GetBool(pconstants.ArgProgress)

	if err != nil {
		fmt.Println()
		error_helpers.ShowError(ctx, err)
		fmt.Println()
		cmd.Help()
		fmt.Println()
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return
	}

	if len(plugins) > 0 && !(cmdconfig.Viper().GetBool(pconstants.ArgAll)) && plugins[0] == pconstants.ArgAll {
		// improve the response to wrong argument "steampipe plugin update all"
		fmt.Println()
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		error_helpers.ShowError(ctx, fmt.Errorf("Did you mean %s?", pconstants.Bold("--all")))
		fmt.Println()
		return
	}

	state, err := installationstate.Load()
	if err != nil {
		error_helpers.ShowError(ctx, fmt.Errorf("could not load state"))
		exitCode = constants.ExitCodePluginLoadingError
		return
	}

	// retrieve the plugin version data from steampipe config
	pluginVersions := steampipeconfig.GlobalConfig.PluginVersions

	var runUpdatesFor []*versionfile.InstalledVersion
	updateResults := make(pplugin.PluginInstallReports, 0, len(plugins))

	// a leading blank line - since we always output multiple lines
	fmt.Println()

	if cmdconfig.Viper().GetBool(pconstants.ArgAll) {
		for k, v := range pluginVersions {
			ref := putils.NewImageRef(k)
			org, name, constraint := ref.GetOrgNameAndStream()
			key := fmt.Sprintf("%s/%s@%s", org, name, constraint)

			plugins = append(plugins, key)
			runUpdatesFor = append(runUpdatesFor, v)
		}
	} else {
		// get the args and retrieve the installed versions
		for _, p := range plugins {
			ref := putils.NewImageRef(p)
			isExists, _ := pplugin.Exists(ctx, p)
			if isExists {
				if strings.HasPrefix(ref.DisplayImageRef(), constants.SteampipeHubOCIBase) {
					runUpdatesFor = append(runUpdatesFor, pluginVersions[ref.DisplayImageRef()])
				} else {
					error_helpers.ShowError(ctx, fmt.Errorf("cannot check updates for plugins not distributed via hub.steampipe.io, you should uninstall then reinstall the plugin to get the latest version"))
					exitCode = constants.ExitCodePluginLoadingError
					return
				}
			} else {
				exitCode = constants.ExitCodePluginNotFound
				updateResults = append(updateResults, &pplugin.PluginInstallReport{
					Skipped:        true,
					Plugin:         p,
					SkipReason:     pconstants.InstallMessagePluginNotInstalled,
					IsUpdateReport: true,
				})
			}
		}
	}

	if len(plugins) == len(updateResults) {
		// we have report for all
		// this may happen if all given plugins are
		// not installed
		pplugin.PrintInstallReports(updateResults, true)
		fmt.Println()
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	statushooks.SetStatus(ctx, "Checking for available updates")
	reports := pplugin.GetUpdateReport(timeoutCtx, state.InstallationID, runUpdatesFor)
	statushooks.Done(ctx)
	if len(reports) == 0 {
		// this happens if for some reason the update server could not be contacted,
		// in which case we get back an empty map
		error_helpers.ShowError(ctx, fmt.Errorf("there was an issue contacting the update server, please try later"))
		exitCode = constants.ExitCodePluginLoadingError
		return
	}

	updateWaitGroup := &sync.WaitGroup{}
	reportChannel := make(chan *pplugin.PluginInstallReport, len(reports))
	progressBars := uiprogress.New()
	if showProgress {
		progressBars.Start()
	}

	sorted := utils.SortedMapKeys(reports)
	for _, key := range sorted {
		report := reports[key]
		updateWaitGroup.Add(1)
		bar := createProgressBar(report.ShortNameWithConstraint(), progressBars)
		go doPluginUpdate(ctx, bar, report, updateWaitGroup, reportChannel)
	}
	go func() {
		updateWaitGroup.Wait()
		close(reportChannel)
	}()
	installCount := 0

	for updateResult := range reportChannel {
		updateResults = append(updateResults, updateResult)
		if !updateResult.Skipped {
			installCount++
		}
	}
	if showProgress {
		progressBars.Stop()
	}

	pplugin.PrintInstallReports(updateResults, true)

	// a concluding blank line - since we always output multiple lines
	fmt.Println()
}

func doPluginUpdate(ctx context.Context, bar *uiprogress.Bar, pvr pplugin.PluginVersionCheckReport, wg *sync.WaitGroup, returnChannel chan *pplugin.PluginInstallReport) {
	var report *pplugin.PluginInstallReport

	if pplugin.UpdateRequired(pvr) {
		// update required, resolve version and install update
		bar.AppendFunc(func(b *uiprogress.Bar) string {
			// set the progress bar to append itself  with the step underway
			if b.Current() == 0 {
				// no install step to display yet
				return ""
			}
			return helpers.Resize(pluginInstallSteps[b.Current()-1], 20)
		})
		rp := pplugin.NewResolvedPluginVersion(pvr.ShortName(), pvr.CheckResponse.Version, pvr.CheckResponse.Constraint)
		report = installPlugin(ctx, rp, true, bar)
	} else {
		// update NOT required, return already installed report
		bar.AppendFunc(func(b *uiprogress.Bar) string {
			// set the progress bar to append itself with "Already Installed"
			return helpers.Resize(pconstants.InstallMessagePluginLatestAlreadyInstalled, 30)
		})
		// set the progress bar to the maximum
		bar.Set(len(pluginInstallSteps))
		report = &pplugin.PluginInstallReport{
			Plugin:         fmt.Sprintf("%s@%s", pvr.CheckResponse.Name, pvr.CheckResponse.Constraint),
			Skipped:        true,
			SkipReason:     pconstants.InstallMessagePluginLatestAlreadyInstalled,
			IsUpdateReport: true,
		}
	}

	returnChannel <- report
	wg.Done()
}

func createProgressBar(plugin string, parentProgressBars *uiprogress.Progress) *uiprogress.Bar {
	bar := parentProgressBars.AddBar(len(pluginInstallSteps))
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return helpers.Resize(plugin, 30)
	})
	return bar
}

func installPlugin(ctx context.Context, resolvedPlugin pplugin.ResolvedPluginVersion, isUpdate bool, bar *uiprogress.Bar) *pplugin.PluginInstallReport {
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
	
	skipConfig := viper.GetBool(pconstants.ArgSkipConfig)
	// we should never install the config file for plugin updates; config files should only be installed during plugin install
	if isUpdate {
		skipConfig = true
	}

	image, err := plugin.Install(ctx, resolvedPlugin, progress, constants.BaseImageRef, ociinstaller.SteampipeMediaTypeProvider{}, putils.WithSkipConfig(skipConfig))
	if err != nil {
		msg := ""
		// used to build data for the plugin install report to be used for display purposes
		_, name, constraint := putils.NewImageRef(resolvedPlugin.GetVersionTag()).GetOrgNameAndStream()
		if isPluginNotFoundErr(err) {
			exitCode = constants.ExitCodePluginNotFound
			msg = pconstants.InstallMessagePluginNotFound
		} else {
			msg = err.Error()
		}
		return &pplugin.PluginInstallReport{
			Plugin:         fmt.Sprintf("%s@%s", name, constraint),
			Skipped:        true,
			SkipReason:     msg,
			IsUpdateReport: isUpdate,
		}
	}

	// used to build data for the plugin install report to be used for display purposes
	org, name, _ := image.ImageRef.GetOrgNameAndStream()
	versionString := ""
	if image.Config.Plugin.Version != "" {
		versionString = " v" + image.Config.Plugin.Version
	}
	docURL := fmt.Sprintf("https://hub.steampipe.io/plugins/%s/%s", org, name)
	if !image.ImageRef.IsFromTurbotHub() {
		docURL = fmt.Sprintf("https://%s/%s", org, name)
	}
	return &pplugin.PluginInstallReport{
		Plugin:         fmt.Sprintf("%s@%s", name, resolvedPlugin.Constraint),
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
		return nil, fmt.Errorf("you need to provide at least one plugin to update or use the %s flag", pconstants.Bold("--all"))
	}

	if len(plugins) > 0 && cmdconfig.Viper().GetBool(pconstants.ArgAll) {
		// we can't allow update and install at the same time
		return nil, fmt.Errorf("%s cannot be used when updating specific plugins", pconstants.Bold("`--all`"))
	}

	return plugins, nil
}

func runPluginListCmd(cmd *cobra.Command, _ []string) {
	// setup a cancel context and start cancel handler
	ctx, cancel := context.WithCancel(cmd.Context())
	contexthelpers.StartCancelHandler(cancel)
	outputFormat := viper.GetString(pconstants.ArgOutput)

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

	err := showPluginListOutput(pluginList, failedPluginMap, missingPluginMap, res, outputFormat)
	if err != nil {
		error_helpers.ShowError(ctx, err)
	}

}

func showPluginListOutput(pluginList []plugin.PluginListItem, failedPluginMap, missingPluginMap map[string][]plugin.PluginConnection, res perror_helpers.ErrorAndWarnings, outputFormat string) error {
	switch outputFormat {
	case "table":
		return showPluginListAsTable(pluginList, failedPluginMap, missingPluginMap, res)
	case "json":
		return showPluginListAsJSON(pluginList, failedPluginMap, missingPluginMap, res)
	default:
		return errors.New("invalid output format")
	}
}

func showPluginListAsTable(pluginList []plugin.PluginListItem, failedPluginMap, missingPluginMap map[string][]plugin.PluginConnection, res perror_helpers.ErrorAndWarnings) error {
	headers := []string{"Installed", "Version", "Connections"}
	var rows [][]string
	// List installed plugins in a table
	if len(pluginList) != 0 {
		for _, item := range pluginList {
			rows = append(rows, []string{item.Name, item.Version.String(), strings.Join(item.Connections, ",")})
		}
	} else {
		rows = append(rows, []string{"", "", ""})
	}
	querydisplay.ShowWrappedTable(headers, rows, &querydisplay.ShowWrappedTableOptions{AutoMerge: false})
	fmt.Printf("\n")

	// List failed/missing plugins in a separate table
	if len(failedPluginMap)+len(missingPluginMap) != 0 {
		headers := []string{"Failed", "Connections", "Reason"}
		var conns []string
		var missingRows [][]string

		// failed plugins
		for p, item := range failedPluginMap {
			for _, conn := range item {
				conns = append(conns, conn.GetName())
			}
			missingRows = append(missingRows, []string{p, strings.Join(conns, ","), pconstants.ConnectionErrorPluginFailedToStart})
			conns = []string{}
		}

		// missing plugins
		for p, item := range missingPluginMap {
			for _, conn := range item {
				conns = append(conns, conn.GetName())
			}
			missingRows = append(missingRows, []string{p, strings.Join(conns, ","), pconstants.InstallMessagePluginNotInstalled})
			conns = []string{}
		}

		querydisplay.ShowWrappedTable(headers, missingRows, &querydisplay.ShowWrappedTableOptions{AutoMerge: false})
		fmt.Println()
	}

	if len(res.Warnings) > 0 {
		fmt.Println()
		res.ShowWarnings()
		fmt.Printf("\n")
	}
	return nil
}

func showPluginListAsJSON(pluginList []plugin.PluginListItem, failedPluginMap, missingPluginMap map[string][]plugin.PluginConnection, res perror_helpers.ErrorAndWarnings) error {
	output := pluginJsonOutput{}

	for _, item := range pluginList {
		installed := installedPlugin{
			Name:        item.Name,
			Version:     item.Version.String(),
			Connections: item.Connections,
		}
		output.Installed = append(output.Installed, installed)
	}

	for p, item := range failedPluginMap {
		connections := make([]string, len(item))
		for i, conn := range item {
			connections[i] = conn.GetName()
		}
		failed := failedPlugin{
			Name:        p,
			Connections: connections,
			Reason:      pconstants.ConnectionErrorPluginFailedToStart,
		}
		output.Failed = append(output.Failed, failed)
	}

	for p, item := range missingPluginMap {
		connections := make([]string, len(item))
		for i, conn := range item {
			connections[i] = conn.GetName()
		}
		missing := failedPlugin{
			Name:        p,
			Connections: connections,
			Reason:      pconstants.InstallMessagePluginNotInstalled,
		}
		output.Failed = append(output.Failed, missing)
	}

	if len(res.Warnings) > 0 {
		output.Warnings = res.Warnings
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonOutput))
	fmt.Println()
	return nil
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

	reports := plugin.PluginRemoveReports{}
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

func getPluginList(ctx context.Context) (pluginList []plugin.PluginListItem, failedPluginMap, missingPluginMap map[string][]plugin.PluginConnection, res perror_helpers.ErrorAndWarnings) {
	statushooks.Show(ctx)
	defer statushooks.Done(ctx)

	// get the maps of available and failed/missing plugins
	pluginConnectionMap, failedPluginMap, missingPluginMap, res := getPluginConnectionMap(ctx)
	if res.Error != nil {
		return nil, nil, nil, res
	}

	// retrieve the plugin version data from steampipe config
	pluginVersions := steampipeconfig.GlobalConfig.PluginVersions

	// TODO do we really need to look at installed plugins - can't we just use the plugin connection map
	// get a list of the installed plugins by inspecting the install location
	// pass pluginConnectionMap so we can populate the connections for each plugin
	pluginList, err := plugin.List(ctx, pluginConnectionMap, pluginVersions)
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

func getPluginConnectionMap(ctx context.Context) (pluginConnectionMap, failedPluginMap, missingPluginMap map[string][]plugin.PluginConnection, res perror_helpers.ErrorAndWarnings) {
	utils.LogTime("cmd.getPluginConnectionMap start")
	defer utils.LogTime("cmd.getPluginConnectionMap end")

	statushooks.SetStatus(ctx, "Fetching connection map")

	res = perror_helpers.ErrorAndWarnings{}

	connectionStateMap, stateRes := getConnectionState(ctx)
	res.Merge(stateRes)
	if res.Error != nil {
		return nil, nil, nil, res
	}

	// create the map of failed/missing plugins and available/loaded plugins
	failedPluginMap = map[string][]plugin.PluginConnection{}
	missingPluginMap = map[string][]plugin.PluginConnection{}
	pluginConnectionMap = make(map[string][]plugin.PluginConnection)

	for _, state := range connectionStateMap {
		connection, ok := steampipeconfig.GlobalConfig.Connections[state.ConnectionName]
		if !ok {
			continue
		}

		if state.State == constants.ConnectionStateError && state.Error() == pconstants.ConnectionErrorPluginFailedToStart {
			failedPluginMap[state.Plugin] = append(failedPluginMap[state.Plugin], connection)
		} else if state.State == constants.ConnectionStateError && state.Error() == pconstants.ConnectionErrorPluginNotInstalled {
			missingPluginMap[state.Plugin] = append(missingPluginMap[state.Plugin], connection)
		}

		pluginConnectionMap[state.Plugin] = append(pluginConnectionMap[state.Plugin], connection)
	}

	return pluginConnectionMap, failedPluginMap, missingPluginMap, res
}

// load the connection state, waiting until all connections are loaded
func getConnectionState(ctx context.Context) (steampipeconfig.ConnectionStateMap, perror_helpers.ErrorAndWarnings) {
	utils.LogTime("cmd.getConnectionState start")
	defer utils.LogTime("cmd.getConnectionState end")

	// start service
	client, res := db_local.GetLocalClient(ctx, constants.InvokerPlugin)
	if res.Error != nil {
		return nil, res
	}
	defer client.Close(ctx)

	conn, err := client.AcquireManagementConnection(ctx)
	if err != nil {
		res.Error = err
		return nil, res
	}
	defer conn.Release()

	// load connection state
	statushooks.SetStatus(ctx, "Loading connection state")
	connectionStateMap, err := steampipeconfig.LoadConnectionState(ctx, conn.Conn(), steampipeconfig.WithWaitUntilReady())
	if err != nil {
		res.Error = err
		return nil, res
	}

	return connectionStateMap, res
}
