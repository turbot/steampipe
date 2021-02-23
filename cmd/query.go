package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/display"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/utils"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(QueryCmd())
}

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
		AddBoolFlag(constants.ArgTimer, "", false, "Turn on the timer which reports query time.")

	return cmd
}

func runQueryCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runQueryCmd start")
	defer logging.LogTime("execute end")

	log.Println("[TRACE] runQueryCmd")
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			utils.ShowError(err)
			os.Exit(1)
		}
	}()

	queryString := getQuery(cmd, args)

	// set a global viper config key so that we know where we are at
	cmdconfig.Viper().Set("query-cmd", true)
	if queryString == "" {
		cmdconfig.Viper().Set("interactive", true)
	}

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db.ExecuteQuery(queryString)
	utils.FailOnError(err)

	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		//signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.Done()
	}
}

// retrieve query from args or determine whether to run the interactive shell
func getQuery(cmd *cobra.Command, args []string) (query string) {
	log.Println("[TRACE] getQuery")
	if len(args) == 0 {
		if cmdconfig.Viper().GetBool(constants.ArgListAllTableNames) {
			query = ".tables"
		} else if table := cmdconfig.Viper().GetString(constants.ArgSelectAll); table != "" {
			query = fmt.Sprintf("select * from %s", table)
		}
	} else {
		query = args[0]
	}
	return
}
