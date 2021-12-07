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
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
	"golang.org/x/sync/semaphore"
)

// ExecutionTree is a structure representing the control hierarchy
type ExecutionTree struct {
	Root      *ResultGroup
	StartTime time.Time
	EndTime   time.Time

	workspace *workspace.Workspace
	client    db_common.Client
	// an optional map of control names used to filter the controls which are run
	controlNameFilterMap map[string]bool
	progress             *ControlProgressRenderer
	// map of dimension property name to property value to color map
	DimensionColorGenerator *DimensionColorGenerator
	// flat list of all control runs
	controlRuns []*ControlRun
}

func NewExecutionTree(ctx context.Context, workspace *workspace.Workspace, client db_common.Client, arg string) (*ExecutionTree, error) {
	// now populate the ExecutionTree
	executionTree := &ExecutionTree{
		workspace: workspace,
		client:    client,
	}
	// if a "--where" or "--tag" parameter was passed, build a map of control manes used to filter the controls to run
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

func (e *ExecutionTree) Execute(ctx context.Context, client db_common.Client) int {
	log.Println("[TRACE]", "begin ExecutionTree.Execute")
	defer log.Println("[TRACE]", "end ExecutionTree.Execute")
	e.StartTime = time.Now()
	e.progress.Start()

	defer func() {
		e.EndTime = time.Now()
		e.progress.Finish()
	}()

	// the number of goroutines parallel to start
	// - we start goroutines as a multiplier of the number of parallel database connections
	// so that go routines receive connections as soon as they are available
	maxParallelGoRoutines := viper.GetInt64(constants.ArgMaxParallel) * constants.ParallelControlMultiplier

	// to limit the number of parallel controls go routines started
	parallelismLock := semaphore.NewWeighted(maxParallelGoRoutines)

	// just execute the root - it will traverse the tree
	e.Root.Execute(ctx, client, parallelismLock)

	executeFinishWaitCtx := ctx
	if ctx.Err() != nil {
		// use a Background context - since the original context has been cancelled
		// this lets us wait for the active control queries to cancel
		c, cancel := context.WithTimeout(context.Background(), constants.QueryCancellationTimeout*time.Second)
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
func (e *ExecutionTree) getExecutionRootFromArg(arg string) ([]modconfig.ModTreeItem, error) {
	var res []modconfig.ModTreeItem
	// special case handling for the string "all"
	if arg == "all" {
		//
		// build list of all workspace mods - these will act as root items
		for _, m := range e.workspace.Mods {
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
		if control, ok := e.workspace.Controls[arg]; ok {
			return []modconfig.ModTreeItem{control}, nil
		}
	case modconfig.BlockTypeBenchmark:
		// look in the workspace control group map for this control group
		if benchmark, ok := e.workspace.Benchmarks[arg]; ok {
			return []modconfig.ModTreeItem{benchmark}, nil
		}
	case modconfig.BlockTypeMod:
		// get all controls for the mod
		if mod, ok := e.workspace.Mods[arg]; ok {
			return []modconfig.ModTreeItem{mod}, nil
		}
	}
	return nil, fmt.Errorf("no controls found matching argument '%s'", arg)
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

	res, err := e.client.ExecuteSync(ctx, query, false)
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
	for _, r := range e.controlRuns {
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
