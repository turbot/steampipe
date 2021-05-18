package cmd

import (
	"context"
	"fmt"

	"github.com/karrick/gows"
	"github.com/turbot/steampipe/control/controldisplay"
	"github.com/turbot/steampipe/control/execute"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

// CheckCmd :: represents the check command
func CheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "check",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runCheckCmd,
		Short:            "Execute one or more controls",
		Long:             `Execute one or more controls."`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgHeader, "", true, "Include column headers csv and table output").
		AddStringFlag(constants.ArgSeparator, "", ",", "Separator string for csv output").
		AddStringFlag(constants.ArgOutput, "", "table", "Output format: line, csv, json or table").
		AddBoolFlag(constants.ArgTimer, "", false, "Turn on the timer which reports check time.").
		AddBoolFlag(constants.ArgWatch, "", true, "Watch SQL files in the current workspace (works only in interactive mode)").
		AddStringSliceFlag(constants.ArgSearchPath, "", []string{}, "Set a custom search_path for the steampipe user for a check session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", []string{}, "Set a prefix to the current search path for a check session (comma-separated)").
		AddStringFlag(constants.ArgWhere, "", "", "SQL 'where' clause , or named query, used to filter controls ").
		AddStringFlag(constants.ArgTheme, "", "dark", "Color scheme").
		AddBoolFlag(constants.ArgProgress, "", true, "Display control execution progress").
		AddBoolFlag(constants.ArgQuiet, "", false, "Display only failed control results").
		AddBoolFlag(constants.ArgColor, "", true, "Display control results in color").
		AddBoolFlag(constants.ArgDryRun, "", false, "Show which controls will be run without running them")

	return cmd
}

func runCheckCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runCheckCmd start")
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, false)

	defer func() {
		logging.LogTime("runCheckCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	startCancelHandler(cancel)

	// start db if necessary
	err := db.EnsureDbAndStartService(db.InvokerCheck)
	utils.FailOnErrorWithMessage(err, "failed to start service")
	defer db.Shutdown(nil, db.InvokerCheck)

	// set color schema
	err = initialiseColorScheme()
	utils.FailOnError(err)

	// load the workspace
	workspace, err := workspace.Load(viper.GetString(constants.ArgWorkspace))
	utils.FailOnErrorWithMessage(err, "failed to load workspace")
	defer workspace.Close()

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
			DisplayControlResults(ctx, executionTree)
		}
	}

	// set global exit code
	exitCode = failures
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

func DisplayControlResults(ctx context.Context, executionTree *execute.ExecutionTree) {
	//bytes, err := json.MarshalIndent(executionTree.Root, "", "  ")
	maxCols := getMaxCols()

	renderer := controldisplay.NewTableRenderer(executionTree, maxCols)

	if ctx.Err() != nil {
		utils.ShowError(ctx.Err())
	}

	fmt.Println(renderer.Render())
}

func getMaxCols() int {
	maxCols, _, _ := gows.GetWinSize()
	// limit to 200
	if maxCols > 200 {
		maxCols = 120
	}
	return maxCols
}
