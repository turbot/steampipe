package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/report/reportserver"
	"github.com/turbot/steampipe/utils"
)

// ReportCmd :: represents the report command
func ReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "report [report]",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runReportCmd,
		Short:            "Run a report",
		Long:             `Run a report...TODO better description!`,
	}

	cmdconfig.OnCmd(cmd)
	return cmd
}

func runReportCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runReportCmd start")
	defer func() {
		logging.LogTime("runReportCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, false)

	ctx, cancel := context.WithCancel(context.Background())
	startCancelHandler(cancel)

	// start db if necessary
	err := db.EnsureDbAndStartService(db.InvokerReport)
	utils.FailOnErrorWithMessage(err, "failed to start service")
	defer db.Shutdown(nil, db.InvokerReport)

	server, err := reportserver.NewServer(ctx)

	if err != nil {
		utils.FailOnError(err)
	}

	defer server.Shutdown()

	server.Start()
}
