package execute

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/query/execute"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
)

// ExecutionTree is a structure representing the control result hierarchy
type ExecutionTree struct {
	Root *ResultGroup

	workspace *workspace.Workspace
	client    *db.Client
	// an optional map of control names used to filter the controls which are run
	controlNameFilterMap map[string]bool
	progress             *ControlProgressRenderer
	// map of dimension property name to property value to color map
	DimensionColorGenerator *DimensionColorGenerator
	// flat list of all control runs
	controlRuns []*ControlRun
}

// NewExecutionTree creates a result group from a ControlTreeItem
func NewExecutionTree(ctx context.Context, workspace *workspace.Workspace, client *db.Client, arg string) (*ExecutionTree, error) {
	// now populate the ExecutionTree
	executionTree := &ExecutionTree{
		workspace: workspace,
		client:    client,
	}
	// if a "--where" parameter was passed, build a map of control manes used to filter the controls to run
	// NOTE: not enabled yet
	err := executionTree.populateControlFilterMap(ctx)

	if err != nil {
		return nil, err
	}

	// now identify the root item of the control list
	rootItems, err := executionTree.getExecutionRootFromArg(arg)
	if err != nil {
		return nil, err
	}

	// build tree of result groups, starting with a synthetic 'root' node
	executionTree.Root = NewRootResultGroup(executionTree, rootItems...)

	// after tree has built, ControlCount will be set - create progress rendered
	executionTree.progress = NewControlProgressRenderer(len(executionTree.controlRuns))

	return executionTree, nil
}

// AddControl checks whether control should be included in the tree
// if so, creates a ControlRun, which is added to the parent group
func (e *ExecutionTree) AddControl(control *modconfig.Control, group *ResultGroup) {
	// note we use short name to determine whether to include a control
	if e.ShouldIncludeControl(control.ShortName) {
		// create new ControlRun with treeItem as the parent
		controlRun := NewControlRun(control, group, e)
		// add it into the group
		group.ControlRuns = append(group.ControlRuns, controlRun)
		// also add it into the execution tree control run list
		e.controlRuns = append(e.controlRuns, controlRun)
	}
}

func (e *ExecutionTree) Execute(ctx context.Context, client *db.Client) int {
	log.Println("[TRACE]", "begin ExecutionTree.Execute")
	defer log.Println("[TRACE]", "end ExecutionTree.Execute")
	e.progress.Start()
	defer e.progress.Finish()
	// just execute the root - it will traverse the tree
	errors := e.Root.Execute(ctx, client)
	// now build map of dimension property name to property value to color map
	e.DimensionColorGenerator, _ = NewDimensionColorGenerator(4, 27)
	e.DimensionColorGenerator.populate(e)

	return errors
}

func (e *ExecutionTree) populateControlFilterMap(ctx context.Context) error {
	//  if a 'where' arg was used, execute this sql to get a list of  control names
	// use this list to build a name map used to determine whether to run a particular control
	if viper.IsSet(constants.ArgWhere) {
		whereArg := viper.GetString(constants.ArgWhere)
		var err error
		e.controlNameFilterMap, err = e.getControlMapFromMetadataQuery(ctx, whereArg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *ExecutionTree) ShouldIncludeControl(controlName string) bool {
	if e.controlNameFilterMap == nil {
		return true
	}
	_, ok := e.controlNameFilterMap[controlName]
	return ok
}

// getExecutionRootFromArg resolves the arg into the execution root
// - if the arg is a control name, the root will be the Control with that name
// - if the arg is a benchmark name, the root will be the Benchmark with that name
// - if the arg is a mod name, the root will be the Mod with that name
// - if the arg is 'all' the root will be a node with all Mods as children
func (e *ExecutionTree) getExecutionRootFromArg(arg string) ([]modconfig.ControlTreeItem, error) {
	var res []modconfig.ControlTreeItem
	// special case handling for the string "all"
	if arg == "all" {
		//
		// build list of all workspace mods - these will act as root items
		for _, m := range e.workspace.ModMap {
			res = append(res, m)
		}
		return res, nil
	}

	// what resource type is arg?
	name, err := modconfig.ParseResourceName(arg)
	if err != nil {
		// just log error
		return nil, fmt.Errorf("failed to parse check argument '%s': %v", arg, err)
	}

	switch name.ItemType {
	case modconfig.BlockTypeControl:
		// check whether the arg is a control name
		if control, ok := e.workspace.ControlMap[arg]; ok {
			return []modconfig.ControlTreeItem{control}, nil
		}
	case modconfig.BlockTypeBenchmark:
		// look in the workspace control group map for this control group
		if benchmark, ok := e.workspace.BenchmarkMap[arg]; ok {
			return []modconfig.ControlTreeItem{benchmark}, nil
		}
	case modconfig.BlockTypeMod:
		// get all controls for the mod
		if mod, ok := e.workspace.ModMap[arg]; ok {
			return []modconfig.ControlTreeItem{mod}, nil
		}
	}
	return nil, fmt.Errorf("no controls found matching argument '%s'", arg)
}

// Get a map of control names from the reflection table steampipe_control
// This is used to implement the `where` control filtering
func (e *ExecutionTree) getControlMapFromMetadataQuery(ctx context.Context, whereArg string) (map[string]bool, error) {
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
	resourceNameColumnIndex := -1

	for i, c := range res.ColTypes {
		if c.Name() == "resource_name" {
			resourceNameColumnIndex = i
		}
	}
	if resourceNameColumnIndex == -1 {
		return nil, fmt.Errorf("the named query passed in the 'where' argument must return the 'resource_name' column")
	}

	var controlNames = make(map[string]bool)
	for _, row := range res.Rows {
		rowResult := row.(*queryresult.RowResult)
		controlName := rowResult.Data[resourceNameColumnIndex].(string)
		controlNames[controlName] = true
	}
	return controlNames, nil
}

func (e *ExecutionTree) GetAllTags() []string {
	tagColumnMap := make(map[string]bool) // to keep track which tags have been added as columns
	var tagColumns []string
	for _, r := range e.controlRuns {
		if r.Control.Tags != nil {
			for tag := range *r.Control.Tags {
				if !tagColumnMap[tag] {
					tagColumns = append(tagColumns, tag)
					tagColumnMap[tag] = true
				}
			}
		}
	}
	return tagColumns
}
