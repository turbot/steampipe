package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/turbot/steampipe/statushooks"

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
		AddBoolFlag(constants.ArgDashboardClient, "", true, "Start a browser based dashboard client automatically.", cmdconfig.FlagOptions.Hidden())
	return cmd
}

func runDashboardCmd(cmd *cobra.Command, args []string) {
	// create context for the dashboard execution
	ctx, cancel := context.WithCancel(cmd.Context())
	// disable all status messages
	dashboardCtx := statushooks.DisableStatusHooks(ctx)

	contexthelpers.StartCancelHandler(cancel)

	logging.LogTime("runDashboardCmd start")
	defer func() {
		logging.LogTime("runDashboardCmd end")
		if r := recover(); r != nil {
			utils.ShowError(dashboardCtx, helpers.ToError(r))
		}
	}()

	serverPort := dashboardserver.ListenPort(viper.GetInt(constants.ArgDashboardPort))
	utils.FailOnError(serverPort.IsValid())

	serverListen := dashboardserver.ListenType(viper.GetString(constants.ArgDashboardListen))
	utils.FailOnError(serverListen.IsValid())

	// ensure dashboard assets are present and extract if not
	err := dashboardassets.Ensure(dashboardCtx)
	utils.FailOnError(err)

	// load the workspace
	w, err := loadWorkspacePromptingForVariables(dashboardCtx)
	utils.FailOnErrorWithMessage(err, "failed to load workspace")

	initData := dashboard.NewInitData(dashboardCtx, w)
	if shouldExit := handleDashboardInitResult(dashboardCtx, initData); shouldExit {
		return
	}
	server, err := dashboardserver.NewServer(dashboardCtx, initData.Client, initData.Workspace)
	if err != nil {
		utils.FailOnError(err)
	}

	server.Start()

	if viper.GetBool(constants.ArgDashboardClient) {
		err = dashboardserver.OpenBrowser(fmt.Sprintf("http://localhost:%d", serverPort))
		if err != nil {
			log.Println("[TRACE] dashboard server started but failed to start client", err)
		}
	}

	// wait for the given context to cancel
	<-dashboardCtx.Done()

	server.Shutdown(dashboardCtx)
}

func handleDashboardInitResult(ctx context.Context, initData *dashboard.InitData) bool {
	// if there is an error or cancellation we bomb out
	// check for the various kinds of failures
	utils.FailOnError(initData.Result.Error)
	// cancelled?
	if ctx != nil {
		utils.FailOnError(ctx.Err())
	}

	// if there is a usage warning we display it
	initData.Result.DisplayMessages()

	// if there is are any warnings, exit politely
	shouldExit := len(initData.Result.Warnings) > 0

	return shouldExit
}
