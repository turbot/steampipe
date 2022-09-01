package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/steampipe/pkg/cloud"
	"github.com/turbot/steampipe/pkg/initialisation"
	"log"
	"os"

	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/workspace"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v4/logging"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardassets"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardserver"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/snapshot"
	"github.com/turbot/steampipe/pkg/utils"
)

func dashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "dashboard [flags] [benchmark/dashboard]",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runDashboardCmd,
		Short:            "Start the local dashboard UI or run a named dashboard",
		Long: `Either runs the a named dashboard or benchmark, or starts a local web server that enables real-time development of dashboards within the current mod.

The current mod is the working directory, or the directory specified by the --workspace-chdir flag.`,
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
		AddStringFlag(constants.ArgSnapshot, "", "", "Create snapshot in Steampipe Cloud with the default (workspace) visibility.", cmdconfig.FlagOptions.NoOptDefVal(constants.ArgShareNoOptDefault)).
		AddStringFlag(constants.ArgShare, "", "", "Create snapshot in Steampipe Cloud with 'anyone_with_link' visibility.", cmdconfig.FlagOptions.NoOptDefVal(constants.ArgShareNoOptDefault)).
		// NOTE: use StringArrayFlag for ArgDashboardInput, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV, where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgDashboardInput, "", nil, "Specify the value of a dashboard input").
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
			utils.ShowError(dashboardCtx, err)
			if isRunningAsService() {
				saveErrorToDashboardState(err)
			}
		}
		setExitCodeForDashboardError(err)

	}()

	// first check whether a dashboard name has been passed as an arg
	dashboardName, err := validateDashboardArgs(args)
	utils.FailOnError(err)
	if dashboardName != "" {
		// run just this dashboard
		err = runSingleDashboard(dashboardCtx, dashboardName)
		utils.FailOnError(err)
		// and we are done
		return
	}

	// retrieve server params
	serverPort := dashboardserver.ListenPort(viper.GetInt(constants.ArgDashboardPort))
	utils.FailOnError(serverPort.IsValid())

	serverListen := dashboardserver.ListenType(viper.GetString(constants.ArgDashboardListen))
	utils.FailOnError(serverListen.IsValid())

	if err := utils.IsPortBindable(int(serverPort)); err != nil {
		exitCode = constants.ExitCodeBindPortUnavailable
		utils.FailOnError(err)
	}

	// create context for the dashboard execution
	dashboardCtx, cancel := context.WithCancel(dashboardCtx)
	contexthelpers.StartCancelHandler(cancel)

	// ensure dashboard assets are present and extract if not
	err = dashboardassets.Ensure(dashboardCtx)
	utils.FailOnError(err)

	// disable all status messages
	dashboardCtx = statushooks.DisableStatusHooks(dashboardCtx)

	// load the workspace
	initData := initDashboard(dashboardCtx, err)
	defer initData.Cleanup(dashboardCtx)
	utils.FailOnError(initData.Result.Error)

	// if there is a usage warning we display it
	initData.Result.DisplayMessages()

	// create the server
	server, err := dashboardserver.NewServer(dashboardCtx, initData.Client, initData.Workspace)
	utils.FailOnError(err)

	// start the server asynchronously - this returns a chan which is signalled when the internal API server terminates
	doneChan := server.Start()

	// cleanup
	defer server.Shutdown()

	// server has started - update state file/start browser, as required
	onServerStarted(serverPort, serverListen)

	// wait for API server to terminate
	<-doneChan

	log.Println("[TRACE] runDashboardCmd exiting")
}

func initDashboard(dashboardCtx context.Context, err error) *initialisation.InitData {
	dashboardserver.OutputWait(dashboardCtx, "Loading Workspace")
	w, err := interactive.LoadWorkspacePromptingForVariables(dashboardCtx)
	utils.FailOnErrorWithMessage(err, "failed to load workspace")

	// initialise
	initData := initialisation.NewInitData(dashboardCtx, w)
	// there must be a modfile
	if !w.ModfileExists() {
		initData.Result.Error = workspace.ErrorNoModDefinition
	}

	return initData
}

func runSingleDashboard(ctx context.Context, dashboardName string) error {
	// so a dashboard name was specified - just call GenerateSnapshot
	snapshot, err := snapshot.GenerateSnapshot(ctx, dashboardName)
	if err != nil {
		return err
	}

	shouldShare := viper.IsSet(constants.ArgShare)
	shouldUpload := viper.IsSet(constants.ArgSnapshot)
	if shouldShare || shouldUpload {
		snapshotUrl, err := cloud.UploadSnapshot(snapshot, shouldShare)
		statushooks.Done(ctx)
		if err != nil {
			return err
		} else {
			fmt.Printf("Snapshot uploaded to %s\n", snapshotUrl)
		}
		return err
	}

	// just display result
	snapshotText, err := json.MarshalIndent(snapshot, "", "  ")
	utils.FailOnError(err)
	fmt.Println(string(snapshotText))
	fmt.Println("")
	return nil
}

func validateDashboardArgs(args []string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("dashboard command accepts 0 or 1 argument")
	}
	dashboardName := ""
	if len(args) == 1 {
		dashboardName = args[0]
	}

	// only 1 of 'share' and 'snapshot' may be set
	shareArg := viper.GetString(constants.ArgShare)
	snapshotArg := viper.GetString(constants.ArgSnapshot)
	if shareArg != "" && snapshotArg != "" {
		return "", fmt.Errorf("only 1 of --share and --dashboard may be set")
	}

	// if either share' or 'snapshot' are set, a dashboard name an dcloud token must be provided
	if shareArg != "" || snapshotArg != "" {
		if dashboardName == "" {
			return "", fmt.Errorf("dashboard name must be provided if --share or --snapshot arg is used")
		}
		snapshotWorkspace := shareArg
		argName := "share"
		if snapshotWorkspace == "" {
			snapshotWorkspace = snapshotArg
			argName = "snapshot"
		}

		// is this is the no-option default, use the workspace arg
		if snapshotWorkspace == constants.ArgShareNoOptDefault {
			snapshotWorkspace = viper.GetString(constants.ArgWorkspace)
		}
		if snapshotWorkspace == "" {
			return "", fmt.Errorf("a Steampipe Cloud workspace name must be provided, either by setting %s=<workspace> or --workspace=<workspace>", argName)
		}

		// now write back the workspace to viper
		viper.Set(constants.ArgWorkspace, snapshotWorkspace)

		// verify cloud token
		if !viper.IsSet(constants.ArgCloudToken) {
			return "", fmt.Errorf("a Steampipe Cloud token must be provided")
		}
	}

	return dashboardName, nil
}

func setExitCodeForDashboardError(err error) {
	// if exit code already set, leave as is
	if exitCode != 0 {
		return
	}

	if err == workspace.ErrorNoModDefinition {
		exitCode = constants.ExitCodeNoModFile
	} else {
		exitCode = constants.ExitCodeUnknownErrorPanic
	}
}

// execute any required actions after successful server startup
func onServerStarted(serverPort dashboardserver.ListenPort, serverListen dashboardserver.ListenType) {
	if isRunningAsService() {
		// for service mode only, save the state
		saveDashboardState(serverPort, serverListen)
	} else {
		// start browser if required
		if viper.GetBool(constants.ArgBrowser) {
			if err := dashboardserver.OpenBrowser(fmt.Sprintf("http://localhost:%d", serverPort)); err != nil {
				log.Println("[TRACE] dashboard server started but failed to start client", err)
			}
		}
	}
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
	utils.FailOnError(dashboardserver.WriteServiceStateFile(state))
}
