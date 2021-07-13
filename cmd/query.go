package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
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
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", []string{}, "Set a prefix to the current search path for a query session (comma-separated)")

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
	var client *db.Client

	defer func() {
		// ensure client is closed after we are done
		// (it will only be non-null for non-interactive queries - interactive close their own client)
		db.Shutdown(client, db.InvokerQuery)

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

	// start db if necessary
	err := db.EnsureDbAndStartService(db.InvokerQuery)
	utils.FailOnErrorWithMessage(err, "failed to start service")

	// perform rest of initialisation async
	initDataChan := make(chan *db.QueryInitData, 1)
	getQueryInitDataAsync(initDataChan, args)

	if interactiveMode {
		execute.RunInteractiveSession(&initDataChan)
	} else {

		// wait for init
		initData := <-initDataChan
		// populate client so it gets closed by defer
		client = initData.Client
		HandleInitResult(initData)

		// if no query is specified, run interactive prompt
		if !interactiveMode && len(initData.Queries) > 0 {

			ctx, cancel := context.WithCancel(context.Background())
			startCancelHandler(cancel)
			// otherwise if we have resolved any queries, run them
			failures := execute.ExecuteQueries(ctx, initData.Queries, initData.Client)
			// set global exit code
			exitCode = failures
		}
	}
}

func HandleInitResult(d *db.QueryInitData) {
	// check for error and warnings
	if d.Result.Error != nil {
		utils.FailOnError(d.Result.Error)
	}
	for _, warning := range d.Result.Warnings {
		fmt.Println(warning)
	}
	for _, message := range d.Result.Messages {
		fmt.Println(message)
	}
}
func getQueryInitDataAsync(initDataChan chan *db.QueryInitData, args []string) {
	go func() {
		log.Printf("[TRACE] getQueryInitDataAsync")

		initData := db.NewInitData()
		defer func() {
			initDataChan <- initData
			close(initDataChan)
			log.Printf("[TRACE] getQueryInitDataAsync complete")
		}()

		// load the workspace
		workspace, err := workspace.Load(viper.GetString(constants.ArgWorkspace))
		if err != nil {
			initData.Result.Error = utils.PrefixError(err, "failed to load workspace")
			return
		}

		// se we have loaded a workspace - be sure to close it
		defer workspace.Close()

		// check if the required plugins are installed
		if err := workspace.CheckRequiredPluginsInstalled(); err != nil {
			initData.Result.Error = err
			return
		}
		initData.Workspace = workspace

		// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
		initData.Queries = execute.GetQueries(args, workspace)

		// get a db client
		client, res := db.NewClient(true)
		if initData.Result.Error != nil {
			initData.Result.Error = res.Error
			return
		}
		if len(res.Warning) > 0 {
			initData.Result.AddWarning(res.Warning)
		}

		initData.Client = client
		// populate the reflection tables
		if err = db.CreateMetadataTables(workspace.GetResourceMaps(), client); err != nil {
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
		cancel()
		close(sigIntChannel)
	}()
}
