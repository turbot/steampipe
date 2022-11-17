package controlexecute

import (
	"context"
	"fmt"
	"github.com/turbot/go-kit/helpers"
	"log"
	"sort"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/workspace"
	"golang.org/x/sync/semaphore"
)

// ExecutionTree is a structure representing the control execution hierarchy
type ExecutionTree struct {
	Root *ResultGroup `json:"root"`
	// flat list of all control runs
	ControlRuns []*ControlRun                  `json:"-"`
	StartTime   time.Time                      `json:"start_time"`
	EndTime     time.Time                      `json:"end_time"`
	Progress    *controlstatus.ControlProgress `json:"progress"`
	// map of dimension property name to property value to color map
	DimensionColorGenerator *DimensionColorGenerator `json:"-"`
	// the current session search path (this may be overidden for specific controls)
	SearchPath []string             `json:"-"`
	Workspace  *workspace.Workspace `json:"-"`
	client     db_common.Client
	// an optional map of control names used to filter the controls which are run
	controlNameFilterMap map[string]bool
}

func NewExecutionTree(ctx context.Context, workspace *workspace.Workspace, client db_common.Client, arg, controlFilterWhereClause string) (*ExecutionTree, error) {
	// now populate the ExecutionTree
	executionTree := &ExecutionTree{
		Workspace:  workspace,
		client:     client,
		SearchPath: utils.UnquoteStringArray(client.GetRequiredSessionSearchPath()),
	}
	// if a "--where" or "--tag" parameter was passed, build a map of control names used to filter the controls to run
	// create a context with status hooks disabled
	noStatusCtx := statushooks.DisableStatusHooks(ctx)
	err := executionTree.populateControlFilterMap(noStatusCtx, controlFilterWhereClause)
	if err != nil {
		return nil, err
	}

	// now identify the root item of the control list
	rootItem, err := executionTree.getExecutionRootFromArg(arg)
	if err != nil {
		return nil, err
	}

	// build tree of result groups, starting with a synthetic 'root' node
	executionTree.Root = NewRootResultGroup(ctx, executionTree, rootItem)

	// after tree has built, ControlCount will be set - create progress rendered
	executionTree.Progress = controlstatus.NewControlProgress(len(executionTree.ControlRuns))

	return executionTree, nil
}

// IsExportSourceData implements ExportSourceData
func (*ExecutionTree) IsExportSourceData() {}

// AddControl checks whether control should be included in the tree
// if so, creates a ControlRun, which is added to the parent group
func (e *ExecutionTree) AddControl(ctx context.Context, control *modconfig.Control, group *ResultGroup) {
	// note we use short name to determine whether to include a control
	if e.ShouldIncludeControl(control.ShortName) {
		// create new ControlRun with treeItem as the parent
		controlRun := NewControlRun(control, group, e)
		// add it into the group
		group.addControl(controlRun)

		// also add it into the execution tree control run list
		e.ControlRuns = append(e.ControlRuns, controlRun)
	}
}

func (e *ExecutionTree) Execute(ctx context.Context) int {
	log.Println("[TRACE]", "begin ExecutionTree.Execute")
	defer log.Println("[TRACE]", "end ExecutionTree.Execute")
	e.StartTime = time.Now()
	e.Progress.Start(ctx)

	defer func() {
		e.EndTime = time.Now()
		e.Progress.Finish(ctx)
	}()

	// the number of goroutines parallel to start
	var maxParallelGoRoutines int64 = constants.DefaultMaxConnections
	if viper.IsSet(constants.ArgMaxParallel) {
		maxParallelGoRoutines = viper.GetInt64(constants.ArgMaxParallel)
	}

	// to limit the number of parallel controls go routines started
	parallelismLock := semaphore.NewWeighted(maxParallelGoRoutines)

	// just execute the root - it will traverse the tree
	e.Root.execute(ctx, e.client, parallelismLock)

	if err := e.waitForActiveRunsToComplete(ctx, parallelismLock, maxParallelGoRoutines); err != nil {
		log.Printf("[WARN] timed out waiting for active runs to complete")
	}

	failures := e.Root.Summary.Status.Alarm + e.Root.Summary.Status.Error

	// now build map of dimension property name to property value to color map
	e.DimensionColorGenerator, _ = NewDimensionColorGenerator(4, 27)
	e.DimensionColorGenerator.populate(e)

	return failures
}

func (e *ExecutionTree) waitForActiveRunsToComplete(ctx context.Context, parallelismLock *semaphore.Weighted, maxParallelGoRoutines int64) error {
	waitCtx := ctx
	// if the context was already cancelled, we must creat ea new one to use  when waiting to acquire the lock
	if ctx.Err() != nil {
		// use a Background context - since the original context has been cancelled
		// this lets us wait for the active control queries to cancel
		c, cancel := context.WithTimeout(context.Background(), constants.ControlQueryCancellationTimeoutSecs*time.Second)
		waitCtx = c
		defer cancel()
	}
	// wait till we can acquire all semaphores - meaning that all active runs have finished
	return parallelismLock.Acquire(waitCtx, maxParallelGoRoutines)
}

func (e *ExecutionTree) populateControlFilterMap(ctx context.Context, controlFilterWhereClause string) error {
	// if we derived or were passed a where clause, run the filter
	if len(controlFilterWhereClause) > 0 {
		log.Println("[TRACE]", "filtering controls with", controlFilterWhereClause)
		var err error
		e.controlNameFilterMap, err = e.getControlMapFromWhereClause(ctx, controlFilterWhereClause)
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
func (e *ExecutionTree) getExecutionRootFromArg(arg string) (modconfig.ModTreeItem, error) {
	// special case handling for the string "all"
	if arg == "all" {
		// if the arg is "all", we want to execute all _direct_ children of the Mod
		// but NOT children which come from dependency mods

		// to achieve this, use a  DirectChildrenModDecorator

		return DirectChildrenModDecorator{e.Workspace.Mod}, nil
	}

	// if the arg is the name of one of the workspace dependendencies, wrap it in DirectChildrenModDecorator
	// so we only execute _its_ direct children
	for _, mod := range e.Workspace.Mods {
		if mod.ShortName == arg {
			return DirectChildrenModDecorator{mod}, nil
		}
	}

	// what resource type is arg?
	parsedName, err := modconfig.ParseResourceName(arg)
	if err != nil {
		// just log error
		return nil, fmt.Errorf("failed to parse check argument '%s': %v", arg, err)
	}

	resource, found := modconfig.GetResource(e.Workspace, parsedName)

	root, ok := resource.(modconfig.ModTreeItem)
	if !found || !ok {
		return nil, fmt.Errorf("no resources found matching argument '%s'", arg)
	}
	// root item must be either a benchmark or a control
	if !helpers.StringSliceContains([]string{modconfig.BlockTypeControl, modconfig.BlockTypeBenchmark}, root.BlockType()) {
		return nil, fmt.Errorf("cannot execute '%s' using check, only controls and benchmarks may be run", resource.Name())
	}
	return root, nil
}

// Get a map of control names from the introspection table steampipe_control
// This is used to implement the 'where' control filtering
func (e *ExecutionTree) getControlMapFromWhereClause(ctx context.Context, whereClause string) (map[string]bool, error) {
	// query may either be a 'where' clause, or a named query
	resolvedQuery, _, err := e.Workspace.ResolveQueryAndArgsFromSQLString(whereClause)
	if err != nil {
		return nil, err
	}
	// did we in fact resolve a named query, or just return the 'name' as the query
	isNamedQuery := resolvedQuery.ExecuteSQL != whereClause

	// if the query is NOT a named query, we need to construct a full query by adding a select
	if !isNamedQuery {
		resolvedQuery.ExecuteSQL = fmt.Sprintf("select resource_name from %s where %s", constants.IntrospectionTableControl, whereClause)
	}

	res, err := e.client.ExecuteSync(ctx, resolvedQuery.ExecuteSQL, resolvedQuery.Args...)
	if err != nil {
		return nil, err
	}

	//
	// find the "resource_name" column index
	resourceNameColumnIndex := -1

	for i, c := range res.Cols {
		if c.Name == "resource_name" {
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
	// map keep track which tags have been added as columns
	tagColumnMap := make(map[string]bool)
	var tagColumns []string
	for _, r := range e.ControlRuns {
		if r.Control.Tags != nil {
			for tag := range r.Control.Tags {
				if !tagColumnMap[tag] {
					tagColumns = append(tagColumns, tag)
					tagColumnMap[tag] = true
				}
			}
		}
	}
	sort.Strings(tagColumns)
	return tagColumns
}
