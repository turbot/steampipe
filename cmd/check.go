package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/turbot/steampipe/db/db_client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controldisplay"
	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

type checkInitData struct {
	ctx           context.Context
	workspace     *workspace.Workspace
	client        db_common.Client
	dbInitialised bool
	result        *db_common.InitResult
}

type exportData struct {
	executionTree *controlexecute.ExecutionTree
	exportFormats []controldisplay.CheckExportTarget
	errorsLock    *sync.Mutex
	errors        []error
	waitGroup     *sync.WaitGroup
}

// checkCmd :: represents the check command
func checkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "check [flags] [mod/benchmark/control/\"all\"]",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runCheckCmd,
		Short:            "Execute one or more controls",
		Long: `Execute one of more Steampipe benchmarks and controls.

You may specify one or more benchmarks or controls to run (separated by a space), or run 'steampipe check all' to run all controls in the workspace.`,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			workspaceResources, err := workspace.LoadResourceNames(viper.GetString(constants.ArgWorkspace))
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
		AddBoolFlag(constants.ArgHeader, "", true, "Include column headers csv and table output").
		AddStringFlag(constants.ArgSeparator, "", ",", "Separator string for csv output").
		AddStringFlag(constants.ArgOutput, "", "text", "Select the console output format. Possible values are json, text, brief, none").
		AddStringFlag(constants.ArgConnectionString, "", "", "Database connection string - used to connect to remote database instead of running database locally").
		AddStringFlag(constants.ArgDatabase, "", "", "The remote database to connect to. If specified, api-key must also be passed").
		AddStringFlag(constants.ArgAPIKey, "", "", "The steampipe cloud api-key to use to retrieve database details").
		AddBoolFlag(constants.ArgTimer, "", false, "Turn on the timer which reports check time.").
		AddStringSliceFlag(constants.ArgSearchPath, "", nil, "Set a custom search_path for the steampipe user for a check session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", nil, "Set a prefix to the current search path for a check session (comma-separated)").
		AddStringFlag(constants.ArgTheme, "", "dark", "Set the output theme, which determines the color scheme for the 'text' control output. Possible values are light, dark, plain").
		AddStringSliceFlag(constants.ArgExport, "", nil, "Export output to files. Multiple exports are allowed.").
		AddBoolFlag(constants.ArgProgress, "", true, "Display control execution progress").
		AddBoolFlag(constants.ArgDryRun, "", false, "Show which controls will be run without running them").
		AddStringFlag(constants.ArgWhere, "", "", "SQL 'where' clause , or named query, used to filter controls. Cannot be used with '--tag'").
		AddStringSliceFlag(constants.ArgTag, "", nil, "Key-Value pairs to filter controls based on the 'tags' property. To be provided as 'key=value'. Multiple can be given and are merged together. Cannot be used with '--where'").
		AddStringSliceFlag(constants.ArgVarFile, "", nil, "Specify a file containing variable values").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, "", nil, "Specify The value of a variable")

	return cmd
}

func runCheckCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runCheckCmd start")
	initData := &checkInitData{}
	defer func() {
		utils.LogTime("runCheckCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}

		if initData.client != nil {
			initData.client.Close()
		}
		if initData.workspace != nil {
			initData.workspace.Close()

		}
	}()

	// verify we have an argument
	if !validateArgs(cmd, args) {
		return
	}

	// initialise
	initData = initialiseCheck()
	if shouldExit := handleCheckInitResult(initData); shouldExit {
		return
	}

	// pull out useful properties
	ctx := initData.ctx
	workspace := initData.workspace
	client := initData.client
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
			d := &exportData{executionTree: executionTree, exportFormats: exportFormats, errorsLock: &exportErrorsLock, errors: exportErrors, waitGroup: &exportWaitGroup}
			exportCheckResult(ctx, d)
		}

		durations = append(durations, executionTree.Root.Duration)
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

func initialiseCheck() *checkInitData {
	initData := &checkInitData{
		result: &db_common.InitResult{},
	}

	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, false)

	err := validateOutputFormat()
	if err != nil {
		initData.result.Error = err
		return initData
	}

	err = validateConnectionStringArgs()
	utils.FailOnError(err)

	ctx, cancel := context.WithCancel(context.Background())
	startCancelHandler(cancel)
	initData.ctx = ctx

	// set color schema
	err = initialiseColorScheme()
	if err != nil {
		initData.result.Error = err
		return initData
	}

	// load workspace
	initData.workspace, err = loadWorkspacePromptingForVariables(ctx)
	if err != nil {
		if !utils.IsCancelledError(err) {
			err = utils.PrefixError(err, "failed to load workspace")
		}
		initData.result.Error = err
		return initData
	}

	// check if the required plugins are installed
	initData.result.Error = initData.workspace.CheckRequiredPluginsInstalled()
	if len(initData.workspace.ControlMap) == 0 {
		initData.result.AddWarnings("no controls found in current workspace")
	}

	// get a client
	var client db_common.Client
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		client, err = db_client.NewDbClient(connectionString)
	} else {
		client, err = db_local.GetLocalClient(constants.InvokerCheck)
	}

	if err != nil {
		initData.result.Error = err
		return initData
	}
	initData.client = client

	refreshResult := initData.client.RefreshConnectionAndSearchPaths()
	if refreshResult.Error != nil {
		initData.result.Error = refreshResult.Error
		return initData
	}
	initData.result.AddWarnings(refreshResult.Warnings...)

	// create the prepared statements
	err = db_common.CreatePreparedStatements(ctx, initData.workspace.GetResourceMaps(), initData.client)
	if err != nil {
		initData.result.Error = err
		return initData
	}

	// populate the introspection tables
	err = db_common.CreateIntrospectionTables(ctx, initData.workspace.GetResourceMaps(), initData.client)
	if err != nil {
		initData.result.Error = err
		return initData
	}

	return initData
}

func handleCheckInitResult(initData *checkInitData) bool {
	shouldExit := false
	// if there is an error or cancellation we bomb out
	// check for the various kinds of failures
	utils.FailOnError(initData.result.Error)
	// cancelled?
	if initData.ctx != nil {
		utils.FailOnError(initData.ctx.Err())
	}

	// if there is a usage warning we display it and exit politely
	initData.result.DisplayMessages()
	shouldExit = len(initData.result.Warnings) > 0

	return shouldExit
}

func exportCheckResult(ctx context.Context, d *exportData) {
	d.waitGroup.Add(1)
	go func() {
		err := exportControlResults(ctx, d.executionTree, d.exportFormats)
		if err != nil {
			d.errorsLock.Lock()
			d.errors = append(d.errors, err...)
			d.errorsLock.Unlock()
		}
		d.waitGroup.Done()
	}()
}

func printTiming(args []string, durations []time.Duration) {
	headers := []string{"", "Duration"}
	var rows [][]string
	for idx, arg := range args {
		rows = append(rows, []string{arg, durations[idx].String()})
	}
	fmt.Println("Timing:")
	display.ShowWrappedTable(headers, rows, false)
}

func validateArgs(cmd *cobra.Command, args []string) bool {
	if len(args) == 0 {
		fmt.Println()
		utils.ShowError(fmt.Errorf("you must provide at least one argument"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		return false
	}
	return true
}

func shouldPrintTiming() bool {
	outputFormat := viper.GetString(constants.ArgOutput)

	return ((viper.GetBool(constants.ArgTimer) && !viper.GetBool(constants.ArgDryRun)) &&
		(outputFormat == controldisplay.OutputFormatText || outputFormat == controldisplay.OutputFormatBrief))
}

func validateOutputFormat() error {
	outputFormat := viper.GetString(constants.ArgOutput)
	// try to get a formatter for the desired output.
	if _, err := controldisplay.GetOutputFormatter(outputFormat); err != nil {
		// could not get a formatter
		return err
	}
	if outputFormat == controldisplay.OutputFormatNone {
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
			if _, fmtError := controldisplay.GetExportFormatter(format); fmtError != nil {
				// this is not a valid format. assume it is a file name
				fileName = format
				// now infer the format from the file name
				format, targetError = controldisplay.InferFormatFromExportFileName(fileName)
			} else {
				// the format was valid, generate default filename
				fileName = generateDefaultExportFileName(format, executing)
			}
		}
		formats = append(formats, controldisplay.NewCheckExportTarget(format, fileName, targetError))
	}
	err := validateExportTargets(formats)

	return formats, err
}

func generateDefaultExportFileName(format string, executing string) string {
	return fmt.Sprintf("%s-%s.%s", executing, time.Now().UTC().Format("20060102150405Z"), format)
}
