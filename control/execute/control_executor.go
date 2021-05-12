package execute

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/turbot/steampipe/query/queryresult"

	"github.com/spf13/viper"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controlresult"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/query/execute"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

type ControlExecutor struct {
	Controls   []*modconfig.Control
	ResultTree *controlresult.ResultTree
	Errors     int
	workspace  *workspace.Workspace
	client     *db.Client
}

// NewExecutor creates a ControlExecutor, which will execute all controls resolved from a single arg
//
// In order to build the executor:
// 1) resolve the arg into one or more controls
// 2) build the (unpopulated) ResultTree, which has a hierarchy matching the control hierarchy
func NewExecutor(ctx context.Context, arg string, workspace *workspace.Workspace, client *db.Client) *ControlExecutor {
	executor := &ControlExecutor{
		workspace: workspace,
		client:    client,
	}

	// get list of controls and unpopulated result tree
	executor.getControlsForArg(arg, workspace)
	if len(executor.Controls) == 0 {
		utils.ShowWarning(fmt.Sprintf("No controls found matching argument: %s", arg))
		return executor
	}

	// 2) if a 'where' arg was used, execute this sql to get a list of  control names
	// - we then filter the controls returned by 1) with those returned by 2)
	if viper.IsSet(constants.ArgWhere) {
		whereArg := viper.GetString(constants.ArgWhere)
		executor.filterControlsWithWhereClause(ctx, whereArg)

		if len(executor.Controls) == 0 {
			utils.ShowWarning(fmt.Sprintf("No controls found matching argument: %s and query: %s", arg, whereArg))
		}
	}

	return executor
}

func (e *ControlExecutor) Execute(ctx context.Context) {
	// set the flag to hide spinner
	cmdconfig.Viper().Set(constants.ConfigKeyShowInteractiveOutput, false)

	totalControls := len(e.Controls)
	pendingControls := totalControls
	completeControls := 0
	errorControls := 0

	// for now we execute controls syncronously
	spinner := utils.ShowSpinner("")
	for _, c := range e.Controls {
		p := c.Path()
		utils.UpdateSpinnerMessage(spinner, fmt.Sprintf("Running %d %s. (%d complete, %d pending, %d errors): executing \"%s\" (%s)", totalControls, utils.Pluralize("control", totalControls), completeControls, pendingControls, errorControls, typeHelpers.SafeString(c.Title), p))

		res := e.executeControl(ctx, c)
		if res.GetRunStatus() == controlresult.ControlRunError {
			errorControls++
		} else {
			completeControls++
		}
		pendingControls--

		e.ResultTree.AddResult(res)
	}
	spinner.Stop()

	e.Errors = errorControls

}

func (e *ControlExecutor) executeControl(ctx context.Context, control *modconfig.Control) *controlresult.ControlRun {
	controlRun := controlresult.NewControlRun(control)
	// resolve the query parameter of the control
	var query string
	// resolve the query parameter of the control
	query, _ = execute.GetQueryFromArg(typeHelpers.SafeString(control.SQL), e.workspace)
	if query == "" {
		controlRun.Error = fmt.Errorf(`cannot run %s - failed to resolve query "%s"`, control.Name(), typeHelpers.SafeString(control.SQL))
		return controlRun
	}

	startTime := time.Now()
	queryResult, err := e.client.ExecuteQuery(ctx, query, false)
	if err != nil {
		controlRun.Error = err
		return controlRun
	}
	// set the control as started
	controlRun.Start(queryResult)

	// TEMPORARY - we will eventually pass the streams to the renderer before completion
	// wait for control to finish
	controlCompletionTimeout := 240 * time.Second
	for {
		// if the control is finished (either successfully or with an error), return the controlRun
		if controlRun.Finished() {
			break
		}
		time.Sleep(50 * time.Millisecond)
		if time.Since(startTime) > controlCompletionTimeout {
			controlRun.SetError(fmt.Errorf("control %s timed out", control.Name()))
		}
	}

	return controlRun
}

// getControlsForArg resolves the arg into one or more controls
// It also returns the ResultTree reflecting the control hierarchy, depending on the nature of the root arg
// - if the arg is a control name, the root will be the Control with that name
// - if the arg is a benchmark name, the root will be the Benchmark with that name
// - if the arg is a mod name, the root will be the Mod with that name
// - if the arg is 'all' the root will be a node with all Mods as children
func (e *ControlExecutor) getControlsForArg(arg string, workspace *workspace.Workspace) {
	// special case handling for the string "all"
	if arg == "all" {
		// get all controls from workspace
		e.Controls = workspace.GetChildControls()
		// build list of all workspace mods
		var mods = []modconfig.ControlTreeItem{workspace.Mod}
		for _, m := range workspace.ModMap {
			mods = append(mods, m)
		}
		e.ResultTree = controlresult.NewResultTree(mods...)
		return
	}

	// what resource type is arg?
	name, err := modconfig.ParseResourceName(arg)
	if err != nil {
		// just log error
		log.Printf("[TRACE] error parsing check argumentr '%s': %v", arg, err)
		return
	}
	switch name.ItemType {
	case modconfig.BlockTypeControl:
		// check whether the arg is a control name
		if control, ok := workspace.ControlMap[arg]; ok {
			e.Controls = []*modconfig.Control{control}
		}
	case modconfig.BlockTypeBenchmark:
		// look in the workspace control group map for this control group
		if benchmark, ok := workspace.BenchmarkMap[arg]; ok {
			// TODO it would be nice to resolve the control parents in the case of multiple parents
			e.Controls = benchmark.GetChildControls()
			e.ResultTree = controlresult.NewResultTree(benchmark)
		}
	case modconfig.BlockTypeMod:
		// get all controls for the mod
		if mod, ok := workspace.ModMap[arg]; ok {
			e.Controls = workspace.Mod.GetChildControls()
			e.ResultTree = controlresult.NewResultTree(mod)
		}
	}
}

func (e *ControlExecutor) filterControlsWithWhereClause(ctx context.Context, whereArg string) {
	filterControlNames, err := e.getControlsFromMetadataQuery(ctx, whereArg)
	utils.FailOnErrorWithMessage(err, "failed to execute '--where' SQL")
	var filteredRes []*modconfig.Control
	for _, control := range e.Controls {
		if _, ok := filterControlNames[control.Name()]; ok {
			filteredRes = append(filteredRes, control)
		}
	}
	e.Controls = filteredRes
}

// Get a list of controls from the reflection table steampipe_control/
// This is used to implement the `where` control filtering
func (e *ControlExecutor) getControlsFromMetadataQuery(ctx context.Context, whereArg string) (map[string]bool, error) {
	// query may either be a 'where' clause, or a named query
	query, isNamedQuery := execute.GetQueryFromArg(whereArg, e.workspace)

	// if the query is NOT a named query, we need to construct a full query by adding a select
	if !isNamedQuery {
		query = fmt.Sprintf("select resource_name from %s where %s", constants.ReflectionTableControl, whereArg)
	}

	res, err := e.client.ExecuteSync(ctx, query)
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
