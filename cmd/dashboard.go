package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v3/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/contexthelpers"
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
		AddStringFlag(constants.ArgDashboardServerListen, "", string(dashboardserver.ListenTypeLocal), "Accept connections from: local (localhost only) or network (open)").
		AddIntFlag(constants.ArgDashboardServerPort, "", constants.DashboardServerDefaultPort, "Dashboard server port.").
		AddBoolFlag(constants.ArgModInstall, "", true, "Specify whether to install mod dependencies before running the dashboard")
	return cmd
}

func runDashboardCmd(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(cmd.Context())
	contexthelpers.StartCancelHandler(cancel)

	logging.LogTime("runDashboardCmd start")
	defer func() {
		logging.LogTime("runDashboardCmd end")
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
		}
	}()

	serverPort := dashboardserver.ListenPort(viper.GetInt(constants.ArgDashboardServerPort))
	utils.FailOnError(serverPort.IsValid())

	serverListen := dashboardserver.ListenType(viper.GetString(constants.ArgDashboardServerListen))
	utils.FailOnError(serverListen.IsValid())

	// ensure dashboard assets are present and extract if not
	err := dashboardassets.Ensure(ctx)
	utils.FailOnError(err)

	server, err := dashboardserver.NewServer(ctx, dbClient)
	if err != nil {
		utils.FailOnError(err)
	}

	server.Start()

	// wait for the given context to cancel
	<-ctx.Done()

	server.Shutdown(ctx)
}
