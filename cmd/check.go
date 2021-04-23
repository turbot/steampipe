package cmd

import (
	"fmt"

	typeHelpers "github.com/turbot/go-kit/types"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/display"
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
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", []string{}, "Set a prefix to the current search path for a check session (comma-separated)")

	return cmd
}

func runCheckCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runCheckCmd start")

	defer func() {
		logging.LogTime("runCheckCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	// start db if necessary
	err := db.EnsureDbAndStartService(db.InvokerCheck)
	utils.FailOnErrorWithMessage(err, "failed to start service")
	defer db.Shutdown(nil, db.InvokerCheck)

	// load the workspace
	workspace, err := workspace.Load(viper.GetString(constants.ArgWorkspace))
	utils.FailOnErrorWithMessage(err, "failed to load workspace")
	defer workspace.Close()

	// convert the check or sql file arg into an array of executable queries - check names queries in the current workspace
	controls, err := getControls(args, workspace)
	utils.FailOnError(err)

	if len(controls) > 0 {
		// otherwise if we have resolvced any queries, run them
		failures := executeControls(controls, workspace)
		// set global exit code
		exitCode = failures
	}
}

// retrieve queries from args - for each arg check if it is a named check or a file,
// before falling back to treating it as sql
func getControls(args []string, workspace *workspace.Workspace) ([]*modconfig.Control, error) {
	var res []*modconfig.Control
	for _, arg := range args {
		if controls, ok := workspace.GetControls(arg); ok {
			res = append(res, controls...)
			continue
		} else {
			return nil, fmt.Errorf("control %s was not found in the workspace", arg)
		}
	}

	return res, nil
}

func executeControls(controls []*modconfig.Control, workspace *workspace.Workspace) int {
	// set the flag to hide spinner
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, false)

	// first get a client - do this once for all controls
	client, err := db.NewClient(true)
	utils.FailOnError(err)
	defer client.Close()

	// run all queries
	failures := 0
	for i, c := range controls {
		if err := executeControl(c, client, workspace); err != nil {
			failures++
			utils.ShowWarning(fmt.Sprintf("check #%d failed: %v", i+1, err))
		}
		if showBlankLineBetweenResults() {
			fmt.Println()
		}
	}

	return failures
}

func executeControl(control *modconfig.Control, client *db.Client, workspace *workspace.Workspace) error {
	// resolve the query patameter of the control
	query := getQueryFromArg(typeHelpers.SafeString(control.Query), workspace)
	if query == "" {
		// TODO is this an error  - for now just do nothing
		return nil
	}

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db.ExecuteQuery(query, client)
	if err != nil {
		return err
	}

	// TODO validate the result has all the required columns
	// TODO encapsulate this in display object
	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		// signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.Done()
	}
	return nil
}
