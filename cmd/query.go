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

	// load the workspace
	workspace, err := workspace.Load(viper.GetString(constants.ArgWorkspace))
	utils.FailOnError(err)
	defer workspace.Close()

	// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
	queries, err := getQueries(args, workspace)
	utils.FailOnError(err)

	for _, q := range queries {
		runQuery(q, workspace)
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
	var query = ""
	if len(args) > 0 {
		if namedQuery, ok := workspace.GetNamedQuery(args[0]); ok {
			query = namedQuery.SQL
		} else {
			query = args[0]
		}
	}
	return []string{query}, nil
}

func runQuery(queryString string, workspace *workspace.Workspace) {
	// set the flag to not show spinner
	showSpinner := queryString == ""
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, showSpinner)

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db.ExecuteQuery(queryString, workspace)
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
