package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/karrick/gows"
	"github.com/turbot/steampipe/control/controldisplay"
	"github.com/turbot/steampipe/control/execute"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

// CheckCmd :: represents the check command
func CheckCmd() *cobra.Command {
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
		//AddStringFlag(constants.ArgWhere, "", "", "SQL 'where' clause , or named query, used to filter controls ").
		AddStringFlag(constants.ArgTheme, "", "dark", "Set the output theme, which determines the color scheme for the 'text' control output. Possible values are light, dark, plain").
		AddBoolFlag(constants.ArgProgress, "", true, "Display control execution progress")

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
	err = db.EnsureDbAndStartService(db.InvokerCheck)
	utils.FailOnErrorWithMessage(err, "failed to start service")
	defer db.Shutdown(nil, db.InvokerCheck)

	// set color schema
	err = initialiseColorScheme()
	utils.FailOnError(err)

	// load the workspace
	workspace, err := workspace.Load(viper.GetString(constants.ArgWorkspace))
	utils.FailOnErrorWithMessage(err, "failed to load workspace")
	defer workspace.Close()

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
	for _, arg := range args {
		select {
		case <-ctx.Done():
			// skip over the next, since the execution was cancelled
			continue
		default:
			executionTree, err := execute.NewExecutionTree(ctx, workspace, client, arg)
			utils.FailOnErrorWithMessage(err, "failed to resolve controls from argument")

			// for now we execute controls synchronously
			// Execute returns the number of failures
			executionTree.Execute(ctx, client)
			err = DisplayControlResults(ctx, executionTree)
			utils.FailOnError(err)
		}
	}

	// set global exit code
	exitCode = failures
}

func validateOutputFormat() error {
	outputFormat := viper.GetString(constants.ArgOutput)
	if !helpers.StringSliceContains([]string{"text", "brief", "json", "none"}, outputFormat) {
		return fmt.Errorf("invalid output format '%s' - must be one of json, text, brief, none", outputFormat)
	}
	if outputFormat == "none" {
		// set progress to false
		viper.Set(constants.ArgProgress, false)
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

func DisplayControlResults(ctx context.Context, executionTree *execute.ExecutionTree) (err error) {
	outputFormat := viper.GetString(constants.ArgOutput)

	switch outputFormat {
	case "text", "brief":
		err = displayTextOutput(ctx, executionTree)
	case "csv":
		err = displayCsvOutput(ctx, executionTree)
	case "json":
		err = displayJsonOutput(ctx, executionTree)
	case "none":
		// set progress to false
		viper.Set(constants.ArgProgress, false)
		// no error, nothing else to do
	default:
		err = fmt.Errorf("invalid output format '%s' - must be one of json,csv,text,brief,none", outputFormat)
	}

	return
}

func displayCsvOutput(context.Context, *execute.ExecutionTree) error {
	return fmt.Errorf("CSV output not supported yet")
}

func displayJsonOutput(ctx context.Context, tree *execute.ExecutionTree) error {
	bytes, err := json.MarshalIndent(tree.Root, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))
	return nil
}

func displayTextOutput(ctx context.Context, executionTree *execute.ExecutionTree) error {
	maxCols := getMaxCols()

	renderer := controldisplay.NewTableRenderer(executionTree, maxCols)

	if ctx.Err() != nil {
		utils.ShowError(ctx.Err())
	}

	fmt.Println(renderer.Render())
	return nil
}

func getMaxCols() int {
	maxCols, _, _ := gows.GetWinSize()
	// limit to 200
	if maxCols > 200 {
		maxCols = 200
	}
	return maxCols
}
