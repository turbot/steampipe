package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/control"
	"github.com/turbot/steampipe/pkg/control/controldisplay"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/workspace"
)

func checkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "check [flags] [mod/benchmark/control/\"all\"]",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runCheckCmd,
		Short:            "Execute one or more controls",
		Long: `Execute one or more Steampipe benchmarks and controls.

You may specify one or more benchmarks or controls to run (separated by a space), or run 'steampipe check all' to run all controls in the workspace.`,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			ctx := cmd.Context()
			workspaceResources, err := workspace.LoadResourceNames(ctx, viper.GetString(constants.ArgModLocation))
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveError
			}

			completions := []string{}

			for _, item := range workspaceResources.GetSortedBenchmarksAndControlNames() {
				if strings.HasPrefix(item, toComplete) {
					completions = append(completions, item)
				}
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmdconfig.
		OnCmd(cmd).
		AddCloudFlags().
		AddWorkspaceDatabaseFlag().
		AddModLocationFlag().
		AddBoolFlag(constants.ArgHeader, true, "Include column headers for csv and table output").
		AddBoolFlag(constants.ArgHelp, false, "Help for check", cmdconfig.FlagOptions.WithShortHand("h")).
		AddStringFlag(constants.ArgSeparator, ",", "Separator string for csv output").
		AddStringFlag(constants.ArgOutput, constants.OutputFormatText, "Output format: brief, csv, html, json, md, text, snapshot or none").
		AddBoolFlag(constants.ArgTiming, false, "Turn on the timer which reports check time").
		AddStringSliceFlag(constants.ArgSearchPath, nil, "Set a custom search_path for the steampipe user for a check session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, nil, "Set a prefix to the current search path for a check session (comma-separated)").
		AddStringFlag(constants.ArgTheme, "dark", "Set the output theme for 'text' output: light, dark or plain").
		AddStringSliceFlag(constants.ArgExport, nil, "Export output to file, supported formats: csv, html, json, md, nunit3, sps (snapshot), asff").
		AddBoolFlag(constants.ArgProgress, true, "Display control execution progress").
		AddBoolFlag(constants.ArgDryRun, false, "Show which controls will be run without running them").
		AddStringSliceFlag(constants.ArgTag, nil, "Filter controls based on their tag values ('--tag key=value')").
		AddStringSliceFlag(constants.ArgVarFile, nil, "Specify an .spvar file containing variable values").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, nil, "Specify the value of a variable").
		AddStringFlag(constants.ArgWhere, "", "SQL 'where' clause, or named query, used to filter controls (cannot be used with '--tag')").
		AddIntFlag(constants.ArgDatabaseQueryTimeout, constants.DatabaseDefaultCheckQueryTimeout, "The query timeout").
		AddIntFlag(constants.ArgMaxParallel, constants.DefaultMaxConnections, "The maximum number of concurrent database connections to open").
		AddBoolFlag(constants.ArgModInstall, true, "Specify whether to install mod dependencies before running the check").
		AddBoolFlag(constants.ArgInput, true, "Enable interactive prompts").
		AddBoolFlag(constants.ArgSnapshot, false, "Create snapshot in Turbot Pipes with the default (workspace) visibility").
		AddBoolFlag(constants.ArgShare, false, "Create snapshot in Turbot Pipes with 'anyone_with_link' visibility").
		AddStringArrayFlag(constants.ArgSnapshotTag, nil, "Specify tags to set on the snapshot").
		AddStringFlag(constants.ArgSnapshotLocation, "", "The location to write snapshots - either a local file path or a Turbot Pipes workspace").
		AddStringFlag(constants.ArgSnapshotTitle, "", "The title to give a snapshot")

	cmd.AddCommand(getListSubCmd(listSubCmdOptions{parentCmd: cmd}))
	return cmd
}

// exitCode=0 no runtime errors, no control alarms or errors
// exitCode=1 no runtime errors, 1 or more control alarms, no control errors
// exitCode=2 no runtime errors, 1 or more control errors
// exitCode=3+ runtime errors

func runCheckCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runCheckCmd start")

	// setup a cancel context and start cancel handler
	ctx, cancel := context.WithCancel(cmd.Context())
	contexthelpers.StartCancelHandler(cancel)

	defer func() {
		utils.LogTime("runCheckCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	// verify we have an argument
	if !validateCheckArgs(ctx, cmd, args) {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return
	}
	// if diagnostic mode is set, print out config and return
	if _, ok := os.LookupEnv(constants.EnvConfigDump); ok {
		cmdconfig.DisplayConfig()
		return
	}

	// verify that no other benchmarks/controls are given with an all
	if helpers.StringSliceContains(args, "all") && len(args) > 1 {
		error_helpers.FailOnError(sperr.New("cannot execute 'all' with other benchmarks/controls"))
	}

	// show the status spinner
	statushooks.Show(ctx)

	// initialise
	statushooks.SetStatus(ctx, "Initializing...")
	// disable status hooks in init - otherwise we will end up
	// getting status updates all the way down from the service layer
	initData := control.NewInitData(ctx)
	if initData.Result.Error != nil {
		exitCode = constants.ExitCodeInitializationFailed
		error_helpers.ShowError(ctx, initData.Result.Error)
		return
	}
	defer initData.Cleanup(ctx)

	// hide the spinner so that warning messages can be shown
	statushooks.Done(ctx)

	// if there is a usage warning we display it
	initData.Result.DisplayMessages()

	// pull out useful properties
	totalAlarms, totalErrors := 0, 0

	// get the execution trees
	// depending on the set of arguments and the export targets, we may get more than one
	// example :
	// "check benchmark.b1 benchmark.b2 --export check.json" would give one merged tree
	// "check benchmark.b1 benchmark.b2 --export json" would give multiple trees
	trees, err := getExecutionTrees(ctx, initData, args...)
	error_helpers.FailOnError(err)

	// execute controls synchronously (execute returns the number of alarms and errors)
	for _, namedTree := range trees {
		err = executeTree(ctx, namedTree.tree, initData)
		if err != nil {
			error_helpers.ShowError(ctx, err)
			continue
		}

		// append the total number of alarms and errors for multiple runs
		totalAlarms += namedTree.tree.Root.Summary.Status.Alarm
		totalErrors += namedTree.tree.Root.Summary.Status.Error

		err = publishSnapshot(ctx, namedTree.tree, viper.GetBool(constants.ArgShare), viper.GetBool(constants.ArgSnapshot))
		if err != nil {
			error_helpers.ShowError(ctx, err)
			continue
		}

		printTiming(namedTree.tree)

		err = exportExecutionTree(ctx, namedTree, initData, viper.GetStringSlice(constants.ArgExport))
		if err != nil {
			error_helpers.ShowError(ctx, err)
			continue
		}
	}

	// set the defined exit code after successful execution
	exitCode = getExitCode(totalAlarms, totalErrors)
}

// exportExecutionTree relies on the fact that the given tree is already executed
func exportExecutionTree(ctx context.Context, namedTree *namedExecutionTree, initData *control.InitData, exportArgs []string) error {
	statushooks.Show(ctx)
	defer statushooks.Done(ctx)

	if error_helpers.IsContextCanceled(ctx) {
		return ctx.Err()
	}

	exportMsg, err := initData.ExportManager.DoExport(ctx, namedTree.name, namedTree.tree, exportArgs)
	if err != nil {
		return err
	}

	// print the location where the file is exported if progress=true
	if len(exportMsg) > 0 && viper.GetBool(constants.ArgProgress) {
		fmt.Printf("\n")
		fmt.Println(strings.Join(exportMsg, "\n"))
		fmt.Printf("\n")
	}

	return nil
}

// executeTree executes and displays the (table) results of an execution
func executeTree(ctx context.Context, tree *controlexecute.ExecutionTree, initData *control.InitData) error {
	// create a context with check status hooks
	checkCtx := createCheckContext(ctx)
	err := tree.Execute(checkCtx)
	if err != nil {
		return err
	}

	err = displayControlResults(checkCtx, tree, initData.OutputFormatter)
	if err != nil {
		return err
	}
	return nil
}

func publishSnapshot(ctx context.Context, executionTree *controlexecute.ExecutionTree, shouldShare bool, shouldUpload bool) error {
	if error_helpers.IsContextCanceled(ctx) {
		return ctx.Err()
	}
	// if the share args are set, create a snapshot and share it
	if shouldShare || shouldUpload {
		statushooks.SetStatus(ctx, "Publishing snapshot")
		return controldisplay.PublishSnapshot(ctx, executionTree, shouldShare)
	}
	return nil
}

// getExecutionTrees returns a list of execution trees with the names of their export targets
// if the --export flag has the name of a file, a single merged tree is generated from the positional arguments
// otherwise, one tree is generated for each argument
//
// this is necessary, since exporters can only export entire execution trees and when a file name is provided, we want to export the whole tree into one file
//
// example :
// "check benchmark.b1 benchmark.b2 --export check.json" would give one merged tree
// "check benchmark.b1 benchmark.b2 --export json" would give multiple trees
func getExecutionTrees(ctx context.Context, initData *control.InitData, args ...string) ([]*namedExecutionTree, error) {
	var trees []*namedExecutionTree

	if initData.ExportManager.HasNamedExport(viper.GetStringSlice(constants.ArgExport)) {
		// create a single merged execution tree from all arguments
		executionTree, err := controlexecute.NewExecutionTree(ctx, initData.Workspace, initData.Client, initData.ControlFilterWhereClause, args...)
		if err != nil {
			return nil, sperr.WrapWithMessage(err, "could not create merged execution tree")
		}
		name := fmt.Sprintf("check.%s", initData.Workspace.Mod.ShortName)
		trees = append(trees, newNamedExecutionTree(name, executionTree))
	} else {
		for _, arg := range args {
			if error_helpers.IsContextCanceled(ctx) {
				return nil, ctx.Err()
			}
			executionTree, err := controlexecute.NewExecutionTree(ctx, initData.Workspace, initData.Client, initData.ControlFilterWhereClause, arg)
			if err != nil {
				return nil, sperr.WrapWithMessage(err, "could not create execution tree for %s", arg)
			}
			name, err := getExportName(arg, initData.Workspace.Mod.ShortName)
			if err != nil {
				return nil, sperr.WrapWithMessage(err, "could not evaluate export name for %s", arg)
			}
			trees = append(trees, newNamedExecutionTree(name, executionTree))
		}
	}
	return trees, ctx.Err()
}

// getExportName resolves the base name of the target file
func getExportName(targetName string, modShortName string) (string, error) {
	parsedName, _ := modconfig.ParseResourceName(targetName)
	if targetName == "all" {
		// there will be no block type = manually construct name
		return fmt.Sprintf("%s.%s", modShortName, parsedName.Name), nil
	}
	// default to just converting to valid resource name
	return parsedName.ToFullNameWithMod(modShortName)
}

// get the exit code for successful check run
func getExitCode(alarms int, errors int) int {
	// 1 or more control errors, return exitCode=2
	if errors > 0 {
		return constants.ExitCodeControlsError
	}
	// 1 or more controls in alarm, return exitCode=1
	if alarms > 0 {
		return constants.ExitCodeControlsAlarm
	}
	// no controls in alarm/error
	return constants.ExitCodeSuccessful
}

// create the context for the check run - add a control status renderer
func createCheckContext(ctx context.Context) context.Context {
	return controlstatus.AddControlHooksToContext(ctx, controlstatus.NewStatusControlHooks())
}

func validateCheckArgs(ctx context.Context, cmd *cobra.Command, args []string) bool {
	if len(args) == 0 {
		fmt.Println()
		error_helpers.ShowError(ctx, fmt.Errorf("you must provide at least one argument"))
		fmt.Println()
		//nolint:errcheck // cmd.Help always returns a nil error
		cmd.Help()
		fmt.Println()
		return false
	}

	if err := cmdconfig.ValidateSnapshotArgs(ctx); err != nil {
		error_helpers.ShowError(ctx, err)
		return false
	}

	// only 1 character is allowed for '--separator'
	if len(viper.GetString(constants.ArgSeparator)) > 1 {
		error_helpers.ShowError(ctx, fmt.Errorf("'--%s' can be 1 character long at most", constants.ArgSeparator))
		return false
	}

	// only 1 of 'share' and 'snapshot' may be set
	if viper.GetBool(constants.ArgShare) && viper.GetBool(constants.ArgSnapshot) {
		error_helpers.ShowError(ctx, fmt.Errorf("only 1 of '--%s' and '--%s' may be set", constants.ArgShare, constants.ArgSnapshot))
		return false
	}

	// if both '--where' and '--tag' have been used, then it's an error
	if viper.IsSet(constants.ArgWhere) && viper.IsSet(constants.ArgTag) {
		error_helpers.ShowError(ctx, fmt.Errorf("only 1 of '--%s' and '--%s' may be set", constants.ArgWhere, constants.ArgTag))
		return false
	}

	return true
}

func printTiming(tree *controlexecute.ExecutionTree) {
	if !shouldPrintTiming() {
		return
	}
	headers := []string{"", "Duration"}
	var rows [][]string

	for _, rg := range tree.Root.Groups {
		if rg.GroupItem.GetUnqualifiedName() == "benchmark.root" {
			// this is the created root benchmark
			// adds the children
			for _, g := range rg.Groups {
				rows = append(rows, []string{g.GroupItem.GetUnqualifiedName(), rg.Duration.String()})
			}
			continue
		}
		rows = append(rows, []string{rg.GroupItem.GetUnqualifiedName(), rg.Duration.String()})
	}
	for _, c := range tree.Root.ControlRuns {
		rows = append(rows, []string{c.Control.GetUnqualifiedName(), c.Duration.String()})
	}
	// blank line after renderer output
	fmt.Println()
	fmt.Println("Timing:")
	display.ShowWrappedTable(headers, rows, &display.ShowWrappedTableOptions{AutoMerge: false})
}

func shouldPrintTiming() bool {
	outputFormat := viper.GetString(constants.ArgOutput)

	return (viper.GetBool(constants.ArgTiming) && !viper.GetBool(constants.ArgDryRun)) &&
		(outputFormat == constants.OutputFormatText || outputFormat == constants.OutputFormatBrief)
}

func displayControlResults(ctx context.Context, executionTree *controlexecute.ExecutionTree, formatter controldisplay.Formatter) error {
	reader, err := formatter.Format(ctx, executionTree)
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, reader)
	return err
}

type namedExecutionTree struct {
	tree *controlexecute.ExecutionTree
	name string
}

func newNamedExecutionTree(name string, tree *controlexecute.ExecutionTree) *namedExecutionTree {
	return &namedExecutionTree{
		tree: tree,
		name: name,
	}
}
