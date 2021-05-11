package cmd

import (
	"github.com/turbot/steampipe/query/execute"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/logging"
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
	logging.LogTime("runQueryCmd start")
	var client *db.Client
	defer func() {
		logging.LogTime("runQueryCmd end")
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

// retrieve queries from args - for each arg check if it is a named query or a file,
// before falling back to treating it as sql
func getQueries(args []string, workspace *workspace.Workspace) []string {
	var queries []string
	for _, arg := range args {
		query, _ := getQueryFromArg(arg, workspace)
		if len(query) > 0 {
			queries = append(queries, query)
		}
	}
	return queries
}

// attempt to resolve 'arg' to a query
// if the arg was a named query or a sql file, return 'true for the second return value
func getQueryFromArg(arg string, workspace *workspace.Workspace) (string, bool) {
	// 1) is this a named query
	if namedQuery, ok := workspace.GetNamedQuery(arg); ok {
		return typeHelpers.SafeString(namedQuery.SQL), true
	}

	// 	2) is this a file
	fileQuery, fileExists, err := getQueryFromFile(arg)
	if fileExists {
		if err != nil {
			utils.ShowWarning(fmt.Sprintf("error opening file '%s': %v", arg, err))
			return "", false
		}
		if len(fileQuery) == 0 {
			utils.ShowWarning(fmt.Sprintf("file '%s' does not contain any data", arg))
			// (just return the empty string - it will be filtered above)
		}
		return fileQuery, true
	}

	// 3) just use the arg string as is and assume it is valid SQL
	return arg, false
}

func runInteractiveSession(workspace *workspace.Workspace, client *db.Client) {
	// start the workspace file watcher
	if viper.GetBool(constants.ArgWatch) {
		err := workspace.SetupWatcher(client)
		utils.FailOnError(err)
	}

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db.RunInteractivePrompt(workspace, client)
	utils.FailOnError(err)

	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		// signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.Done()
	}
}

func executeQueries(ctx context.Context, queries []string, client *db.Client) int {
	// run all queries
	failures := 0
	for i, q := range queries {
		select {
		case <-ctx.Done():
			// add to failures
			failures++
			// skip ahead to the end
			continue
		default:
			if err := executeQuery(ctx, q, client); err != nil {
				failures++
				utils.ShowWarning(fmt.Sprintf("executeQueries: query %d of %d failed: %v", i+1, len(queries), utils.TrimDriversFromErrMsg(err.Error())))
			}
			if showBlankLineBetweenResults() {
				fmt.Println()
			}
		}
	}

	return failures
}

func executeQuery(ctx context.Context, queryString string, client *db.Client) error {
	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db.ExecuteQuery(ctx, queryString, client)
	if err != nil {
		return err
	}

	// TODO encapsulate this in display object
	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		// signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.Done()
	}
	return nil
}

// if we are displaying csv with no header, do not include lines between the query results
func showBlankLineBetweenResults() bool {
	return !(viper.GetString(constants.ArgOutput) == "csv" && !viper.GetBool(constants.ArgHeader))
}

func getQueryFromFile(filename string) (string, bool, error) {
	log.Println("[TRACE] getQueryFromFiles: ", filename)

	// get absolute filename
	path, err := filepath.Abs(filename)
	if err != nil {
		return "", false, nil
	}
	// does it exist?
	if _, err := os.Stat(path); err != nil {
		// if this gives any error, return not exist. we may get a not found or a path too long for example
		return "", false, nil
	}

	// read file
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", true, err
	}

	return string(fileBytes), true, nil
}
