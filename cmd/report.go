package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/contexthelpers"
	"github.com/turbot/steampipe/report/reportassets"
	"github.com/turbot/steampipe/report/reportserver"
	"github.com/turbot/steampipe/utils"
)

func reportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "report",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runReportCmd,
		Short:            "Start the local report UI",
		Long: `Starts a local web server that enables real-time development of reports within the current mod.

The current mod is the working directory, or the directory specified by the --workspace-chdir flag.`,
	}

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for report").
		AddIntFlag(constants.ArgReportServerPort, "", constants.ReportServerDefaultPort, "Report server port.")
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

	// ensure report assets are present and extract if not
	err := reportassets.Ensure(ctx)
	utils.FailOnError(err)

	ctx, cancel := context.WithCancel(cmd.Context())
	contexthelpers.StartCancelHandler(cancel)

	server, err := reportserver.NewServer(ctx)
	if err != nil {
		utils.FailOnError(err)
	}

	defer server.Shutdown(ctx)

	server.Start()

}
