package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
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
			w, err := workspace.LoadResourceNames(viper.GetString(constants.ArgModLocation))
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
		AddBoolFlag(constants.ArgHelp, false, "Help for query", cmdconfig.FlagOptions.WithShortHand("h")).
		AddBoolFlag(constants.ArgHeader, true, "Include column headers csv and table output").
		AddStringFlag(constants.ArgSeparator, ",", "Separator string for csv output").
		AddStringFlag(constants.ArgOutput, "table", "Output format: line, csv, json, table or snapshot").
		AddBoolFlag(constants.ArgTiming, false, "Turn on the timer which reports query time").
		AddBoolFlag(constants.ArgWatch, true, "Watch SQL files in the current workspace (works only in interactive mode)").
		AddStringSliceFlag(constants.ArgSearchPath, nil, "Set a custom search_path for the steampipe user for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, nil, "Set a prefix to the current search path for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgVarFile, nil, "Specify a file containing variable values").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, nil, "Specify the value of a variable").
		AddBoolFlag(constants.ArgInput, true, "Enable interactive prompts").
		AddBoolFlag(constants.ArgSnapshot, false, "Create snapshot in Steampipe Cloud with the default (workspace) visibility").
		AddBoolFlag(constants.ArgShare, false, "Create snapshot in Steampipe Cloud with 'anyone_with_link' visibility").
		AddStringArrayFlag(constants.ArgSnapshotTag, nil, "Specify tags to set on the snapshot").
		AddStringFlag(constants.ArgSnapshotTitle, "", "The title to give a snapshot").
		AddIntFlag(constants.ArgDatabaseQueryTimeout, 0, "The query timeout").
		AddStringSliceFlag(constants.ArgExport, nil, "Export output to file, supported format: sps (snapshot)").
		AddStringFlag(constants.ArgSnapshotLocation, "", "The location to write snapshots - either a local file path or a Steampipe Cloud workspace").
		AddBoolFlag(constants.ArgProgress, true, "Display snapshot upload status")

	cmd.AddCommand(getListSubCmd(listSubCmdOptions{parentCmd: cmd}))

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
	err := validateQueryArgs(ctx, args)
	error_helpers.FailOnError(err)

	// if diagnostic mode is set, print out config and return
	if _, ok := os.LookupEnv(constants.EnvDiagnostics); ok {
		cmdconfig.DisplayConfig()
		return
	}

	// enable spinner only in interactive mode
	interactiveMode := len(args) == 0
	// set config to indicate whether we are running an interactive query
	viper.Set(constants.ConfigKeyInteractive, interactiveMode)

	// start the initializer
	initData := query.NewInitData(ctx, args)
	error_helpers.FailOnError(initData.Result.Error)
	defer initData.Cleanup(ctx)

	switch {
	case interactiveMode:
		queryexecute.RunInteractiveSession(ctx, initData)
	case snapshotRequired():
		// if we are either outputting snapshot format, or sharing the results as a snapshot, execute the query
		// as a dashboard
		exitCode = executeSnapshotQuery(initData, ctx)
	default:
		// NOTE: disable any status updates - we do not want 'loading' output from any queries
		ctx = statushooks.DisableStatusHooks(ctx)

		// fall through to running a batch query
		// set global exit code
		exitCode = queryexecute.RunBatchSession(ctx, initData)
	}
}

func validateQueryArgs(ctx context.Context, args []string) error {
	interactiveMode := len(args) == 0
	if interactiveMode && (viper.IsSet(constants.ArgSnapshot) || viper.IsSet(constants.ArgShare)) {
		return fmt.Errorf("cannot share snapshots in interactive mode")
	}
	if interactiveMode && len(viper.GetStringSlice(constants.ArgExport)) > 0 {
		return fmt.Errorf("cannot export query results in interactive mode")
	}
	// if share or snapshot args are set, there must be a query specified
	err := cmdconfig.ValidateSnapshotArgs(ctx)
	if err != nil {
		return err
	}

	validOutputFormats := []string{constants.OutputFormatLine, constants.OutputFormatCSV, constants.OutputFormatTable, constants.OutputFormatJSON, constants.OutputFormatSnapshot, constants.OutputFormatSnapshotShort, constants.OutputFormatNone}
	output := viper.GetString(constants.ArgOutput)
	if !helpers.StringSliceContains(validOutputFormats, output) {
		return fmt.Errorf("invalid output format: '%s', must be one of [%s]", output, strings.Join(validOutputFormats, ", "))
	}

	return nil
}

func executeSnapshotQuery(initData *query.InitData, ctx context.Context) int {
	// start cancel handler to intercept interrupts and cancel the context
	// NOTE: use the initData Cancel function to ensure any initialisation is cancelled if needed
	contexthelpers.StartCancelHandler(initData.Cancel)

	// wait for init
	<-initData.Loaded
	if err := initData.Result.Error; err != nil {
		error_helpers.FailOnError(err)
	}

	// build ordered list of queries
	// (ordered for testing repeatability)
	var queryNames = utils.SortedMapKeys(initData.Queries)

	if len(queryNames) > 0 {
		for _, name := range queryNames {
			resolvedQuery := initData.Queries[name]
			// if a manual query is being run (i.e. not a named query), convert into a query and add to workspace
			// this is to allow us to use existing dashboard execution code
			queryProvider, existingResource := ensureQueryResource(name, resolvedQuery, initData.Workspace)

			// we need to pass the embedded initData to  GenerateSnapshot
			baseInitData := &initData.InitData

			// so a dashboard name was specified - just call GenerateSnapshot
			snap, err := dashboardexecute.GenerateSnapshot(ctx, queryProvider.Name(), baseInitData, nil)
			error_helpers.FailOnError(err)

			// set the filename root for the snapshot (in case needed)
			if !existingResource {
				snap.FileNameRoot = "query"
			}

			// display the result
			switch viper.GetString(constants.ArgOutput) {
			case constants.OutputFormatNone:
				// do nothing
			case constants.OutputFormatSnapshot, constants.OutputFormatSnapshotShort:
				// if the format is snapshot, just dump it out
				jsonOutput, err := json.MarshalIndent(snap, "", "  ")
				if err != nil {
					error_helpers.FailOnErrorWithMessage(err, "failed to display result as snapshot")
				}
				fmt.Println(string(jsonOutput))
			default:
				// otherwise convert the snapshot into a query result
				result, err := snapshotToQueryResult(snap, queryProvider.Name())
				error_helpers.FailOnErrorWithMessage(err, "failed to display result as snapshot")
				display.ShowOutput(ctx, result, display.DisableTiming())
			}

			// share the snapshot if necessary
			err = publishSnapshotIfNeeded(ctx, snap)
			error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("failed to publish snapshot to %s", viper.GetString(constants.ArgSnapshotLocation)))

			// export the result if necessary
			exportArgs := viper.GetStringSlice(constants.ArgExport)
			err = initData.ExportManager.DoExport(ctx, snap.FileNameRoot, snap, exportArgs)
			error_helpers.FailOnErrorWithMessage(err, "failed to export snapshot")
		}
	}
	return 0
}

func snapshotToQueryResult(snap *dashboardtypes.SteampipeSnapshot, name string) (*queryresult.Result, error) {
	// find chart node - we expect only 1
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
	// check for error
	if err := chartRun.GetError(); err != nil {
		return nil, error_helpers.DecodePgError(err)
	}

	res := queryresult.NewResult(chartRun.Data.Columns)

	// start a goroutine to stream the results as rows
	go func() {
		rowVals := make([]interface{}, len(chartRun.Data.Columns))
		for _, d := range chartRun.Data.Rows {
			for i, c := range chartRun.Data.Columns {
				rowVals[i] = d[c.Name]
			}
			res.StreamRow(rowVals)
		}
		res.TimingResult <- chartRun.TimingResult
		res.Close()
	}()

	return res, nil
}

// convert the given command line query into a query resource and add to workspace
// this is to allow us to use existing dashboard execution code
func ensureQueryResource(name string, resolvedQuery *modconfig.ResolvedQuery, w *workspace.Workspace) (queryProvider modconfig.HclResource, existingResource bool) {
	// is this an existing resource?
	if parsedName, err := modconfig.ParseResourceName(name); err == nil {
		if resource, found := modconfig.GetResource(w, parsedName); found {
			return resource, true
		}
	}

	// build name
	shortName := "command_line_query"

	// this is NOT a named query - create the query using RawSql
	q := modconfig.NewQuery(&hcl.Block{}, w.Mod, shortName).(*modconfig.Query)
	q.SQL = utils.ToStringPointer(resolvedQuery.RawSQL)
	q.SetArgs(resolvedQuery.QueryArgs())
	// add empty metadata
	q.SetMetadata(&modconfig.ResourceMetadata{})

	// add to the workspace mod so the dashboard execution code can find it
	w.Mod.AddResource(q)
	// return the new resource name
	return q, false
}

func snapshotRequired() bool {
	SnapshotFormatNames := []string{constants.OutputFormatSnapshot, constants.OutputFormatSnapshotShort}
	// if a snapshot exporter is specified return true
	for _, e := range viper.GetStringSlice(constants.ArgExport) {
		if helpers.StringSliceContains(SnapshotFormatNames, e) || path.Ext(e) == constants.SnapshotExtension {
			return true
		}
	}
	// if share/snapshot args are set or output is snapshot, return true
	return viper.IsSet(constants.ArgShare) ||
		viper.IsSet(constants.ArgSnapshot) ||
		helpers.StringSliceContains(SnapshotFormatNames, viper.GetString(constants.ArgOutput))

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
