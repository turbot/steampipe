package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/workspace"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v3/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/contexthelpers"
	"github.com/turbot/steampipe/dashboard"
	"github.com/turbot/steampipe/dashboard/dashboardassets"
	"github.com/turbot/steampipe/dashboard/dashboardserver"
	"github.com/turbot/steampipe/utils"
)

func dashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "dashboard",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runDashboardCmd,
		Short:            "Start the local dashboard UI",
		Long: `Starts a local web server that enables real-time development of dashboards within the current mod.

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
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, "", nil, "Specify the value of a variable").

		// hidden flags that are used internally
		AddBoolFlag(constants.ArgServiceMode, "", false, "Hidden flag to specify whether this is starting as a service", cmdconfig.FlagOptions.Hidden())

	return cmd
}

func runDashboardCmd(cmd *cobra.Command, args []string) {
	dashboardCtx := cmd.Context()

	logging.LogTime("runDashboardCmd start")
	defer func() {
		logging.LogTime("runDashboardCmd end")
		if r := recover(); r != nil {
			utils.ShowError(dashboardCtx, helpers.ToError(r))
			if isRunningAsService() {
				saveErrorToDashboardState(helpers.ToError(r))
			}
		}
	}()

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
	err := dashboardassets.Ensure(dashboardCtx)
	utils.FailOnError(err)

	// disable all status messages
	dashboardCtx = statushooks.DisableStatusHooks(dashboardCtx)

	// load the workspace
	dashboardserver.OutputWait(dashboardCtx, "Loading Workspace")
	w, err := loadWorkspacePromptingForVariables(dashboardCtx)
	utils.FailOnErrorWithMessage(err, "failed to load workspace")

	initData := dashboard.NewInitData(dashboardCtx, w)
	// shutdown the service on exit
	defer initData.Cleanup(dashboardCtx)

	if shouldExit, err := handleDashboardInitResult(dashboardCtx, initData); shouldExit {
		// if there was an error, display it
		utils.FailOnError(err)
		return
	}

	server, err := dashboardserver.NewServer(dashboardCtx, initData.Client, initData.Workspace)
	if err != nil {
		utils.FailOnError(err)
	}
	// start the server asynchronously - this returns a chan which is signalled when the internal API server terminates
	doneChan := server.Start()

	// cleanup
	defer server.Shutdown()

	// server has started - update state file/start browser, as required
	onServerStarted(serverPort, serverListen, err)

	// wait for API server to terminate
	<-doneChan

	log.Println("[TRACE] runDashboardCmd exiting")
}

// inspect the init result ands
func handleDashboardInitResult(ctx context.Context, initData *dashboard.InitData) (bool, error) {
	// if there is an error or cancellation we bomb out
	if err := initData.Result.Error; err != nil {
		setExitCodeForDashboardError(err)
		return true, initData.Result.Error
	}
	// cancelled?
	if ctx != nil && ctx.Err() != nil {
		return true, ctx.Err()
	}
	// if there is a usage warning we display it
	initData.Result.DisplayMessages()

	// if there is are any warnings, exit politely
	shouldExit := len(initData.Result.Warnings) > 0

	return shouldExit, nil
}

func setExitCodeForDashboardError(err error) {
	if err == workspace.ErrorNoModDefinition {
		exitCode = constants.ExitCodeNoModFile
	} else {
		exitCode = constants.ExitCodeUnknownErrorPanic
	}
}

// execute any required actions after successfult server startup
func onServerStarted(serverPort dashboardserver.ListenPort, serverListen dashboardserver.ListenType, err error) {
	if isRunningAsService() {
		// for service mode only, save the state
		saveDashboardState(serverPort, serverListen)
	} else {
		// start browser if required
		if viper.GetBool(constants.ArgBrowser) {
			if err = dashboardserver.OpenBrowser(fmt.Sprintf("http://localhost:%d", serverPort)); err != nil {
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
