package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v4/logging"
	"github.com/turbot/steampipe/pkg/cloud"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardassets"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardexecute"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardserver"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/workspace"
)

func dashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "dashboard [flags] [benchmark/dashboard]",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runDashboardCmd,
		Short:            "Start the local dashboard UI or run a named dashboard",
		Long: `Either runs the a named dashboard or benchmark, or starts a local web server that enables real-time development of dashboards within the current mod.

The current mod is the working directory, or the directory specified by the --mod-location flag.`,
	}

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for dashboard").
		AddBoolFlag(constants.ArgModInstall, "", true, "Specify whether to install mod dependencies before running the dashboard").
		AddStringFlag(constants.ArgDashboardListen, "", string(dashboardserver.ListenTypeLocal), "Accept connections from: local (localhost only) or network (open)").
		AddIntFlag(constants.ArgDashboardPort, "", constants.DashboardServerDefaultPort, "Dashboard server port.").
		AddBoolFlag(constants.ArgBrowser, "", true, "Specify whether to launch the browser after starting the dashboard server").
		AddStringSliceFlag(constants.ArgSearchPath, "", nil, "Set a custom search_path for the steampipe user for a check session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", nil, "Set a prefix to the current search path for a check session (comma-separated)").
		AddStringSliceFlag(constants.ArgVarFile, "", nil, "Specify an .spvar file containing variable values").
		AddBoolFlag(constants.ArgProgress, "", true, "Display dashboard execution progress respected when a dashboard name argument is passed").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV, where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, "", nil, "Specify the value of a variable").
		AddBoolFlag(constants.ArgInput, "", true, "Enable interactive prompts").
		AddStringFlag(constants.ArgOutput, "", constants.OutputFormatSnapshot, "Select a console output format: snapshot").
		AddBoolFlag(constants.ArgSnapshot, "", false, "Create snapshot in Steampipe Cloud with the default (workspace) visibility.").
		AddBoolFlag(constants.ArgShare, "", false, "Create snapshot in Steampipe Cloud with 'anyone_with_link' visibility.").
		AddStringFlag(constants.ArgSnapshotLocation, "", "", "The cloud workspace... ").
		// NOTE: use StringArrayFlag for ArgDashboardInput, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV, where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgDashboardInput, "", nil, "Specify the value of a dashboard input").
		AddStringSliceFlag(constants.ArgSnapshotTag, "", nil, "Specify tags to set on the snapshot").
		AddStringArrayFlag(constants.ArgSourceSnapshot, "", nil, "Specify one or more snapshots to display").
		AddStringSliceFlag(constants.ArgExport, "", nil, "Export output to a snapshot file").
		// hidden flags that are used internally
		AddBoolFlag(constants.ArgServiceMode, "", false, "Hidden flag to specify whether this is starting as a service", cmdconfig.FlagOptions.Hidden())

	return cmd
}

func runDashboardCmd(cmd *cobra.Command, args []string) {
	dashboardCtx := cmd.Context()

	var err error
	logging.LogTime("runDashboardCmd start")
	defer func() {
		logging.LogTime("runDashboardCmd end")
		if r := recover(); r != nil {
			err = helpers.ToError(r)
			error_helpers.ShowError(dashboardCtx, err)
			if isRunningAsService() {
				saveErrorToDashboardState(err)
			}
		}
		setExitCodeForDashboardError(err)

	}()

	// first check whether a dashboard name has been passed as an arg
	dashboardName, err := validateDashboardArgs(cmd, args)
	error_helpers.FailOnError(err)
	if dashboardName != "" {
		inputs, err := collectInputs()
		error_helpers.FailOnError(err)

		// run just this dashboard
		err = runSingleDashboard(dashboardCtx, dashboardName, inputs)
		error_helpers.FailOnError(err)

		// and we are done
		return
	}

	// retrieve server params
	serverPort := dashboardserver.ListenPort(viper.GetInt(constants.ArgDashboardPort))
	error_helpers.FailOnError(serverPort.IsValid())

	serverListen := dashboardserver.ListenType(viper.GetString(constants.ArgDashboardListen))
	error_helpers.FailOnError(serverListen.IsValid())

	if err := utils.IsPortBindable(int(serverPort)); err != nil {
		exitCode = constants.ExitCodeBindPortUnavailable
		error_helpers.FailOnError(err)
	}

	// create context for the dashboard execution
	dashboardCtx, cancel := context.WithCancel(dashboardCtx)
	contexthelpers.StartCancelHandler(cancel)

	// ensure dashboard assets are present and extract if not
	err = dashboardassets.Ensure(dashboardCtx)
	error_helpers.FailOnError(err)

	// disable all status messages
	dashboardCtx = statushooks.DisableStatusHooks(dashboardCtx)

	// load the workspace
	initData := initDashboard(dashboardCtx)
	defer initData.Cleanup(dashboardCtx)
	error_helpers.FailOnError(initData.Result.Error)

	// if there is a usage warning we display it
	initData.Result.DisplayMessage = dashboardserver.OutputMessage
	initData.Result.DisplayWarning = dashboardserver.OutputWarning
	initData.Result.DisplayMessages()

	// create the server
	server, err := dashboardserver.NewServer(dashboardCtx, initData.Client, initData.Workspace)
	error_helpers.FailOnError(err)

	// start the server asynchronously - this returns a chan which is signalled when the internal API server terminates
	doneChan := server.Start()

	// cleanup
	defer server.Shutdown()

	// server has started - update state file/start browser, as required
	onServerStarted(serverPort, serverListen, initData.Workspace)

	// wait for API server to terminate
	<-doneChan

	log.Println("[TRACE] runDashboardCmd exiting")
}

// validate the args and extract a dashboard name, if provided
func validateDashboardArgs(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("dashboard command accepts 0 or 1 argument")
	}
	dashboardName := ""
	if len(args) == 1 {
		dashboardName = args[0]
	}

	err := cmdconfig.ValidateCloudArgs()
	if err != nil {
		return "", err
	}

	// only 1 of 'share' and 'snapshot' may be set
	share := viper.GetBool(constants.ArgShare)
	snapshot := viper.GetBool(constants.ArgSnapshot)

	// if either share' or 'snapshot' are set, a dashboard name and cloud token must be provided
	if share || snapshot {
		if dashboardName == "" {
			return "", fmt.Errorf("dashboard name must be provided if --share or --snapshot arg is used")
		}

		if viper.GetString(constants.ArgSnapshotLocation) == "" {
			return "", fmt.Errorf("a snapshot location must be set")
		}

		// verify cloud token
		if !viper.IsSet(constants.ArgCloudToken) {
			return "", fmt.Errorf("a Steampipe Cloud token must be provided")
		}
	}

	validOutputFormats := []string{constants.OutputFormatSnapshot, constants.OutputFormatNone}
	output := viper.GetString(constants.ArgOutput)
	if !helpers.StringSliceContains(validOutputFormats, output) {
		return "", fmt.Errorf("invalid output format '%s', must be one of %s", output, strings.Join(validOutputFormats, ","))
	}

	return dashboardName, nil
}

func displaySnapshot(snapshot *dashboardtypes.SteampipeSnapshot) {
	switch viper.GetString(constants.ArgOutput) {
	case constants.OutputFormatSnapshot:
		// just display result
		snapshotText, err := json.MarshalIndent(snapshot, "", "  ")
		error_helpers.FailOnError(err)
		fmt.Println(string(snapshotText))
	}
}

func initDashboard(ctx context.Context) *initialisation.InitData {
	sourceSnapshots := viper.GetStringSlice(constants.ArgSourceSnapshot)
	if len(sourceSnapshots) > 0 {
		for _, s := range sourceSnapshots {
			if !filehelpers.FileExists(s) {
				return initialisation.NewErrorInitData(fmt.Errorf("source snapshot' %s' does not exist", s))
			}
		}
		dashboardserver.OutputWait(ctx, "Loading Source Snapshots")
		w := workspace.NewSourceSnapshotWorkspace(sourceSnapshots)
		// return init data containing only this workspace - do not initialise it
		return initialisation.NewInitData(w)
	}

	dashboardserver.OutputWait(ctx, "Loading Workspace")
	w, err := interactive.LoadWorkspacePromptingForVariables(ctx)
	if err != nil {
		return initialisation.NewErrorInitData(fmt.Errorf("failed to load workspace: %s", err.Error()))
	}

	// initialise
	initData := getInitData(ctx, w)

	// there must be a mod-file
	if !w.ModfileExists() {
		initData.Result.Error = workspace.ErrorNoModDefinition
	}

	return initData
}

func getInitData(ctx context.Context, w *workspace.Workspace) *initialisation.InitData {
	initData := initialisation.NewInitData(w).
		RegisterExporters(dashboardExporters()...).
		Init(ctx, constants.InvokerDashboard)
	return initData
}

func dashboardExporters() []export.Exporter {
	return []export.Exporter{&export.SnapshotExporter{}}
}

func runSingleDashboard(ctx context.Context, targetName string, inputs map[string]interface{}) error {
	w, err := interactive.LoadWorkspacePromptingForVariables(ctx)
	error_helpers.FailOnErrorWithMessage(err, "failed to load workspace")

	// targetName must be a named resource
	// parse the name to verify
	if err := verifyNamedResource(targetName, w); err != nil {
		return err
	}

	initData := getInitData(ctx, w)

	// shutdown the service on exit
	defer initData.Cleanup(ctx)
	if err := initData.Result.Error; err != nil {
		return initData.Result.Error
	}

	// if there is a usage warning we display it
	initData.Result.DisplayMessages()

	// so a dashboard name was specified - just call GenerateSnapshot
	snap, err := dashboardexecute.GenerateSnapshot(ctx, targetName, initData, inputs)
	if err != nil {
		return err
	}
	// display the snapshot result (if needed)
	displaySnapshot(snap)

	// upload the snapshot (if needed)
	err = uploadSnapshot(snap)
	error_helpers.FailOnErrorWithMessage(err, "failed to upload snapshot")

	// export the result (if needed)
	exportArgs := viper.GetStringSlice(constants.ArgExport)
	err = initData.ExportManager.DoExport(ctx, targetName, snap, exportArgs)
	error_helpers.FailOnErrorWithMessage(err, "failed to export snapshot")

	return nil
}

func verifyNamedResource(targetName string, w *workspace.Workspace) error {
	parsedName, err := modconfig.ParseResourceName(targetName)
	if err != nil {
		return fmt.Errorf("dashboard command cannot run arbitrary SQL")
	}
	if parsedName.ItemType == "" {
		return fmt.Errorf("dashboard command cannot run arbitrary SQL")
	}
	if _, found := modconfig.GetResource(w, parsedName); !found {
		return fmt.Errorf("'%s' not found in workspace", targetName)
	}
	return nil
}

func uploadSnapshot(snapshot *dashboardtypes.SteampipeSnapshot) error {
	shouldShare := viper.IsSet(constants.ArgShare)
	shouldUpload := viper.IsSet(constants.ArgSnapshot)
	if shouldShare || shouldUpload {
		snapshotUrl, err := cloud.UploadSnapshot(snapshot, shouldShare)
		if err != nil {
			return err
		} else {
			if viper.GetBool(constants.ArgProgress) {
				fmt.Printf("Snapshot uploaded to %s\n", snapshotUrl)
			}
		}
	}
	return nil
}

func setExitCodeForDashboardError(err error) {
	// if exit code already set, leave as is
	if exitCode != 0 || err == nil {
		return
	}

	if err == workspace.ErrorNoModDefinition {
		exitCode = constants.ExitCodeNoModFile
	} else {
		exitCode = constants.ExitCodeUnknownErrorPanic
	}
}

// execute any required actions after successful server startup
func onServerStarted(serverPort dashboardserver.ListenPort, serverListen dashboardserver.ListenType, w *workspace.Workspace) {
	if isRunningAsService() {
		// for service mode only, save the state
		saveDashboardState(serverPort, serverListen)
	} else {
		// start browser if required
		if viper.GetBool(constants.ArgBrowser) {
			url := buildDashboardURL(serverPort, w)

			if err := dashboardserver.OpenBrowser(url); err != nil {
				log.Println("[TRACE] dashboard server started but failed to start client", err)
			}
		}
	}
}

func buildDashboardURL(serverPort dashboardserver.ListenPort, w *workspace.Workspace) string {
	url := fmt.Sprintf("http://localhost:%d", serverPort)
	if len(w.SourceSnapshots) == 1 {
		for snapshotName := range w.GetResourceMaps().Snapshots {
			url += fmt.Sprintf("/%s", snapshotName)
			break
		}
	}
	return url
}

// is this dashboard server running as a service?
func isRunningAsService() bool {
	return viper.GetBool(constants.ArgServiceMode)
}

// persist the error to the dashboard state file
func saveErrorToDashboardState(err error) {
	state, _ := dashboardserver.GetDashboardServiceState()
	if state == nil {
		// write the state file with an error, only if it doesn't exist already
		// if it exists, that means dashboard stated properly and 'service start' already known about it
		state = &dashboardserver.DashboardServiceState{
			State: dashboardserver.ServiceStateError,
			Error: err.Error(),
		}
		dashboardserver.WriteServiceStateFile(state)
	}
}

// save the dashboard state file
func saveDashboardState(serverPort dashboardserver.ListenPort, serverListen dashboardserver.ListenType) {
	state := &dashboardserver.DashboardServiceState{
		State:      dashboardserver.ServiceStateRunning,
		Error:      "",
		Pid:        os.Getpid(),
		Port:       int(serverPort),
		ListenType: string(serverListen),
		Listen:     constants.DatabaseListenAddresses,
	}

	if serverListen == dashboardserver.ListenTypeNetwork {
		addrs, _ := utils.LocalAddresses()
		state.Listen = append(state.Listen, addrs...)
	}
	error_helpers.FailOnError(dashboardserver.WriteServiceStateFile(state))
}

func collectInputs() (map[string]interface{}, error) {
	res := make(map[string]interface{})
	inputArgs := viper.GetStringSlice(constants.ArgDashboardInput)
	for _, variableArg := range inputArgs {
		// Value should be in the form "name=value", where value is a string
		raw := variableArg
		eq := strings.Index(raw, "=")
		if eq == -1 {
			return nil, fmt.Errorf("the --dashboard-input argument '%s' is not correctly specified. It must be an input name and value separated an equals sign: --dashboard-input key=value", raw)
		}
		name := raw[:eq]
		rawVal := raw[eq+1:]
		if _, ok := res[name]; ok {
			return nil, fmt.Errorf("the dashboard-input option '%s' is provided more than once", name)
		}
		// TACTICAL: add `input. to start of name
		key := fmt.Sprintf("input.%s", name)
		res[key] = rawVal
	}

	return res, nil

}
