package cmd

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"golang.org/x/exp/maps"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardexecute"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/query"
	"github.com/turbot/steampipe/pkg/query/queryexecute"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/workspace"
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
			w, err := workspace.LoadResourceNames(viper.GetString(constants.ArgWorkspaceChDir))
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveError
			}
			namedQueries := []string{}
			for _, name := range w.GetSortedNamedQueryNames() {
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
		AddStringFlag(constants.ArgOutput, "", "table", "Output format: line, csv, json, table or snapshot").
		AddBoolFlag(constants.ArgTiming, "", false, "Turn on the timer which reports query time.").
		AddBoolFlag(constants.ArgWatch, "", true, "Watch SQL files in the current workspace (works only in interactive mode)").
		AddStringSliceFlag(constants.ArgSearchPath, "", nil, "Set a custom search_path for the steampipe user for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", nil, "Set a prefix to the current search path for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgVarFile, "", nil, "Specify a file containing variable values").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, "", nil, "Specify the value of a variable").
		AddBoolFlag(constants.ArgInput, "", true, "Enable interactive prompts").
		AddStringFlag(constants.ArgSnapshot, "", "", "Create snapshot in Steampipe Cloud with the default (workspace) visibility.", cmdconfig.FlagOptions.NoOptDefVal(constants.ArgShareNoOptDefault)).
		AddStringFlag(constants.ArgShare, "", "", "Create snapshot in Steampipe Cloud with 'anyone_with_link' visibility.", cmdconfig.FlagOptions.NoOptDefVal(constants.ArgShareNoOptDefault)).
		AddStringArrayFlag(constants.ArgSnapshotTag, "", nil, "Specify the value of a tag to set on the snapshot").
		AddStringFlag(constants.ArgWorkspace, "", "", "The cloud workspace... ")

	return cmd
}

func runQueryCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runQueryCmd start")
	defer func() {
		utils.LogTime("cmd.runQueryCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
		}
	}()

	if stdinData := getPipedStdinData(); len(stdinData) > 0 {
		args = append(args, stdinData)
	}

	// validate args
	error_helpers.FailOnError(validateQueryArgs(cmd))

	cloudMetadata, err := cmdconfig.GetCloudMetadata()
	error_helpers.FailOnError(err)

	// enable spinner only in interactive mode
	interactiveMode := len(args) == 0
	// set config to indicate whether we are running an interactive query
	viper.Set(constants.ConfigKeyInteractive, interactiveMode)

	// load the workspace
	w, err := interactive.LoadWorkspacePromptingForVariables(ctx)
	error_helpers.FailOnErrorWithMessage(err, "failed to load workspace")

	// set cloud metadata (may be nil)
	w.CloudMetadata = cloudMetadata

	// so we have loaded a workspace - be sure to close it
	defer w.Close()

	// start the initializer
	initData := query.NewInitData(ctx, w, args)

	switch {
	case interactiveMode:
		queryexecute.RunInteractiveSession(ctx, initData)
	case snapshotRequired():
		// if we are either outputting snapshot format, or sharing the results as a snapshot, execute the query
		// as a dashboard

		// if display is not explicitly set, set to none
		if !cmdconfig.FlagSetByUser(cmd, constants.ArgOutput) {
			viper.Set(constants.ArgOutput, constants.OutputFormatNone)
		}

		exitCode = executeSnapshotQuery(initData, w, ctx)

	default:
		// NOTE: disable any status updates - we do not want 'loading' output from any queries
		ctx = statushooks.DisableStatusHooks(ctx)

		// fall through to running a batch query
		// set global exit code
		exitCode = queryexecute.RunBatchSession(ctx, initData)
	}
}

func validateQueryArgs(cmd *cobra.Command) error {
	err := validateSnapshotArgs()
	if err != nil {
		return err
	}

	validOutputFormats := []string{constants.OutputFormatLine, constants.OutputFormatCSV, constants.OutputFormatTable, constants.OutputFormatJSON, constants.OutputFormatSnapshot}
	if !helpers.StringSliceContains(validOutputFormats, viper.GetString(constants.ArgOutput)) {
		return fmt.Errorf("invalid output format, must be one of %s", strings.Join(validOutputFormats, ","))
	}

	// if workspace-database has not been explicitly set, check whether workspace has been set
	// and if so use that
	if !cmdconfig.FlagSetByUser(cmd, constants.ArgWorkspaceDatabase) {
		if w := viper.GetString(constants.ArgWorkspace); w != "" {
			viper.Set(constants.ArgWorkspace, w)
		}
	}

	return nil
}

func executeSnapshotQuery(initData *query.InitData, w *workspace.Workspace, ctx context.Context) int {
	// ensure we close client
	defer initData.Cleanup(ctx)

	// start cancel handler to intercept interrupts and cancel the context
	// NOTE: use the initData Cancel function to ensure any initialisation is cancelled if needed
	contexthelpers.StartCancelHandler(initData.Cancel)

	// wait for init
	<-initData.Loaded
	if err := initData.Result.Error; err != nil {
		utils.FailOnError(err)
	}

	// build ordered list of queries
	// (ordered for testing repeatability)
	var queryNames []string = utils.SortedMapKeys(initData.Queries)

	if len(queryNames) > 0 {
		for i, name := range queryNames {
			query := initData.Queries[name]
			// if a manual query is being run (i.e. not a named query), convert into a query and add to workspace
			// this is to allow us to use existing dashboard execution code
			targetName := ensureQueryResource(name, query, i, w)

			// we need to pass the embedded initData to  GenerateSnapshot
			baseInitData := &initData.InitData

			// so a dashboard name was specified - just call GenerateSnapshot
			snap, err := dashboardexecute.GenerateSnapshot(ctx, targetName, baseInitData, nil)
			utils.FailOnError(err)

			// display the result
			// if the format is snapshot, just dump it out
			if viper.GetString(constants.ArgOutput) == constants.OutputFormatSnapshot {
				jsonOutput, err := json.MarshalIndent(snap, "", "  ")
				if err != nil {
					utils.FailOnErrorWithMessage(err, "failed to display result as snapshot")
				}
				fmt.Println(string(jsonOutput))
			} else {
				// otherwise convert the snapshot into a query result
				result, err := snapshotToQueryResult(snap, targetName)
				utils.FailOnErrorWithMessage(err, "failed to display result as snapshot")
				display.ShowOutput(ctx, result)
			}

			// share the snapshot if necessary
			err = uploadSnapshot(snap)
			utils.FailOnErrorWithMessage(err, "failed to share snapshot")
		}
	}
	return 0
}

func snapshotToQueryResult(snap *dashboardtypes.SteampipeSnapshot, name string) (*queryresult.Result, error) {
	// find chart  nde - we expect only 1
	parsedName, err := modconfig.ParseResourceName(name)
	if err != nil {
		return nil, err
	}
	tableName := modconfig.BuildFullResourceName(parsedName.Mod, modconfig.BlockTypeTable, parsedName.Name)
	tablePanel, ok := snap.Panels[tableName]
	if !ok {
		return nil, fmt.Errorf("dashboard does not contain table result for query")
	}
	chartRun := tablePanel.(*dashboardexecute.LeafRun)
	if !ok {
		return nil, fmt.Errorf("failed to read query result from snapshot")
	}

	colTypes := make([]*sql.ColumnType, len(chartRun.Data.Columns))
	for i, c := range chartRun.Data.Columns {
		colTypes[i] = c.SqlColumnType
	}
	res := queryresult.NewQueryResult(colTypes)

	// TODO for now we do not support timing for snapshot query output - this need implementation
	// close timing channel to avoid lockup
	close(res.TimingResult)
	// start a goroutine to stream the results as rows
	go func() {
		for _, d := range chartRun.Data.Rows {
			res.StreamRow(maps.Values(d))
		}
		res.Close()
	}()

	return res, nil
}

func ensureQueryResource(name string, query string, queryIdx int, w *workspace.Workspace) string {
	var found bool
	var resource modconfig.HclResource
	if parsedName, err := modconfig.ParseResourceName(name); err == nil {
		resource, found = modconfig.GetResource(w, parsedName)
	}
	if found {
		return resource.Name()
	}
	// so this must be an ad hoc query - create a query resource and add to mod
	shortName := fmt.Sprintf("command_line_query_%d", queryIdx)
	title := fmt.Sprintf("Command line query %d", queryIdx)
	q := modconfig.NewQuery(&hcl.Block{}, w.Mod, shortName)
	q.Title = utils.ToStringPointer(title)
	q.SQL = utils.ToStringPointer(query)
	// add empty metadata
	q.SetMetadata(&modconfig.ResourceMetadata{})

	// add this to the workspace mod so the dashboard execution code can find it
	w.Mod.AddResource(q)
	// return the new resource name
	return q.Name()
}

func snapshotRequired() bool {
	return viper.IsSet(constants.ArgShare) ||
		viper.IsSet(constants.ArgSnapshot) ||
		viper.GetString(constants.ArgOutput) == constants.OutputFormatSnapshot

}

// getPipedStdinData reads the Standard Input and returns the available data as a string
// if and only if the data was piped to the process
func getPipedStdinData() string {
	fi, err := os.Stdin.Stat()
	if err != nil {
		error_helpers.ShowWarning("could not fetch information about STDIN")
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
