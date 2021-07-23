package cmd

import (
	"context"

	"github.com/turbot/steampipe/db/local_db"

	"github.com/turbot/steampipe/query/queryexecute"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// queryCmd :: represents the query command
func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "query2",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runNewQueryCmd,
		Short:            "Execute SQL queries interactively or by argument",
		Long: `Execute SQL queries interactively, or by a query argument.

Open a interactive SQL query console to Steampipe to explore your data and run
multiple queries. If QUERY is passed on the command line then it will be run
immediately and the command will exit.

Examples:

 # Open an interactive query console
 steampipe query

 # Run a specific query directly
 steampipe query "select * from cloud"`,
	}

	// Notes:
	// * In the future we may add --csv and --json flags as shortcuts for --output
	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgHeader, "", true, "Include column headers csv and table output").
		AddStringFlag(constants.ArgSeparator, "", ",", "Separator string for csv output").
		AddStringFlag(constants.ArgOutput, "", "table", "Output format: line, csv, json or table").
		AddBoolFlag(constants.ArgTimer, "", false, "Turn on the timer which reports query time.").
		AddBoolFlag(constants.ArgWatch, "", true, "Watch SQL files in the current workspace (works only in interactive mode)").
		AddStringSliceFlag(constants.ArgSearchPath, "", []string{}, "Set a custom search_path for the steampipe user for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", []string{}, "Set a prefix to the current search path for a query session (comma-separated)")

	return cmd
}

func runNewQueryCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runQueryCmd start")
	var client *local_db.LocalClient

	defer func() {
		// ensure client is closed after we are done
		// (it will only be non-null for non-interactive queries - interactive close their own client)
		local_db.Shutdown(client, local_db.InvokerQuery)

		utils.LogTime("cmd.runQueryCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	if stdinData := getPipedStdinData(); len(stdinData) > 0 {
		args = append(args, stdinData)
	}

	// enable spinner only in interactive mode
	interactiveMode := len(args) == 0
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, interactiveMode)
	// set config to indicate whether we are running an interactive query
	viper.Set(constants.ConfigKeyInteractive, interactiveMode)

	// start db if necessary - do not refresh connections as we do it as part of the async startup
	err := local_db.EnsureDbAndStartService(local_db.InvokerQuery, false)
	utils.FailOnErrorWithMessage(err, "failed to start service")

	ctx, cancel := context.WithCancel(context.Background())
	startCancelHandler(cancel)

	// perform rest of initialisation async
	initDataChan := make(chan *local_db.QueryInitData, 1)
	getQueryInitDataAsync(ctx, initDataChan, args)

	if interactiveMode {
		queryexecute.RunInteractiveSession(&initDataChan)
		return
	}

	// wait for init
	initData := <-initDataChan
	if err := initData.Result.Error; err != nil {
		utils.FailOnError(err)
	}
	// check for cancellation
	utils.FailOnError(utils.HandleCancelError(ctx.Err()))

	// display any initialisation messages/warnings
	initData.Result.DisplayMessages()
	// populate client so it gets closed by defer
	client = initData.Client

	if len(initData.Queries) > 0 {
		// otherwise if we have resolved any queries, run them
		failures := queryexecute.ExecuteQueries(ctx, initData.Queries, initData.Client)
		// set global exit code
		exitCode = failures
	}

}
