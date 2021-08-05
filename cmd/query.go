package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/interactive"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/turbot/steampipe/query/queryexecute"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

// queryCmd :: represents the query command
func queryCmd() *cobra.Command {
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
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", []string{}, "Set a prefix to the current search path for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgVarFile, "", []string{}, "Specify a file containing variable values").
		AddStringSliceFlag(constants.ArgVariable, "", []string{}, "Specify The value of a variable")
	return cmd
}

// getPipedStdinData reads the Standard Input and returns the available data as a string
// if and only if the data was piped to the process
func getPipedStdinData() string {
	fi, err := os.Stdin.Stat()
	if err != nil {
		utils.ShowWarning("could not fetch information about STDIN")
		return ""
	}
	stdinData := ""
	if (fi.Mode()&os.ModeCharDevice) == 0 && fi.Size() > 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdinData = fmt.Sprintf("%s%s", stdinData, scanner.Text())
		}
	}
	return stdinData
}

func runQueryCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runQueryCmd start")

	defer func() {
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

	ctx := context.Background()

	// load the workspace
	w, err := loadWorkspacePromptingForVariables(ctx)
	utils.FailOnErrorWithMessage(err, "failed to load workspace")

	// se we have loaded a workspace - be sure to close it
	defer w.Close()

	// perform rest of initialisation async
	initDataChan := make(chan *db_common.QueryInitData, 1)
	getQueryInitDataAsync(ctx, w, initDataChan, args)

	if interactiveMode {
		queryexecute.RunInteractiveSession(&initDataChan)
	} else {
		ctx, cancel := context.WithCancel(ctx)
		startCancelHandler(cancel)
		// set global exit code
		exitCode = queryexecute.RunBatchSession(ctx, initDataChan)

	}

}

func loadWorkspacePromptingForVariables(ctx context.Context) (*workspace.Workspace, error) {
	w, err := workspace.Load(viper.GetString(constants.ArgWorkspace))
	if err == nil {
		return w, nil
	}
	// if there was an erro which is NOT a MissingVariableError, return it
	missingVariablesError, ok := err.(modconfig.MissingVariableError)
	if !ok {
		return nil, err
	}
	if err := interactive.PromptForMissingVariables(ctx, missingVariablesError.MissingVariables); err != nil {
		log.Printf("[TRACE] Interactive variables prompting returned error %v", err)
		return nil, nil
	}
	// ok we should have all variables now - reload workspace
	return workspace.Load(viper.GetString(constants.ArgWorkspace))
}

func getQueryInitDataAsync(ctx context.Context, workspace *workspace.Workspace, initDataChan chan *db_common.QueryInitData, args []string) {
	go func() {
		initData := db_common.NewQueryInitData()
		defer func() {
			if r := recover(); r != nil {
				initData.Result.Error = helpers.ToError(r)
			}
			initDataChan <- initData
			close(initDataChan)
		}()

		// get a db client
		client, err := db.GetClient(constants.InvokerQuery)
		if err != nil {
			initData.Result.Error = err
			return
		}

		// check if the required plugins are installed
		if err := workspace.CheckRequiredPluginsInstalled(); err != nil {
			initData.Result.Error = err
			return
		}
		initData.Workspace = workspace

		// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
		initData.Queries = queryexecute.GetQueries(args, workspace)

		res := client.RefreshConnectionAndSearchPaths()
		if res.Error != nil {
			initData.Result.Error = res.Error
			return
		}
		initData.Result.AddWarnings(res.Warnings...)

		initData.Client = client
		// populate the reflection tables
		if err = db_common.CreateMetadataTables(ctx, workspace.GetResourceMaps(), client); err != nil {
			initData.Result.Error = err
			return
		}
	}()
}

func startCancelHandler(cancel context.CancelFunc) {
	sigIntChannel := make(chan os.Signal, 1)
	signal.Notify(sigIntChannel, os.Interrupt)
	go func() {
		<-sigIntChannel
		// call context cancellation function
		cancel()
		// leave the channel open - any subsequent interrupts hits will be ignored
	}()
}
