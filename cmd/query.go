package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/turbot/steampipe/query/execute"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

// QueryCmd :: represents the query command
func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "query",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runQueryCmd,
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

func runQueryCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runQueryCmd start")
	var client *db.Client
	defer func() {
		utils.LogTime("runQueryCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	// enable spinner only in interactive mode
	interactiveMode := len(args) == 0
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, interactiveMode)
	// set config to indicate whether we are running an interactive query
	viper.Set(constants.ConfigKeyInteractive, interactiveMode)

	// start db if necessary
	err := db.EnsureDbAndStartService(db.InvokerQuery)
	utils.FailOnErrorWithMessage(err, "failed to start service")
	defer func() {
		db.Shutdown(client, db.InvokerQuery)
	}()

	// load the workspace
	workspace, err := workspace.Load(viper.GetString(constants.ArgWorkspace))
	utils.FailOnErrorWithMessage(err, "failed to load workspace")
	defer workspace.Close()

	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	queries := execute.GetQueries(args, workspace)

	// get a db client
	client, err = db.NewClient(true)
	utils.FailOnError(err)

	// populate the reflection tables
	err = db.CreateMetadataTables(workspace.GetResourceMaps(), client)
	utils.FailOnError(err)

	// if no query is specified, run interactive prompt
	if interactiveMode {
		// interactive session creates its own client
		execute.RunInteractiveSession(workspace, client)
	} else if len(queries) > 0 {
		// ensure client is closed
		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		startCancelHandler(cancel)
		// otherwise if we have resolved any queries, run them
		failures := execute.ExecuteQueries(ctx, queries, client)
		// set global exit code
		exitCode = failures
	}
}

func startCancelHandler(cancel context.CancelFunc) {
	sigIntChannel := make(chan os.Signal, 1)
	signal.Notify(sigIntChannel, os.Interrupt)
	go func() {
		<-sigIntChannel
		cancel()
		close(sigIntChannel)
	}()
}
