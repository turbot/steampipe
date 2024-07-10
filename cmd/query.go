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
	"github.com/thediveo/enumflag/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/cloud"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/connection_sync"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/query"
	"github.com/turbot/steampipe/pkg/query/queryexecute"
	"github.com/turbot/steampipe/pkg/snapshot2"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/workspace"
)

// variable used to assign the timing mode flag
var queryTimingMode = constants.QueryTimingModeOff

// variable used to assign the output mode flag
var queryOutputMode = constants.QueryOutputModeTable

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
			ctx := cmd.Context()
			w, err := workspace.LoadResourceNames(ctx, viper.GetString(constants.ArgModLocation))
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
		AddCloudFlags().
		AddWorkspaceDatabaseFlag().
		AddModLocationFlag().
		AddBoolFlag(constants.ArgHelp, false, "Help for query", cmdconfig.FlagOptions.WithShortHand("h")).
		AddBoolFlag(constants.ArgHeader, true, "Include column headers csv and table output").
		AddStringFlag(constants.ArgSeparator, ",", "Separator string for csv output").
		AddVarFlag(enumflag.New(&queryOutputMode, constants.ArgOutput, constants.QueryOutputModeIds, enumflag.EnumCaseInsensitive),
			constants.ArgOutput,
			fmt.Sprintf("Output format; one of: %s", strings.Join(constants.FlagValues(constants.QueryOutputModeIds), ", "))).
		AddVarFlag(enumflag.New(&queryTimingMode, constants.ArgTiming, constants.QueryTimingModeIds, enumflag.EnumCaseInsensitive),
			constants.ArgTiming,
			fmt.Sprintf("Display query timing; one of: %s", strings.Join(constants.FlagValues(constants.QueryTimingModeIds), ", ")),
			cmdconfig.FlagOptions.NoOptDefVal(constants.ArgOn)).
		AddBoolFlag(constants.ArgWatch, true, "Watch SQL files in the current workspace (works only in interactive mode)").
		AddStringSliceFlag(constants.ArgSearchPath, nil, "Set a custom search_path for the steampipe user for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, nil, "Set a prefix to the current search path for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgVarFile, nil, "Specify a file containing variable values").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, nil, "Specify the value of a variable").
		AddBoolFlag(constants.ArgInput, true, "Enable interactive prompts").
		AddBoolFlag(constants.ArgSnapshot, false, "Create snapshot in Turbot Pipes with the default (workspace) visibility").
		AddBoolFlag(constants.ArgShare, false, "Create snapshot in Turbot Pipes with 'anyone_with_link' visibility").
		AddStringArrayFlag(constants.ArgSnapshotTag, nil, "Specify tags to set on the snapshot").
		AddStringFlag(constants.ArgSnapshotTitle, "", "The title to give a snapshot").
		AddIntFlag(constants.ArgDatabaseQueryTimeout, 0, "The query timeout").
		AddStringSliceFlag(constants.ArgExport, nil, "Export output to file, supported format: sps (snapshot)").
		AddStringFlag(constants.ArgSnapshotLocation, "", "The location to write snapshots - either a local file path or a Turbot Pipes workspace").
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

	// validate args
	err := validateQueryArgs(ctx, args)
	error_helpers.FailOnError(err)

	// if diagnostic mode is set, print out config and return
	if _, ok := os.LookupEnv(constants.EnvConfigDump); ok {
		cmdconfig.DisplayConfig()
		return
	}

	if len(args) == 0 {
		// no positional arguments - check if there's anything on stdin
		if stdinData := getPipedStdinData(); len(stdinData) > 0 {
			// we have data - treat this as an argument
			args = append(args, stdinData)
		}
	}

	// enable paging only in interactive mode
	interactiveMode := len(args) == 0
	// set config to indicate whether we are running an interactive query
	viper.Set(constants.ConfigKeyInteractive, interactiveMode)

	// initialize the cancel handler - for context cancellation
	initCtx, cancel := context.WithCancel(ctx)
	contexthelpers.StartCancelHandler(cancel)

	// start the initializer
	initData := query.NewInitData(initCtx, args)
	if initData.Result.Error != nil {
		exitCode = constants.ExitCodeInitializationFailed
		error_helpers.ShowError(ctx, initData.Result.Error)
		return
	}
	defer initData.Cleanup(ctx)

	var failures int
	switch {
	case interactiveMode:
		err = queryexecute.RunInteractiveSession(ctx, initData)
	//case snapshotRequired():
	//	// if we are either outputting snapshot format, or sharing the results as a snapshot, execute the query
	//	// as a dashboard
	//	failures = executeSnapshotQuery(initData, ctx)
	default:
		// NOTE: disable any status updates - we do not want 'loading' output from any queries
		ctx = statushooks.DisableStatusHooks(ctx)

		// fall through to running a batch query
		failures, err = queryexecute.RunBatchSession(ctx, initData)
	}

	// check for err and set the exit code else set the exit code if some queries failed or some rows returned an error
	if err != nil {
		exitCode = constants.ExitCodeInitializationFailed
		error_helpers.ShowError(ctx, err)
	} else if failures > 0 {
		exitCode = constants.ExitCodeQueryExecutionFailed
	}
}

func validateQueryArgs(ctx context.Context, args []string) error {
	interactiveMode := len(args) == 0
	if interactiveMode && (viper.IsSet(constants.ArgSnapshot) || viper.IsSet(constants.ArgShare)) {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return sperr.New("cannot share snapshots in interactive mode")
	}
	if interactiveMode && len(viper.GetStringSlice(constants.ArgExport)) > 0 {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return sperr.New("cannot export query results in interactive mode")
	}
	// if share or snapshot args are set, there must be a query specified
	err := cmdconfig.ValidateSnapshotArgs(ctx)
	if err != nil {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return err
	}

	validOutputFormats := []string{constants.OutputFormatLine, constants.OutputFormatCSV, constants.OutputFormatTable, constants.OutputFormatJSON, constants.OutputFormatSnapshot, constants.OutputFormatSnapshotShort, constants.OutputFormatNone}
	output := viper.GetString(constants.ArgOutput)
	if !helpers.StringSliceContains(validOutputFormats, output) {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return sperr.New("invalid output format: '%s', must be one of [%s]", output, strings.Join(validOutputFormats, ", "))
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
		exitCode = constants.ExitCodeInitializationFailed
		error_helpers.FailOnError(err)
	}

	// if there is a custom search path, wait until the first connection of each plugin has loaded
	if customSearchPath := initData.Client.GetCustomSearchPath(); customSearchPath != nil {
		if err := connection_sync.WaitForSearchPathSchemas(ctx, initData.Client, customSearchPath); err != nil {
			exitCode = constants.ExitCodeInitializationFailed
			error_helpers.FailOnError(err)
		}
	}

	for _, resolvedQuery := range initData.Queries {
		// if a manual query is being run (i.e. not a named query), convert into a query and add to workspace
		// this is to allow us to use existing dashboard execution code
		queryProvider, existingResource := ensureSnapshotQueryResource(resolvedQuery.Name, resolvedQuery, initData.Workspace)

		// we need to pass the embedded initData to  GenerateSnapshot
		baseInitData := &initData.InitData

		// so a dashboard name was specified - just call GenerateSnapshot
		snap, err := snapshot2.GenerateSnapshot(ctx, queryProvider.Name(), baseInitData, nil)
		if err != nil {
			exitCode = constants.ExitCodeSnapshotCreationFailed
			error_helpers.FailOnError(err)
		}

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
			// result, err := snapshotToQueryResult(snap)
			// error_helpers.FailOnErrorWithMessage(err, "failed to display result as snapshot")
			fmt.Println()
			// display.ShowOutput(ctx, result, display.WithTimingDisabled())
		}

		// share the snapshot if necessary
		// err = publishSnapshotIfNeeded(ctx, snap)
		// if err != nil {
		// 	exitCode = constants.ExitCodeSnapshotUploadFailed
		// 	error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("failed to publish snapshot to %s", viper.GetString(constants.ArgSnapshotLocation)))
		// }

		// export the result if necessary
		// exportArgs := viper.GetStringSlice(constants.ArgExport)
		// exportMsg, err := initData.ExportManager.DoExport(ctx, snap.FileNameRoot, snap, exportArgs)
		// if err != nil {
		// 	exitCode = constants.ExitCodeSnapshotCreationFailed
		// 	error_helpers.FailOnErrorWithMessage(err, "failed to export snapshot")
		// }
		// // print the location where the file is exported
		// if len(exportMsg) > 0 && viper.GetBool(constants.ArgProgress) {
		// 	fmt.Printf("\n")
		// 	fmt.Println(strings.Join(exportMsg, "\n"))
		// 	fmt.Printf("\n")
		// }
	}
	return 0
}

// func snapshotToQueryResult(snap *snapshot2.SteampipeSnapshot) (*queryresult.Result, error) {
// 	// the table of a snapshot query has a fixed name
// 	tablePanel, ok := snap.Panels[modconfig.SnapshotQueryTableName]
// 	if !ok {
// 		return nil, sperr.New("dashboard does not contain table result for query")
// 	}
// 	chartRun := tablePanel.(*snapshot.LeafRun)
// 	if !ok {
// 		return nil, sperr.New("failed to read query result from snapshot")
// 	}
// 	// check for error
// 	if err := chartRun.GetError(); err != nil {
// 		return nil, error_helpers.DecodePgError(err)
// 	}

// 	res := queryresult.NewResult(chartRun.Data.Columns)

// 	// start a goroutine to stream the results as rows
// 	go func() {
// 		for _, d := range chartRun.Data.Rows {
// 			// we need to allocate a new slice everytime, since this gets read
// 			// asynchronously on the other end and we need to make sure that we don't overwrite
// 			// data already sent
// 			rowVals := make([]interface{}, len(chartRun.Data.Columns))
// 			for i, c := range chartRun.Data.Columns {
// 				rowVals[i] = d[c.Name]
// 			}
// 			res.StreamRow(rowVals)
// 		}
// 		res.TimingResult <- chartRun.TimingResult
// 		res.Close()
// 	}()

// 	return res, nil
// }

// convert the given command line query into a query resource and add to workspace
// this is to allow us to use existing dashboard execution code
func ensureSnapshotQueryResource(name string, resolvedQuery *modconfig.ResolvedQuery, w *workspace.Workspace) (queryProvider modconfig.HclResource, existingResource bool) {
	// is this an existing resource?
	if parsedName, err := modconfig.ParseResourceName(name); err == nil {
		if resource, found := w.GetResource(parsedName); found {
			return resource, true
		}
	}

	// build name
	shortName := "command_line_query"

	// this is NOT a named query - create the query using RawSql
	q := modconfig.NewQuery(&hcl.Block{Type: modconfig.BlockTypeQuery}, w.Mod, shortName).(*modconfig.Query)
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

func publishSnapshotIfNeeded(ctx context.Context, snapshot *dashboardtypes.SteampipeSnapshot) error {
	shouldShare := viper.GetBool(constants.ArgShare)
	shouldUpload := viper.GetBool(constants.ArgSnapshot)

	if !(shouldShare || shouldUpload) {
		return nil
	}

	message, err := cloud.PublishSnapshot(ctx, snapshot, shouldShare)
	if err != nil {
		// reword "402 Payment Required" error
		return handlePublishSnapshotError(err)
	}
	if viper.GetBool(constants.ArgProgress) {
		fmt.Println(message)
	}
	return nil
}

func handlePublishSnapshotError(err error) error {
	if err.Error() == "402 Payment Required" {
		return fmt.Errorf("maximum number of snapshots reached")
	}
	return err
}
