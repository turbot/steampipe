package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/definitions/results"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
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
		AddStringFlag(constants.ArgWhere, "", "", "SQL 'where' clause , or named query, used to filter controls ")

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

	// first get a client - do this once for all controls
	client, err := db.NewClient(true)
	utils.FailOnError(err)
	defer client.Close()

	// populate the reflection tables
	err = db.CreateMetadataTables(workspace.GetResourceMaps(), client)
	utils.FailOnError(err)

	// convert the check or sql file arg into an array of executable queries - check names queries in the current workspace
	controls, err := getControls(args, workspace, client)
	utils.FailOnError(err)

	if len(controls) > 0 {
		// otherwise if we have resolved any queries, run them
		failures := executeControls(controls, workspace, client)
		// set global exit code
		exitCode = failures
	}
}

// retrieve queries from args - for each arg check if it is a named check or a file,
// before falling back to treating it as sql
func getControls(args []string, workspace *workspace.Workspace, client *db.Client) ([]*modconfig.Control, error) {

	// 1)  build list of all controls corresponding to the scope args
	var res []*modconfig.Control
	for _, arg := range args {
		if controls := workspace.GetControlsForArg(arg); len(controls) > 0 {
			res = append(res, controls...)
		}
	}
	if len(res) == 0 {
		utils.ShowWarning(fmt.Sprintf("No controls found matching %s: %s", utils.Pluralize("argument", len(args)), strings.Join(args, ",")))
		return res, nil
	}

	// 2) if a 'where' arg was used, execute this sql to get a list of  control names
	// - we then filter the controls returned by 1) with those returned by 2)
	if viper.IsSet(constants.ArgWhere) {
		whereArg := viper.GetString(constants.ArgWhere)
		filterControlNames, err := getControlsFromMetadataQuery(whereArg, workspace, client)
		utils.FailOnErrorWithMessage(err, "failed to execute '--where' SQL")
		var filteredRes []*modconfig.Control
		for _, control := range res {
			if _, ok := filterControlNames[control.Name()]; ok {
				filteredRes = append(filteredRes, control)
			}
		}
		res = filteredRes

		if len(res) == 0 {
			utils.ShowWarning(fmt.Sprintf("No controls found matching %s: %s and query: %s", utils.Pluralize("argument", len(args)), args, whereArg))
		}
	}
	return res, nil
}

// query the steampipe_controls table, using the given query
func getControlsFromMetadataQuery(whereArg string, workspace *workspace.Workspace, client *db.Client) (map[string]bool, error) {
	// query may either be a 'where' clause, or a named query
	query, isNamedQuery := getQueryFromArg(whereArg, workspace)

	// if the query is NOT a named query, we need to construct a full query by adding a select
	if !isNamedQuery {
		query = fmt.Sprintf("select resource_name from steampipe_controls where %s", whereArg)
	}

	res, err := client.ExecuteSync(query)
	if err != nil {
		return nil, err
	}

	//
	// find the "resource_name" column index
	resource_name_column_index := -1

	for i, c := range res.ColTypes {
		if c.Name() == "resource_name" {
			resource_name_column_index = i
		}
	}
	if resource_name_column_index == -1 {
		return nil, fmt.Errorf("the named query passed in the 'where' argument must return the 'resource_name' column")
	}

	var controlNames = make(map[string]bool)
	for _, row := range res.Rows {
		rowResult := row.(*results.RowResult)
		controlName := rowResult.Data[resource_name_column_index].(string)
		controlNames[controlName] = true
	}
	return controlNames, nil
}

func executeControls(controls []*modconfig.Control, workspace *workspace.Workspace, client *db.Client) int {
	// set the flag to hide spinner
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, false)

	// run all queries
	failures := 0
	for i, c := range controls {
		if err := executeControl(c, workspace, client); err != nil {
			failures++
			utils.ShowWarning(fmt.Sprintf("check #%d failed: %v", i+1, err))
		}
		if showBlankLineBetweenResults() {
			fmt.Println()
		}
	}

	return failures
}

func executeControl(control *modconfig.Control, workspace *workspace.Workspace, client *db.Client) error {
	// resolve the query patameter of the control
	query, _ := getQueryFromArg(typeHelpers.SafeString(control.Query), workspace)
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
