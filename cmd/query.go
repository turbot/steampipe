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
		AddStringSliceFlag(constants.ArgSqlFile, "", nil, "Specifies one or more sql files to execute.").
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

	//log.Printf("[WARN] start service")
	// start db if necessary
	err := db.StartServiceForQuery()
	utils.FailOnErrorWithMessage(err, "failed to start service")
	defer db.Shutdown(nil, db.InvokerQuery)

	//log.Printf("[WARN] load workspace")
	// load the workspace (do not do this until after service start as watcher interferes with service start)
	workspace, err := workspace.Load(viper.GetString(constants.ArgWorkspace))

	utils.FailOnErrorWithMessage(err, "failed to load workspace")
	defer workspace.Close()

	//log.Printf("[WARN] get queries")

	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	queries, err := getQueries(args, workspace)
	utils.FailOnError(err)

	// if no query is specified, run interactive prompt
	if len(queries) == 0 {
		// interactive session creates its own client
		runInteractiveSession(workspace)
	} else {
		executeQueries(queries)
	}
}

// retrieve queries from args or determine whether to run the interactive shell
func getQueries(args []string, workspace *workspace.Workspace) ([]string, error) {
	// was the sql-file flag used?
	if sqlFiles := viper.GetStringSlice(constants.ArgSqlFile); len(sqlFiles) > 0 {
		// cobra only takes the first string after a flag as the flag value, so if more than one file is specified,
		// and they are NOT comma separated, all but the first file will appear in 'args'
		// instead of being assigned to the sql-file flag - so append args to the list of files
		// NOTE: this does mean if there are any other unclaimed args, they will be treated as a file
		// (and probably cause a not-exists error)
		sqlFiles = append(sqlFiles, args...)
		return getQueriesFromFiles(sqlFiles)
	}

	// otherwise either the query was passed as an argument, or no query was passed (interactive mode)
	// just return the first arg (if there is one)

	// if no query is specified in the args, we must pass a single empty query to trigger interactive mode
	var queries []string
	if len(args) > 0 {
		if namedQuery, ok := workspace.GetNamedQuery(args[0]); ok {
			queries = []string{namedQuery.SQL}
		} else {
			queries = []string{args[0]}
		}
	}
	return queries, nil
}

func runInteractiveSession(workspace *workspace.Workspace) {
	// set the flag to not show spinner
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, false)

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

func executeQueries(queries []string) {
	// first get a client - do this once for all queries
	client, err := db.NewClient(true)
	utils.FailOnError(err)
	defer client.Close()

	// run all queries
	for _, q := range queries {
		runQuery(q, client)
	}
}

func runQuery(queryString string, client *db.Client) {
	// set the flag to show spinner
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, true)

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db.ExecuteQuery(queryString, client)
	utils.FailOnError(err)

	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		//signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.Done()
	}
}

func getQueriesFromFiles(files []string) ([]string, error) {
	log.Println("[TRACE] getQueriesFromFiles: ", files)
	// build list of queries
	var result []string
	for _, filename := range files {
		// get absolute filename
		path, err := filepath.Abs(filename)
		if err != nil {
			return nil, err
		}
		// does it exist?
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("file '%s' does not exist", path)
		}

		// read file
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if len(fileBytes) == 0 {
			// empty file - ignore
			continue
		}

		// add to list of queries
		result = append(result, string(fileBytes))
	}
	return result, nil
}
