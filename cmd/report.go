package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/contexthelpers"
	"github.com/turbot/steampipe/report/reportserver"
	"github.com/turbot/steampipe/utils"
)

func reportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "report [report]",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runReportCmd,
		Short:            "Run a report",
		Long:             `Run a report...TODO better description!`,
	}

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for report")
	return cmd
}

func runReportCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logging.LogTime("runReportCmd start")
	defer func() {
		logging.LogTime("runReportCmd end")
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
		}
	}()

	ctx, cancel := context.WithCancel(cmd.Context())
	contexthelpers.StartCancelHandler(cancel)

	// start db if necessary
	//err := db_local.EnsureDbAndStartService(constants.InvokerReport, true)
	//utils.FailOnErrorWithMessage(err, "failed to start service")
	//defer db_local.ShutdownService(constants.InvokerReport)

	server, err := reportserver.NewServer(ctx)

	if err != nil {
		utils.FailOnError(err)
	}

	defer server.Shutdown(ctx)

	server.Start()
}
