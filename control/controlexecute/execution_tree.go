package controlexecute

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
	"golang.org/x/sync/semaphore"
)

// ExecutionTree is a structure representing the control execution hierarchy
type ExecutionTree struct {
	// root node of the execution tree
	Root      *ResultGroup
	StartTime time.Time
	EndTime   time.Time
	// map of dimension property name to property value to color map
	DimensionColorGenerator *DimensionColorGenerator
	// flat list of all control runs
	ControlRuns []*ControlRun

	workspace *workspace.Workspace
	client    db_common.Client
	// an optional map of control names used to filter the controls which are run
	controlNameFilterMap map[string]bool
	progress             *ControlProgressRenderer
}

func NewExecutionTree(ctx context.Context, workspace *workspace.Workspace, client db_common.Client, arg string) (*ExecutionTree, error) {
	// now populate the ExecutionTree
	executionTree := &ExecutionTree{
		workspace: workspace,
		client:    client,
	}
	// if a "--where" or "--tag" parameter was passed, build a map of control names used to filter the controls to run
	// create a context with status hooks disabled
	noStatusCtx := statushooks.DisableStatusHooks(ctx)
	err := executionTree.populateControlFilterMap(noStatusCtx)

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
	executionTree.progress = NewControlProgressRenderer(len(executionTree.ControlRuns))

	return executionTree, nil
}

// AddControl checks whether control should be included in the tree
// if so, creates a ControlRun, which is added to the parent group
func (e *ExecutionTree) AddControl(ctx context.Context, control *modconfig.Control, group *ResultGroup) {
	// note we use short name to determine whether to include a control
	if e.ShouldIncludeControl(control.ShortName) {
		// create new ControlRun with treeItem as the parent
		controlRun := NewControlRun(control, group, e)
		// add it into the group
		group.ControlRuns = append(group.ControlRuns, controlRun)
		// also add it into the execution tree control run list
		e.ControlRuns = append(e.ControlRuns, controlRun)
	}
}

func (e *ExecutionTree) Execute(ctx context.Context) int {
	log.Println("[TRACE]", "begin ExecutionTree.Execute")
	defer log.Println("[TRACE]", "end ExecutionTree.Execute")
	e.StartTime = time.Now()
	e.progress.Start(ctx)

	defer func() {
		e.EndTime = time.Now()
		e.progress.Finish(ctx)
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

	executeFinishWaitCtx := ctx
	if ctx.Err() != nil {
		// use a Background context - since the original context has been cancelled
		// this lets us wait for the active control queries to cancel
		c, cancel := context.WithTimeout(context.Background(), constants.ControlQueryCancellationTimeoutSecs*time.Second)
		executeFinishWaitCtx = c
		defer cancel()
	}
	// wait till we can acquire all semaphores - meaning that all active runs have finished
	parallelismLock.Acquire(executeFinishWaitCtx, maxParallelGoRoutines)

	failures := e.Root.Summary.Status.Alarm + e.Root.Summary.Status.Error

	// now build map of dimension property name to property value to color map
	e.DimensionColorGenerator, _ = NewDimensionColorGenerator(4, 27)
	e.DimensionColorGenerator.populate(e)

	return failures
}

func (e *ExecutionTree) populateControlFilterMap(ctx context.Context) error {
	// if both '--where' and '--tag' have been used, then it's an error
	if viper.IsSet(constants.ArgWhere) && viper.IsSet(constants.ArgTag) {
		return fmt.Errorf("'--%s' and '--%s' cannot be used together", constants.ArgWhere, constants.ArgTag)
	}

	controlFilterWhereClause := ""

	if viper.IsSet(constants.ArgTag) {
		// if '--tag' args were used, derive the whereClause from them
		tags := viper.GetStringSlice(constants.ArgTag)
		controlFilterWhereClause = e.generateWhereClauseFromTags(tags)
	} else if viper.IsSet(constants.ArgWhere) {
		// if a 'where' arg was used, execute this sql to get a list of  control names
		// use this list to build a name map used to determine whether to run a particular control
		controlFilterWhereClause = viper.GetString(constants.ArgWhere)
	}

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

func (e *ExecutionTree) generateWhereClauseFromTags(tags []string) string {
	whereMap := map[string][]string{}

	// 'tags' should be KV Pairs of the form: 'benchmark=pic' or 'cis_level=1'
	for _, tag := range tags {
		value, _ := url.ParseQuery(tag)
		for k, v := range value {
			if _, found := whereMap[k]; !found {
				whereMap[k] = []string{}
			}
			whereMap[k] = append(whereMap[k], v...)
		}
	}
	whereComponents := []string{}
	for key, values := range whereMap {
		thisComponent := []string{}
		for _, x := range values {
			if len(x) == 0 {
				// ignore
				continue
			}
			thisComponent = append(thisComponent, fmt.Sprintf("tags->>'%s'='%s'", key, x))
		}
		whereComponents = append(whereComponents, fmt.Sprintf("(%s)", strings.Join(thisComponent, " OR ")))
	}

	return strings.Join(whereComponents, " AND ")
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
		// return the workspace mod as root
		return e.workspace.Mod, nil
	}

	// if the arg is the name of one of the workjspace dependen
	for _, mod := range e.workspace.Mods {
		if mod.ShortName == arg {
			return mod, nil
		}
	}

	// what resource type is arg?
	parsedName, err := modconfig.ParseResourceName(arg)
	if err != nil {
		// just log error
		return nil, fmt.Errorf("failed to parse check argument '%s': %v", arg, err)
	}

	resource, found := modconfig.GetResource(e.workspace, parsedName)

	// if the resource is a ReportControl, get its underlying control
	if reportControl, ok := resource.(*modconfig.ReportControl); ok {
		resource = reportControl.GetControl()
	}

	root, ok := resource.(modconfig.ModTreeItem)
	if !found || !ok {
		return nil, fmt.Errorf("no resources found matching argument '%s'", arg)
	}
	return root, nil
}

// Get a map of control names from the introspection table steampipe_control
// This is used to implement the 'where' control filtering
func (e *ExecutionTree) getControlMapFromWhereClause(ctx context.Context, whereClause string) (map[string]bool, error) {
	// query may either be a 'where' clause, or a named query
	query, _, err := e.workspace.ResolveQueryAndArgs(whereClause)
	if err != nil {
		return nil, err
	}
	// did we in fact resolve a named query, or just return the 'name' as the query
	isNamedQuery := query != whereClause

	// if the query is NOT a named query, we need to construct a full query by adding a select
	if !isNamedQuery {
		query = fmt.Sprintf("select resource_name from %s where %s", constants.IntrospectionTableControl, whereClause)
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
	return tagColumns
}
