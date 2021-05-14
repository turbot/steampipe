package execute

import (
	"context"
	"fmt"
	"log"

	"github.com/turbot/steampipe/query/queryresult"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controlresult"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/query/execute"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
)

type ControlResolver struct {
	ResultTree *controlresult.ResultTree
	Errors     int
	Controls   []*modconfig.Control
	workspace  *workspace.Workspace
	client     *db.Client
}

// NewControlResolver creates a ControlResolver, which will execute all controls resolved from a single arg
//
// In order to build the executor:
// 1) resolve the arg into one or more controls
// 2) build the (unpopulated) ResultTree, which has a hierarchy matching the control hierarchy
func NewControlResolver(ctx context.Context, arg string, workspace *workspace.Workspace, client *db.Client) (*ControlResolver, error) {
	executor := &ControlResolver{
		workspace: workspace,
		client:    client,
	}

	var includeControlPredicate = func(string) bool { return true }
	//  if a 'where' arg was used, execute this sql to get a list of  control names
	// - we then filter the controls returned by 1) with those returned by 2)
	if viper.IsSet(constants.ArgWhere) {
		whereArg := viper.GetString(constants.ArgWhere)
		controlNameMap, err := executor.getControlsFromMetadataQuery(ctx, whereArg)
		if err != nil {
			return nil, err
		}
		includeControlPredicate = func(name string) bool {
			_, ok := controlNameMap[name]
			return ok
		}
		//
		//if len(executor.Controls) == 0 {
		//	utils.ShowWarning(fmt.Sprintf("No controls found matching argument: %s and query: %s", arg, whereArg))
		//}
	}

	// get list of controls and unpopulated result tree
	executor.getControlsForArg(arg, workspace, includeControlPredicate)

	// TODO zero control warning

	return executor, nil
}

func (e *ControlResolver) Execute(ctx context.Context) {

}

// getControlsForArg resolves the arg into one or more controls
// It also returns the ResultTree reflecting the control hierarchy, depending on the nature of the root arg
// - if the arg is a control name, the root will be the Control with that name
// - if the arg is a benchmark name, the root will be the Benchmark with that name
// - if the arg is a mod name, the root will be the Mod with that name
// - if the arg is 'all' the root will be a node with all Mods as children
func (e *ControlResolver) getControlsForArg(arg string, workspace *workspace.Workspace, includeControlPredicate func(string) bool) {
	// special case handling for the string "all"
	if arg == "all" {
		// get all controls from workspace
		e.Controls = workspace.GetChildControls()
		// build list of all workspace mods
		var mods = []modconfig.ControlTreeItem{workspace.Mod}
		for _, m := range workspace.ModMap {
			mods = append(mods, m)
		}
		e.ResultTree = controlresult.NewResultTree(includeControlPredicate, workspace, mods...)
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
			// TODO still build results tree
			// TODO delete
			e.Controls = []*modconfig.Control{control}
		}
	case modconfig.BlockTypeBenchmark:
		// look in the workspace control group map for this control group
		if benchmark, ok := workspace.BenchmarkMap[arg]; ok {

			// TODO delete
			e.Controls = benchmark.GetChildControls()

			e.ResultTree = controlresult.NewResultTree(includeControlPredicate, workspace, benchmark)
		}
	case modconfig.BlockTypeMod:
		// get all controls for the mod
		if mod, ok := workspace.ModMap[arg]; ok {
			// TODO delete
			e.Controls = workspace.Mod.GetChildControls()

			e.ResultTree = controlresult.NewResultTree(includeControlPredicate, workspace, mod)
		}
	}
}

// Get a list of controls from the reflection table steampipe_control/
// This is used to implement the `where` control filtering
func (e *ControlResolver) getControlsFromMetadataQuery(ctx context.Context, whereArg string) (map[string]bool, error) {
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
