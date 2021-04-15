package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/display"
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
		AddBoolFlag(constants.ArgWatch, "", false, "Watch SQL files in the current workspace (works only in interactive mode)").
		AddStringSliceFlag(constants.ArgSearchPath, "", []string{}, "Set a custom search_path for the steampipe user for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", []string{}, "Set a prefix to the current search path for a query session (comma-separated)")

	return cmd
}

func runQueryCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runQueryCmd start")

	defer func() {
		logging.LogTime("runQueryCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	// start db if necessary
	err := db.StartServiceForQuery()
	utils.FailOnErrorWithMessage(err, "failed to start service")
	defer db.Shutdown(nil, db.InvokerQuery)

	// load the workspace (do not do this until after service start as watcher interferes with service start)
	workspace, err := workspace.Load(viper.GetString(constants.ArgWorkspace))
	utils.FailOnErrorWithMessage(err, "failed to load workspace")
	defer workspace.Close()

	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	queries := getQueries(args, workspace)

	// if no query is specified, run interactive prompt
	if len(args) == 0 {
		// interactive session creates its own client
		runInteractiveSession(workspace)
	} else if len(queries) > 0 {
		// otherwsie if we have resolvced any queries, run them
		failures := executeQueries(queries)
		// set global exit code
		exitCode = failures
	}
}

// retrieve queries from args - for each arg check if it is a named query or a file,
// before falling back to treating it as sql
func getQueries(args []string, workspace *workspace.Workspace) []string {
	var queries []string
	for _, arg := range args {
		if namedQuery, ok := workspace.GetNamedQuery(arg); ok {
			queries = append(queries, namedQuery.SQL)
			continue
		}
		fileQuery, fileExists, err := getQueryFromFile(arg)
		if fileExists {
			if err != nil {
				utils.ShowWarning(fmt.Sprintf("error opening file '%s': %v", arg, err))
			} else if len(fileQuery) == 0 {
				utils.ShowWarning(fmt.Sprintf("file '%s' does not contain any data", arg))
			} else {
				queries = append(queries, fileQuery)
			}
			continue
		}

		queries = append(queries, arg)
	}

	return queries
}

func runInteractiveSession(workspace *workspace.Workspace) {
	// start the workspace file watcher
	if viper.GetBool(constants.ArgWatch) {
		err := workspace.SetupWatcher()
		utils.FailOnError(err)
	}

	// set the flag to not show spinner
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, true)

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db.RunInteractivePrompt(workspace)
	utils.FailOnError(err)

	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		// signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.Done()
	}
}

func executeQueries(queries []string) int {
	// set the flag to show spinner
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, false)

	// first get a client - do this once for all queries
	client, err := db.NewClient(true)
	utils.FailOnError(err)
	defer client.Close()

	// run all queries
	failures := 0
	for _, q := range queries {
		if err := runQuery(q, client); err != nil {
			failures++
			utils.ShowWarning(fmt.Sprintf("query '%s' failed: %v", q, err))
		}
		fmt.Println()
	}

	return failures
}

func runQuery(queryString string, client *db.Client) error {
	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db.ExecuteQuery(queryString, client)
	if err != nil {
		return err
	}

	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		// signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.Done()
	}
	return nil
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
