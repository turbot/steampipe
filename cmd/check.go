package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/control"
	"github.com/turbot/steampipe/pkg/control/controldisplay"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
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
			workspaceResources, err := workspace.LoadResourceNames(viper.GetString(constants.ArgModLocation))
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
		AddBoolFlag(constants.ArgSnapshot, false, "Create snapshot in Steampipe Cloud with the default (workspace) visibility").
		AddBoolFlag(constants.ArgShare, false, "Create snapshot in Steampipe Cloud with 'anyone_with_link' visibility").
		AddStringArrayFlag(constants.ArgSnapshotTag, nil, "Specify tags to set on the snapshot").
		AddStringFlag(constants.ArgSnapshotLocation, "", "The location to write snapshots - either a local file path or a Steampipe Cloud workspace").
		AddStringFlag(constants.ArgSnapshotTitle, "", "The title to give a snapshot")

	cmd.AddCommand(getListSubCmd(listSubCmdOptions{parentCmd: cmd}))
	return cmd
}

// exitCode=1 For unknown errors resulting in panics
// exitCode=2 For insufficient args

func runCheckCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runCheckCmd start")

	// setup a cancel context and start cancel handler
	ctx, cancel := context.WithCancel(cmd.Context())
	contexthelpers.StartCancelHandler(cancel)
	// create a context with check status hooks
	ctx = createCheckContext(ctx)

	defer func() {
		utils.LogTime("runCheckCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	// verify we have an argument
	if !validateCheckArgs(ctx, cmd, args) {
		exitCode = constants.ExitCodeInsufficientOrWrongArguments
		return
	}
	// if diagnostic mode is set, print out config and return
	if _, ok := os.LookupEnv(constants.EnvDiagnostics); ok {
		cmdconfig.DisplayConfig()
		return
	}

	// initialise
	initData := control.NewInitData(ctx)
	error_helpers.FailOnError(initData.Result.Error)
	defer initData.Cleanup(ctx)

	// if there is a usage warning we display it
	initData.Result.DisplayMessages()

	// pull out useful properties
	w := initData.Workspace
	client := initData.Client
	failures := 0
	var durations []time.Duration

	shouldShare := viper.GetBool(constants.ArgShare)
	shouldUpload := viper.GetBool(constants.ArgSnapshot)
	generateSnapshot := shouldShare || shouldUpload

	// treat each arg as a separate execution
	for _, targetName := range args {
		if utils.IsContextCancelled(ctx) {
			durations = append(durations, 0)
			// skip over this arg, since the execution was cancelled
			// (do not just quit as we want to populate the durations)
			continue
		}

		// create the execution tree
		executionTree, err := controlexecute.NewExecutionTree(ctx, w, client, targetName, initData.ControlFilterWhereClause)
		error_helpers.FailOnError(err)

		// execute controls synchronously (execute returns the number of failures)
		failures += executionTree.Execute(ctx)
		err = displayControlResults(ctx, executionTree, initData.OutputFormatter)
		error_helpers.FailOnError(err)

		// add the mod name to the target name to get the export file root
		parsedName, _ := modconfig.ParseResourceName(targetName)
		exportName := parsedName.ToFullNameWithMod(w.Mod.ShortName)

		exportArgs := viper.GetStringSlice(constants.ArgExport)
		err = initData.ExportManager.DoExport(ctx, exportName, executionTree, exportArgs)
		error_helpers.FailOnError(err)

		// if the share args are set, create a snapshot and share it
		if generateSnapshot {
			err = controldisplay.PublishSnapshot(ctx, executionTree, shouldShare)
			error_helpers.FailOnError(err)
		}

		durations = append(durations, executionTree.EndTime.Sub(executionTree.StartTime))
	}

	if shouldPrintTiming() {
		printTiming(args, durations)
	}
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

func printTiming(args []string, durations []time.Duration) {
	headers := []string{"", "Duration"}
	var rows [][]string
	for idx, arg := range args {
		rows = append(rows, []string{arg, durations[idx].String()})
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
