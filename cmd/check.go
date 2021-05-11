package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controlresult"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/query/queryresult"
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

	// treat aech arg as a separeate execution
	failures := 0
	for _, arg := range args {
		controls, resultTree := getControls(arg, workspace, client)
		if len(controls) == 0 {
			continue
		}
		// run the controls
		failures += executeControls(controls, resultTree, workspace, client)
	}
	// set global exit code
	exitCode = failures
}

type C9ontrolExecutor struct {
	Controls   []*modconfig.Control
	ResultTree *controlresult.ResultTree
	Workspace  *workspace.Workspace
	Client     *db.Client
}

// retrieve queries from args - for each arg check if it is a named check or a file,
// before falling back to treating it as sql
func getControls(arg string, workspace *workspace.Workspace, client *db.Client) ([]*modconfig.Control, *controlresult.ResultTree) {
	controls, resultTree := getControlsForArg(arg, workspace)

	if len(controls) == 0 {
		utils.ShowWarning(fmt.Sprintf("No controls found matching argument: %s", arg))
		return nil, nil
	}

	// 2) if a 'where' arg was used, execute this sql to get a list of  control names
	// - we then filter the controls returned by 1) with those returned by 2)
	if viper.IsSet(constants.ArgWhere) {
		whereArg := viper.GetString(constants.ArgWhere)
		filterControlNames, err := getControlsFromMetadataQuery(whereArg, workspace, client)
		utils.FailOnErrorWithMessage(err, "failed to execute '--where' SQL")
		var filteredRes []*modconfig.Control
		for _, control := range controls {
			if _, ok := filterControlNames[control.Name()]; ok {
				filteredRes = append(filteredRes, control)
			}
		}
		controls = filteredRes

		if len(controls) == 0 {
			utils.ShowWarning(fmt.Sprintf("No controls found matching argument: %s and query: %s", arg, whereArg))
		}
	}
	return controls, resultTree
}

// getControlsForArg resolves the arg into one or more controls
// It also returns the root item of the control hierarchy
//
// - if the arg is a control name, the root will be the Control with that name
// - if the arg is a benchmark name, the root will be the Benchmark with that name
// - if the arg is a mod name, the root will be the Mod with that name
// - if the arg is 'all' the root will be a node with all Mods as children
func getControlsForArg(arg string, workspace *workspace.Workspace) ([]*modconfig.Control, *controlresult.ResultTree) {
	// 1)  build list of all controls corresponding to the scope arg

	// identify the 'root' node
	var resultTree *controlresult.ResultTree
	var controls []*modconfig.Control

	// special case handling for the string "all"
	if arg == "all" {
		// get all controls from workspace
		controls := workspace.GetChildControls()
		resultTree = controlresult.NewResultTree(workspace.Mod)
		return controls, resultTree
	}

	// if arg is in fact a benchmark,  get all controls underneath the it
	name, err := modconfig.ParseResourceName(arg)
	if err != nil {
		// just log error
		log.Printf("[TRACE] error parsing check argumentr '%s': %v", arg, err)
		return nil, nil
	}
	switch name.ItemType {
	case modconfig.BlockTypeControl:
		// check whether the arg is a control name
		if control, ok := workspace.ControlMap[arg]; ok {
			controls = []*modconfig.Control{control}
		}
	case modconfig.BlockTypeBenchmark:
		// look in the workspace control group map for this control group
		if benchmark, ok := workspace.BenchmarkMap[arg]; ok {
			controls = benchmark.GetChildControls()
			resultTree = controlresult.NewResultTree(benchmark)
		}
	case modconfig.BlockTypeMod:
		// get all controls for the mod
		if mod, ok := workspace.ModMap[arg]; ok {
			controls := workspace.Mod.GetChildControls()
			resultTree = controlresult.NewResultTree(mod)
			return controls, resultTree
		}
	}
	return controls, resultTree
}

// query the steampipe_control table, using the given query
func getControlsFromMetadataQuery(whereArg string, workspace *workspace.Workspace, client *db.Client) (map[string]bool, error) {
	// query may either be a 'where' clause, or a named query
	query, isNamedQuery := getQueryFromArg(whereArg, workspace)

	// if the query is NOT a named query, we need to construct a full query by adding a select
	if !isNamedQuery {
		query = fmt.Sprintf("select resource_name from %s where %s", constants.ReflectionTableControl, whereArg)
	}

	ctx, _ := context.WithCancel(context.Background())
	res, err := client.ExecuteSync(ctx, query)
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
		rowResult := row.(*queryresult.RowResult)
		controlName := rowResult.Data[resource_name_column_index].(string)
		controlNames[controlName] = true
	}
	return controlNames, nil
}

func executeControls(controls []*modconfig.Control, resultTree *controlresult.ResultTree, workspace *workspace.Workspace, client *db.Client) int {
	// set the flag to hide spinner
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, false)

	totalControls := len(controls)
	pendingControls := len(controls)
	completeControls := 0
	errorControls := 0

	// for now we execute controls syncronously
	spinner := utils.ShowSpinner("")
	for _, c := range controls {
		p := c.Path()
		utils.UpdateSpinnerMessage(spinner, fmt.Sprintf("Running %d %s. (%d complete, %d pending, %d errors): executing \"%s\" (%s)", totalControls, utils.Pluralize("control", totalControls), completeControls, pendingControls, errorControls, typeHelpers.SafeString(c.Title), p))

		res := executeControl(c, workspace, client)
		if res.GetStatus() == controlresult.ControlRunError {
			errorControls++
		} else {
			// TODO for now this is synchronous
			completeControls++
		}
		pendingControls--

		resultTree.AddResult(res)
	}
	spinner.Stop()

	DisplayControlResults(resultTree)

	return errorControls
}

func executeControl(control *modconfig.Control, workspace *workspace.Workspace, client *db.Client) *controlresult.Result {

	controlResult := controlresult.NewControlResult(control)
	// resolve the query parameter of the control
	var query string
	// resolve the query parameter of the control
	query, _ = getQueryFromArg(typeHelpers.SafeString(control.SQL), workspace)
	if query == "" {
		controlResult.Error = fmt.Errorf(`cannot run %s - failed to resolve query "%s"`, control.Name(), typeHelpers.SafeString(control.SQL))
		return controlResult
	}

	// queryResult contains a controlResult channel
	startTime := time.Now()
	queryResult, err := client.ExecuteQuery(context.TODO(), query, false)
	if err != nil {
		controlResult.Error = err
		return controlResult
	}
	// set the control as started
	controlResult.Start(queryResult)

	// TEMPORARY - we will eventually pass the streams to the renderer before completion
	// wait for control to finish
	controlCompletionTimeout := 240 * time.Second
	for {
		// if the control is finished (either successfully or with an error), return the controlResult
		if controlResult.Finished() {
			return controlResult
		}
		time.Sleep(50 * time.Millisecond)
		if time.Since(startTime) > controlCompletionTimeout {
			controlResult.SetError(fmt.Errorf("control %s timed out", control.Name()))
		}
	}

	return controlResult
}

func DisplayControlResults(controlResults *controlresult.ResultTree) {
	// NOTE: for now we can assume all results are complete
	// todo summary and hierarchy
	for _, res := range controlResults.Root.Results {
		fmt.Println()
		fmt.Printf("%s [%s]\n", typeHelpers.SafeString(res.Control.Title), res.Control.ShortName)
		if res.Error != nil {
			fmt.Printf("  Execution error: %v\n", res.Error)
			continue
		}
		for _, item := range res.Rows {
			if item == nil {
				// should never happen!
				panic("NIL RESULT")
			}
			resString := fmt.Sprintf("  [%s] [%s] %s", item.Status, item.Resource, item.Reason)
			dimensionString := getDimension(item)
			fmt.Printf("%s %s\n", resString, dimensionString)

		}
	}
	fmt.Println()
}

func getDimension(item *controlresult.ResultRow) string {
	var dimensions []string

	for _, v := range item.Dimensions {
		dimensions = append(dimensions, v)
	}

	return strings.Join(dimensions, "  ")
}
