package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controldisplay"
	"github.com/turbot/steampipe/control/execute"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

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
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgHeader, "", true, "Include column headers csv and table output").
		AddStringFlag(constants.ArgSeparator, "", ",", "Separator string for csv output").
		AddStringFlag(constants.ArgOutput, "", "text", "Select the console output format. Possible values are json, text, brief, none").
		AddBoolFlag(constants.ArgTimer, "", false, "Turn on the timer which reports check time.").
		AddBoolFlag(constants.ArgWatch, "", true, "Watch SQL files in the current workspace (works only in interactive mode)").
		AddStringSliceFlag(constants.ArgSearchPath, "", []string{}, "Set a custom search_path for the steampipe user for a check session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", []string{}, "Set a prefix to the current search path for a check session (comma-separated)").
		AddStringFlag(constants.ArgTheme, "", "dark", "Set the output theme, which determines the color scheme for the 'text' control output. Possible values are light, dark, plain").
		AddStringSliceFlag(constants.ArgExport, "", []string{}, "Export output to files. Multiple exports are allowed.").
		AddBoolFlag(constants.ArgProgress, "", true, "Display control execution progress").
		AddBoolFlag(constants.ArgDryRun, "", false, "Show which controls will be run without running them").
		AddStringFlag(constants.ArgWhere, "", "", "SQL 'where' clause , or named query, used to filter controls. Cannot be used with '--tag'").
		AddStringSliceFlag(constants.ArgTag, "", []string{}, "Key-Value pairs to filter controls based on the 'tags' property. To be provided as 'key=value'. Multiple can be given and are merged together. Cannot be used with '--where'")

	return cmd
}

func runCheckCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runCheckCmd start")
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, false)

	// verify we have an argument
	if len(args) == 0 {
		fmt.Println()
		utils.ShowError(fmt.Errorf("you must provide at least one argument"))
		fmt.Println()
		cmd.Help()
		fmt.Println()
		return
	}

	defer func() {
		utils.LogTime("runCheckCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	err := validateOutputFormat()
	utils.FailOnError(err)

	ctx, cancel := context.WithCancel(context.Background())
	startCancelHandler(cancel)

	// start db if necessary
	err = db.EnsureDbAndStartService(db.InvokerCheck).Error
	utils.FailOnErrorWithMessage(err, "failed to start service")
	defer db.Shutdown(nil, db.InvokerCheck)

	// set color schema
	err = initialiseColorScheme()
	utils.FailOnError(err)

	// load the workspace
	workspace, err := workspace.Load(viper.GetString(constants.ArgWorkspace))
	utils.FailOnErrorWithMessage(err, "failed to load workspace")
	defer workspace.Close()

	// check if the required plugins are installed
	err = workspace.CheckRequiredPluginsInstalled()
	utils.FailOnError(err)

	if len(workspace.ControlMap) == 0 {
		utils.ShowWarning("no controls found in current workspace")
		return
	}

	// first get a client - do this once for all controls
	client, err := db.NewClient(true)
	utils.FailOnError(err)
	defer client.Close()

	// populate the reflection tables
	err = db.CreateMetadataTables(workspace.GetResourceMaps(), client)
	utils.FailOnError(err)

	// treat each arg as a separate execution
	failures := 0
	var exportErrors []error
	exportErrorsLock := sync.Mutex{}

	exportWaitGroup := sync.WaitGroup{}
	durations := []time.Duration{}
	for _, arg := range args {
		// get the export formats for this argument
		exportFormats := getExportTargets(arg)
		err = validateExportTargets(exportFormats)
		if err != nil {
			utils.ShowError(err)
			return
		}

		select {
		case <-ctx.Done():
			durations = append(durations, 0)
			// skip over the next, since the execution was cancelled
			continue
		default:
			executionTree, err := execute.NewExecutionTree(ctx, workspace, client, arg)
			utils.FailOnErrorWithMessage(err, "failed to resolve controls from argument")

			// for now we execute controls synchronously
			// Execute returns the number of failures
			failures += executionTree.Execute(ctx, client)
			err = displayControlResults(ctx, executionTree)
			utils.FailOnError(err)

			exportWaitGroup.Add(1)
			go func() {
				err := exportControlResults(ctx, executionTree, exportFormats)
				if err != nil {
					exportErrorsLock.Lock()
					exportErrors = append(exportErrors, err...)
					exportErrorsLock.Unlock()
				}
				exportWaitGroup.Done()
			}()
			durations = append(durations, executionTree.Root.Duration)
		}
	}

	exportWaitGroup.Wait()

	if len(exportErrors) > 0 {
		utils.ShowError(utils.CombineErrors(exportErrors...))
	}

	if shouldPrintTiming() {
		headers := []string{"", "Duration"}
		rows := [][]string{}
		for idx, arg := range args {
			rows = append(rows, []string{arg, durations[idx].String()})
		}
		fmt.Println("Timing:")
		display.ShowWrappedTable(headers, rows, false)
	}

	// set global exit code
	exitCode = failures
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

func displayControlResults(ctx context.Context, executionTree *execute.ExecutionTree) error {
	outputFormat := viper.GetString(constants.ArgOutput)
	formatter, _ := controldisplay.GetOutputFormatter(outputFormat)
	formattedReader, err := formatter.Format(ctx, executionTree)
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, formattedReader)

	return err
}

func exportControlResults(ctx context.Context, executionTree *execute.ExecutionTree, formats []controldisplay.CheckExportTarget) []error {
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

func getExportTargets(executing string) []controldisplay.CheckExportTarget {
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
	return formats
}

func generateDefaultExportFileName(format string, executing string) string {
	return fmt.Sprintf("%s-%s.%s", executing, time.Now().UTC().Format("20060102150405Z"), format)
}
