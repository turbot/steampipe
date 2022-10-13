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
	"github.com/turbot/steampipe/pkg/control/controldisplay"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/statushooks"
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
			workspaceResources, err := workspace.LoadResourceNames(viper.GetString(constants.ArgWorkspaceChDir))
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
		AddBoolFlag(constants.ArgHeader, "", true, "Include column headers for csv and table output").
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for check").
		AddStringFlag(constants.ArgSeparator, "", ",", "Separator string for csv output").
		AddStringFlag(constants.ArgOutput, "", constants.OutputFormatText, "Select a console output format: brief, csv, html, json, md, text or none").
		AddBoolFlag(constants.ArgTiming, "", false, "Turn on the timer which reports check time").
		AddStringSliceFlag(constants.ArgSearchPath, "", nil, "Set a custom search_path for the steampipe user for a check session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", nil, "Set a prefix to the current search path for a check session (comma-separated)").
		AddStringFlag(constants.ArgTheme, "", "dark", "Set the output theme for 'text' output: light, dark or plain").
		AddStringSliceFlag(constants.ArgExport, "", nil, "Export output to files in various output formats: csv, html, json, md, nunit3, snapshot or asff (json)").
		AddBoolFlag(constants.ArgProgress, "", true, "Display control execution progress").
		AddBoolFlag(constants.ArgDryRun, "", false, "Show which controls will be run without running them").
		AddStringSliceFlag(constants.ArgTag, "", nil, "Filter controls based on their tag values ('--tag key=value')").
		AddStringSliceFlag(constants.ArgVarFile, "", nil, "Specify an .spvar file containing variable values").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, "", nil, "Specify the value of a variable").
		AddStringFlag(constants.ArgWhere, "", "", "SQL 'where' clause, or named query, used to filter controls (cannot be used with '--tag')").
		AddIntFlag(constants.ArgMaxParallel, "", constants.DefaultMaxConnections, "The maximum number of parallel executions", cmdconfig.FlagOptions.Hidden()).
		AddBoolFlag(constants.ArgModInstall, "", true, "Specify whether to install mod dependencies before running the check").
		AddBoolFlag(constants.ArgInput, "", true, "Enable interactive prompts").
		AddStringFlag(constants.ArgSnapshot, "", "", "Create snapshot in Steampipe Cloud with the default (workspace) visibility.", cmdconfig.FlagOptions.NoOptDefVal(constants.ArgShareNoOptDefault)).
		AddStringFlag(constants.ArgShare, "", "", "Create snapshot in Steampipe Cloud with 'anyone_with_link' visibility.", cmdconfig.FlagOptions.NoOptDefVal(constants.ArgShareNoOptDefault)).
		AddStringArrayFlag(constants.ArgSnapshotTag, "", nil, "Specify the value of a tag to set on the snapshot").
		AddStringFlag(constants.ArgWorkspace, "", "", "The cloud workspace... ")

	return cmd
}

// exitCode=1 For unknown errors resulting in panics
// exitCode=2 For insufficient args

func runCheckCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runCheckCmd start")
	initData := &initialisation.InitData{}

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

		initData.Cleanup(ctx)
	}()

	// verify we have an argument
	if !validateCheckArgs(ctx, cmd, args) {
		exitCode = constants.ExitCodeInsufficientOrWrongArguments
		return
	}

	// initialise
	initData = initialiseCheck(ctx)
	error_helpers.FailOnError(initData.Result.Error)
	// if there is a usage warning we display it
	initData.Result.DisplayMessages()

	// pull out useful properties
	workspace := initData.Workspace
	client := initData.Client
	failures := 0
	var durations []time.Duration

	shouldShare := viper.IsSet(constants.ArgShare)
	shouldUpload := viper.IsSet(constants.ArgSnapshot)
	generateSnapshot := shouldShare || shouldUpload
	if generateSnapshot {
		// if no output explicitly set, show nothing
		if !viper.IsSet(constants.ArgOutput) {
			viper.Set(constants.ArgOutput, constants.OutputFormatNone)
		}
	}

	// treat each arg as a separate execution
	for _, targetName := range args {
		if utils.IsContextCancelled(ctx) {
			durations = append(durations, 0)
			// skip over this arg, since the execution was cancelled
			// (do not just quit as we want to populate the durations)
			continue
		}

		// create the execution tree
		executionTree, err := controlexecute.NewExecutionTree(ctx, workspace, client, targetName)
		error_helpers.FailOnError(err)

		// execute controls synchronously (execute returns the number of failures)
		failures += executionTree.Execute(ctx)
		err = displayControlResults(ctx, executionTree)
		error_helpers.FailOnError(err)

		exportArgs := viper.GetStringSlice(constants.ArgExport)
		err = initData.ExportManager.DoExport(ctx, targetName, executionTree, exportArgs)
		error_helpers.FailOnError(err)

		// if the share args are set, create a snapshot and share it
		if generateSnapshot {
			controldisplay.ShareAsSnapshot(executionTree, shouldShare)
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

	if err := validateCloudArgs(); err != nil {
		error_helpers.ShowError(ctx, err)
		return false
	}
	// only 1 of 'share' and 'snapshot' may be set
	if len(viper.GetString(constants.ArgShare)) > 0 && len(viper.GetString(constants.ArgSnapshot)) > 0 {
		error_helpers.ShowError(ctx, fmt.Errorf("only 1 of 'share' and 'snapshot' may be set"))
		return false
	}

	return true
}

func initialiseCheck(ctx context.Context) *initialisation.InitData {
	statushooks.SetStatus(ctx, "Initializing...")
	defer statushooks.Done(ctx)

	// load the workspace
	w, err := interactive.LoadWorkspacePromptingForVariables(ctx)
	error_helpers.FailOnErrorWithMessage(err, "failed to load workspace")

	initData := initialisation.NewInitData(w).Init(ctx, constants.InvokerCheck)
	if initData.Result.Error != nil {
		return initData
	}

	if len(viper.GetStringSlice(constants.ArgExport)) > 0 {
		registerCheckExporters(initData)
	}

	// control specific init
	if !w.ModfileExists() {
		initData.Result.Error = workspace.ErrorNoModDefinition
	}

	if viper.GetString(constants.ArgOutput) == constants.OutputFormatNone {
		// set progress to false
		viper.Set(constants.ArgProgress, false)
	}
	// set color schema
	err = initialiseCheckColorScheme()
	if err != nil {
		initData.Result.Error = err
		return initData
	}

	if len(initData.Workspace.GetResourceMaps().Controls) == 0 {
		initData.Result.AddWarnings("no controls found in current workspace")
	}

	if err := controldisplay.EnsureTemplates(); err != nil {
		initData.Result.Error = err
		return initData
	}

	return initData
}

// register exporters for each of the supported check formats
func registerCheckExporters(initData *initialisation.InitData) {
	exporters, err := controldisplay.GetExporters()
	error_helpers.FailOnErrorWithMessage(err, "failed to load exporters")

	// register all exporters
	initData.RegisterExporters(exporters...)
}

func initialiseCheckColorScheme() error {
	theme := viper.GetString(constants.ArgTheme)
	if !viper.GetBool(constants.ConfigKeyIsTerminalTTY) {
		// enforce plain output for non-terminals
		theme = "plain"
	}
	themeDef, ok := controldisplay.ColorSchemes[theme]
	if !ok {
		return fmt.Errorf("invalid theme '%s'", theme)
	}
	scheme, err := controldisplay.NewControlColorScheme(themeDef)
	if err != nil {
		return err
	}
	controldisplay.ControlColors = scheme
	return nil
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
	display.ShowWrappedTable(headers, rows, false)
}

func shouldPrintTiming() bool {
	outputFormat := viper.GetString(constants.ArgOutput)

	return (viper.GetBool(constants.ArgTiming) && !viper.GetBool(constants.ArgDryRun)) &&
		(outputFormat == constants.OutputFormatText || outputFormat == constants.OutputFormatBrief)
}

func displayControlResults(ctx context.Context, executionTree *controlexecute.ExecutionTree) error {
	output := viper.GetString(constants.ArgOutput)
	formatter, err := parseOutputArg(output)
	if err != nil {
		fmt.Println(err)
		return err
	}
	reader, err := formatter.Format(ctx, executionTree)
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, reader)
	return err
}

// parseOutputArg parses the --output flag value and returns the Formatter that can format the data
func parseOutputArg(arg string) (formatter controldisplay.Formatter, err error) {
	formatResolver, err := controldisplay.NewFormatResolver()
	if err != nil {
		return nil, err
	}

	return formatResolver.GetFormatter(arg)
}
