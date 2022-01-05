package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/contexthelpers"
	"github.com/turbot/steampipe/control"
	"github.com/turbot/steampipe/control/controldisplay"
	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/db/db_client"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/modinstaller"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
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
		AddStringFlag(constants.ArgOutput, "", "text", "Select a console output format: brief, csv, html, json, md, text or none").
		AddBoolFlag(constants.ArgTimer, "", false, "Turn on the timer which reports check time").
		AddStringSliceFlag(constants.ArgSearchPath, "", nil, "Set a custom search_path for the steampipe user for a check session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", nil, "Set a prefix to the current search path for a check session (comma-separated)").
		AddStringFlag(constants.ArgTheme, "", "dark", "Set the output theme for 'text' output: light, dark or plain").
		AddStringSliceFlag(constants.ArgExport, "", nil, "Export output to files in various output formats: csv, html, json or md").
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
		AddBoolFlag(constants.ArgModInstall, "", true, "Specify whether to install mod depdencies before runnign the check")

	return cmd
}

// exitCode=1 For unknown errors resulting in panics
// exitCode=2 For insufficient args

func runCheckCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runCheckCmd start")
	initData := &control.InitData{}

	// setup a cancel context and start cancel handler
	ctx, cancel := context.WithCancel(cmd.Context())
	contexthelpers.StartCancelHandler(cancel)

	defer func() {
		utils.LogTime("runCheckCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
			exitCode = 1
		}

		if initData.Client != nil {
			log.Printf("[TRACE] close client")
			initData.Client.Close(ctx)
		}
		if initData.Workspace != nil {
			initData.Workspace.Close()
		}
	}()

	// verify we have an argument
	if !validateArgs(cmd, args) {
		return
	}

	// if progress is disabled, update context to contain a null status hooks object
	if !viper.GetBool(constants.ArgProgress) {
		statushooks.DisableStatusHooks(ctx)
	}

	// initialise
	initData = initialiseCheck(ctx)
	if shouldExit := handleCheckInitResult(ctx, initData); shouldExit {
		return
	}

	// pull out useful properties
	workspace := initData.Workspace
	client := initData.Client
	failures := 0
	var exportErrors []error
	exportErrorsLock := sync.Mutex{}
	exportWaitGroup := sync.WaitGroup{}
	var durations []time.Duration

	// treat each arg as a separate execution
	for _, arg := range args {
		if utils.IsContextCancelled(ctx) {
			durations = append(durations, 0)
			// skip over this arg, since the execution was cancelled
			// (do not just quit as we want to populate the durations)
			continue
		}

		// get the export formats for this argument
		exportFormats, err := getExportTargets(arg)
		utils.FailOnError(err)

		// create the execution tree
		executionTree, err := controlexecute.NewExecutionTree(ctx, workspace, client, arg)
		utils.FailOnErrorWithMessage(err, "failed to resolve controls from argument")

		// execute controls synchronously (execute returns the number of failures)
		failures += executionTree.Execute(ctx, client)
		err = displayControlResults(ctx, executionTree)
		utils.FailOnError(err)

		if len(exportFormats) > 0 {
			d := control.ExportData{ExecutionTree: executionTree, ExportFormats: exportFormats, ErrorsLock: &exportErrorsLock, Errors: exportErrors, WaitGroup: &exportWaitGroup}
			exportCheckResult(ctx, &d)
		}

		durations = append(durations, executionTree.EndTime.Sub(executionTree.StartTime))
	}

	// wait for exports to complete
	exportWaitGroup.Wait()

	if len(exportErrors) > 0 {
		utils.ShowError(utils.CombineErrors(exportErrors...))
	}

	if shouldPrintTiming() {
		printTiming(args, durations)
	}

	// set global exit code
	exitCode = failures
}

func validateArgs(cmd *cobra.Command, args []string) bool {
	if len(args) == 0 {
		fmt.Println()
		utils.ShowError(fmt.Errorf("you must provide at least one argument"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		exitCode = 2
		return false
	}
	return true
}

func initialiseCheck(ctx context.Context) *control.InitData {
	statushooks.SetStatus(ctx, "Initializing...")
	defer statushooks.Done(ctx)

	initData := &control.InitData{
		Result: &db_common.InitResult{},
	}
	if viper.GetBool(constants.ArgModInstall) {
		opts := &modinstaller.InstallOpts{WorkspacePath: viper.GetString(constants.ArgWorkspaceChDir)}
		_, err := modinstaller.InstallWorkspaceDependencies(opts)
		if err != nil {
			initData.Result.Error = err
			return initData
		}
	}
	err := validateOutputFormat()
	if err != nil {
		initData.Result.Error = err
		return initData
	}

	err = cmdconfig.ValidateConnectionStringArgs()
	if err != nil {
		initData.Result.Error = err
		return initData
	}

	// set color schema
	err = initialiseColorScheme()
	if err != nil {
		initData.Result.Error = err
		return initData
	}
	// load workspace
	initData.Workspace, err = loadWorkspacePromptingForVariables(ctx)
	if err != nil {
		if !utils.IsCancelledError(err) {
			err = utils.PrefixError(err, "failed to load workspace")
		}
		initData.Result.Error = err
		return initData
	}

	// check if the required plugins are installed
	err = initData.Workspace.CheckRequiredPluginsInstalled()
	if err != nil {
		initData.Result.Error = err
		return initData
	}

	if len(initData.Workspace.Controls) == 0 {
		initData.Result.AddWarnings("no controls found in current workspace")
	}

	statushooks.SetStatus(ctx, "Connecting to service...")
	// get a client
	var client db_common.Client
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		client, err = db_client.NewDbClient(ctx, connectionString)
	} else {
		// when starting the database, installers may trigger their own spinners
		client, err = db_local.GetLocalClient(ctx, constants.InvokerCheck)
	}

	if err != nil {
		initData.Result.Error = err
		return initData
	}
	initData.Client = client

	refreshResult := initData.Client.RefreshConnectionAndSearchPaths(ctx)
	if refreshResult.Error != nil {
		initData.Result.Error = refreshResult.Error
		return initData
	}
	initData.Result.AddWarnings(refreshResult.Warnings...)

	// setup the session data - prepared statements and introspection tables
	sessionDataSource := workspace.NewSessionDataSource(initData.Workspace, nil)

	// register EnsureSessionData as a callback on the client.
	// if the underlying SQL client has certain errors (for example context expiry) it will reset the session
	// so our client object calls this callback to restore the session data
	initData.Client.SetEnsureSessionDataFunc(func(localCtx context.Context, conn *db_common.DatabaseSession) (error, []string) {
		return workspace.EnsureSessionData(localCtx, sessionDataSource, conn)
	})

	return initData
}

func handleCheckInitResult(ctx context.Context, initData *control.InitData) bool {
	// if there is an error or cancellation we bomb out
	// check for the various kinds of failures
	utils.FailOnError(initData.Result.Error)
	// cancelled?
	if ctx != nil {
		utils.FailOnError(ctx.Err())
	}

	// if there is a usage warning we display it
	initData.Result.DisplayMessages()

	// if there is are any warnings, exit politely
	shouldExit := len(initData.Result.Warnings) > 0

	// alternative approach - only stop the control run if there are no controls
	//shouldExit := initData.workspace == nil || len(initData.workspace.Controls) == 0

	return shouldExit
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

	return (viper.GetBool(constants.ArgTimer) && !viper.GetBool(constants.ArgDryRun)) &&
		(outputFormat == constants.OutputFormatText || outputFormat == constants.OutputFormatBrief)
}

func validateOutputFormat() error {
	outputFormat := viper.GetString(constants.ArgOutput)
	// try to get a formatter for the desired output.
	if _, err := controldisplay.GetOutputFormatter(outputFormat); err != nil {
		// could not get a formatter
		return err
	}
	if outputFormat == constants.OutputFormatNone {
		// set progress to false
		viper.Set(constants.ArgProgress, false)
	}
	return nil
}

func validateExportTargets(exportTargets []controldisplay.CheckExportTarget) error {
	var targetErrors []error

	for _, exportTarget := range exportTargets {
		if exportTarget.Error != nil {
			targetErrors = append(targetErrors, exportTarget.Error)
		} else if _, err := controldisplay.GetExportFormatter(exportTarget.Format); err != nil {
			targetErrors = append(targetErrors, err)
		}
	}
	if len(targetErrors) > 0 {
		message := fmt.Sprintf("%d export %s failed validation", len(targetErrors), utils.Pluralize("target", len(targetErrors)))
		return utils.CombineErrorsWithPrefix(message, targetErrors...)
	}
	return nil

}

func initialiseColorScheme() error {
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

func exportCheckResult(ctx context.Context, d *control.ExportData) {
	d.WaitGroup.Add(1)
	go func() {
		err := exportControlResults(ctx, d.ExecutionTree, d.ExportFormats)
		if len(err) > 0 {
			d.AddErrors(err)
		}
		d.WaitGroup.Done()
	}()
}

func displayControlResults(ctx context.Context, executionTree *controlexecute.ExecutionTree) error {
	outputFormat := viper.GetString(constants.ArgOutput)
	formatter, _ := controldisplay.GetOutputFormatter(outputFormat)
	formattedReader, err := formatter.Format(ctx, executionTree)
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, formattedReader)

	return err
}

func exportControlResults(ctx context.Context, executionTree *controlexecute.ExecutionTree, formats []controldisplay.CheckExportTarget) []error {
	errors := []error{}
	for _, format := range formats {
		if utils.IsContextCancelled(ctx) {
			// set the error
			errors = append(errors, ctx.Err())
			// and skip forward
			continue
		}
		formatter, err := controldisplay.GetExportFormatter(format.Format)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		formattedReader, err := formatter.Format(ctx, executionTree)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if utils.IsContextCancelled(ctx) {
			errors = append(errors, err)
			continue
		}
		// create the output file
		destination, err := os.Create(format.File)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		_, err = io.Copy(destination, formattedReader)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		destination.Close()
	}

	return errors
}

func getExportTargets(executing string) ([]controldisplay.CheckExportTarget, error) {
	formats := []controldisplay.CheckExportTarget{}
	exports := viper.GetStringSlice(constants.ArgExport)
	for _, export := range exports {
		var targetError error

		if len(strings.TrimSpace(export)) == 0 {
			// if this is an empty string, ignore
			continue
		}

		parts := strings.SplitN(export, ":", 2)

		var format, fileName string

		if len(parts) == 2 {
			// we have two distinct parts - life is good
			format = parts[0]
			fileName = parts[1]
			fileName, targetError = helpers.Tildefy(fileName)
		} else {
			format = parts[0]

			// try to get an export formatter
			if formatter, fmtError := controldisplay.GetExportFormatter(format); fmtError != nil {
				// this is not a valid format. assume it is a file name
				fileName = format
				// now infer the format from the file name
				format, targetError = controldisplay.InferFormatFromExportFileName(fileName)
			} else {
				// the format was valid, generate default filename
				fileName = generateDefaultExportFileName(formatter, executing)
			}
		}
		formats = append(formats, controldisplay.NewCheckExportTarget(format, fileName, targetError))
	}
	err := validateExportTargets(formats)

	return formats, err
}

func generateDefaultExportFileName(formatter controldisplay.Formatter, executing string) string {
	now := time.Now()
	timeFormatted := fmt.Sprintf("%d%02d%02d-%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	return fmt.Sprintf("%s-%s.%s", executing, timeFormatted, formatter.FileExtension())
}
