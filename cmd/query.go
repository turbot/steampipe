package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_client"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/interactive"
	"github.com/turbot/steampipe/query/queryexecute"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

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

		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			workspace, err := workspace.LoadResourceNames(viper.GetString(constants.ArgWorkspace))
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveError
			}
			namedQueries := []string{}
			for _, name := range workspace.GetSortedNamedQueryNames() {
				if strings.HasPrefix(name, toComplete) {
					namedQueries = append(namedQueries, name)
				}
			}
			return namedQueries, cobra.ShellCompDirectiveNoFileComp
		},
	}

	// Notes:
	// * In the future we may add --csv and --json flags as shortcuts for --output
	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for query").
		AddBoolFlag(constants.ArgHeader, "", true, "Include column headers csv and table output").
		AddStringFlag(constants.ArgSeparator, "", ",", "Separator string for csv output").
		AddStringFlag(constants.ArgOutput, "", "table", "Output format: line, csv, json or table").
		AddBoolFlag(constants.ArgTimer, "", false, "Turn on the timer which reports query time.").
		AddBoolFlag(constants.ArgWatch, "", true, "Watch SQL files in the current workspace (works only in interactive mode)").
		AddStringSliceFlag(constants.ArgSearchPath, "", nil, "Set a custom search_path for the steampipe user for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", nil, "Set a prefix to the current search path for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgVarFile, "", nil, "Specify a file containing variable values").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, "", nil, "Specify the value of a variable")
	return cmd
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

	err := validateConnectionStringArgs()
	utils.FailOnError(err)

	// enable spinner only in interactive mode
	interactiveMode := len(args) == 0
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, interactiveMode)
	// set config to indicate whether we are running an interactive query
	viper.Set(constants.ConfigKeyInteractive, interactiveMode)

	ctx := context.Background()

	// load the workspace
	w, err := loadWorkspacePromptingForVariables(ctx, nil)
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

func validateConnectionStringArgs() error {
	backendEnvVar, backendDefined := os.LookupEnv(constants.EnvDatabaseBackend)

	if !backendDefined {
		// no database set - so no connection string
		return nil
	}
	connectionString := backendEnvVar

	// so a backend was set - is it a connection string or a database name
	if !strings.HasPrefix(backendEnvVar, "postgresql://") {
		// it must be a database name - verify the cloud token was provided
		cloudToken, gotCloudToken := os.LookupEnv(constants.EnvCloudToken)
		if !gotCloudToken {
			return fmt.Errorf("if %s is set as a workspace name, %s must be set", constants.EnvDatabaseBackend, constants.EnvCloudToken)
		}

		// so we have a database and a token - build the connection string and set it in viper
		var err error
		if connectionString, err = db_common.GetConnectionString(backendEnvVar, cloudToken); err != nil {
			return err
		}
	}

	// now set the connection string in viper
	viper.Set(constants.ArgConnectionString, connectionString)

	return nil
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

func loadWorkspacePromptingForVariables(ctx context.Context, spinner *spinner.Spinner) (*workspace.Workspace, error) {
	workspacePath := viper.GetString(constants.ArgWorkspace)

	w, err := workspace.Load(workspacePath)
	if err == nil {
		return w, nil
	}
	missingVariablesError, ok := err.(modconfig.MissingVariableError)
	// if there was an error which is NOT a MissingVariableError, return it
	if !ok {
		return nil, err
	}
	if spinner != nil {
		spinner.Stop()
	}
	// so we have missing variables - prompt for them
	if err := interactive.PromptForMissingVariables(ctx, missingVariablesError.MissingVariables); err != nil {
		log.Printf("[TRACE] Interactive variables prompting returned error %v", err)
		return nil, err
	}
	if spinner != nil {
		spinner.Start()
	}
	// ok we should have all variables now - reload workspace
	return workspace.Load(workspacePath)
}

func getQueryInitDataAsync(ctx context.Context, w *workspace.Workspace, initDataChan chan *db_common.QueryInitData, args []string) {
	go func() {
		utils.LogTime("cmd.getQueryInitDataAsync start")
		defer utils.LogTime("cmd.getQueryInitDataAsync end")
		initData := db_common.NewQueryInitData()
		defer func() {
			if r := recover(); r != nil {
				initData.Result.Error = helpers.ToError(r)
			}
			initDataChan <- initData
			close(initDataChan)
		}()

		// set max DB connections to 1
		viper.Set(constants.ArgMaxParallel, 1)
		// get a db client
		var client db_common.Client
		var err error
		if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
			client, err = db_client.NewDbClient(connectionString)
		} else {
			client, err = db_local.GetLocalClient(constants.InvokerQuery)
		}
		if err != nil {
			initData.Result.Error = err
			return
		}
		initData.Client = client

		// check if the required plugins are installed
		if err := w.CheckRequiredPluginsInstalled(); err != nil {
			initData.Result.Error = err
			return
		}
		initData.Workspace = w

		// convert the query or sql file arg into an array of executable queries - check names queries in the current workspace
		queries, _, err := w.GetQueriesFromArgs(args)
		if err != nil {
			initData.Result.Error = err
			return
		}
		initData.Queries = queries

		res := client.RefreshConnectionAndSearchPaths()
		if res.Error != nil {
			initData.Result.Error = res.Error
			return
		}
		initData.Result.AddWarnings(res.Warnings...)

		//// set up the session data - prepared statements and introspection tables
		//// this defaults to creating prepared statements for all queries
		//sessionDataSource := workspace.NewSessionDataSource(w.GetResourceMaps())
		//// if queries were provided as args, only create prepared statements required for these queries
		//if len(queries) > 0 {
		//	log.Printf("[TRACE] only creating prepared statements for command line queries")
		//	sessionDataSource.PreparedStatementSource = preparedStatementSource
		//}

		// register EnsureSessionData as a callback on the client.
		// if the underlying SQL client has certain errors (for example context expiry) it will reset the session
		// so our client object calls this callback to restore the session data
		initData.Client.SetEnsureSessionDataFunc(func(ctx context.Context, session *db_common.DatabaseSession) (error, []string) {
			// TODO only create for queries
			return workspace.EnsureSessionData(ctx, w, session)
		})

		// force creation of session data - se we see any prepared statement errors at once
		session, err, warnings := initData.Client.AcquireSession(ctx)
		initData.Result.AddWarnings(warnings...)
		if err != nil {
			initData.Result.Error = fmt.Errorf("error acquiring database connection, %s", err.Error())
		} else {
			session.Close()
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
