package snapshot

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/turbot/steampipe/pkg/connection_sync"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/workspace"
	"golang.org/x/exp/maps"
)

// DashboardExecutionTree is a structure representing the control result hierarchy
type DashboardExecutionTree struct {
	Root dashboardtypes.DashboardTreeRun

	dashboardName string
	sessionId     string
	client        db_common.Client
	// map of executing runs, keyed by full name
	runs        map[string]dashboardtypes.DashboardTreeRun
	workspace   *workspace.Workspace
	runComplete chan dashboardtypes.DashboardTreeRun

	// map of subscribers to notify when an input value changes
	cancel      context.CancelFunc
	inputLock   sync.Mutex
	inputValues map[string]any
	id          string
}

func NewDashboardExecutionTree(rootName string, sessionId string, client db_common.Client, workspace *workspace.Workspace) (*DashboardExecutionTree, error) {
	// now populate the DashboardExecutionTree
	executionTree := &DashboardExecutionTree{
		dashboardName: rootName,
		sessionId:     sessionId,
		client:        client,
		runs:          make(map[string]dashboardtypes.DashboardTreeRun),
		workspace:     workspace,
		runComplete:   make(chan dashboardtypes.DashboardTreeRun, 1),
		inputValues:   make(map[string]any),
	}
	executionTree.id = fmt.Sprintf("%p", executionTree)

	// create the root run node (either a report run or a counter run)
	root, err := executionTree.createRootItem(rootName)
	if err != nil {
		return nil, err
	}

	executionTree.Root = root
	return executionTree, nil
}

func (e *DashboardExecutionTree) createRootItem(rootName string) (dashboardtypes.DashboardTreeRun, error) {
	parsedName, err := modconfig.ParseResourceName(rootName)
	if err != nil {
		return nil, err
	}
	fullName, err := parsedName.ToFullName()
	if err != nil {
		return nil, err
	}
	if parsedName.ItemType == "" {
		return nil, fmt.Errorf("root item is not valid named resource")
	}
	// if no mod is specified, assume the workspace mod
	if parsedName.Mod == "" {
		parsedName.Mod = e.workspace.Mod.ShortName
		rootName = fullName
	}
	switch parsedName.ItemType {
	case modconfig.BlockTypeQuery:
		// wrap in a table
		query, ok := e.workspace.GetResourceMaps().Queries[rootName]
		if !ok {
			return nil, fmt.Errorf("query '%s' does not exist in workspace", rootName)
		}
		// wrap this in a chart and a dashboard
		dashboard, err := modconfig.NewQueryDashboard(query)
		// TACTICAL - set the execution tree dashboard name from the query dashboard
		e.dashboardName = dashboard.Name()
		if err != nil {
			return nil, err
		}
		return NewDashboardRun(dashboard, e, e)
	default:
		return nil, fmt.Errorf("reporting type %s cannot be executed as dashboard", parsedName.ItemType)
	}
}

func (e *DashboardExecutionTree) Execute(ctx context.Context) {
	startTime := time.Now()

	searchPath := e.client.GetRequiredSessionSearchPath()

	// store context
	cancelCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	workspace := e.workspace

	// perform any necessary initialisation
	// (e.g. check run creates the control execution tree)
	e.Root.Initialise(cancelCtx)
	if e.Root.GetError() != nil {
		return
	}

	// TODO should we always wait even with non custom search path?
	// if there is a custom search path, wait until the first connection of each plugin has loaded
	if customSearchPath := e.client.GetCustomSearchPath(); customSearchPath != nil {
		if err := connection_sync.WaitForSearchPathSchemas(ctx, e.client, customSearchPath); err != nil {
			e.Root.SetError(ctx, err)
			return
		}
	}

	panels := e.BuildSnapshotPanels()
	// build map of those variables referenced by the dashboard run
	referencedVariables := GetReferencedVariables(e.Root, e.workspace)

	immutablePanels, err := utils.JsonCloneToMap(panels)
	if err != nil {
		e.SetError(ctx, err)
		return
	}
	workspace.PublishDashboardEvent(ctx, &dashboardevents.ExecutionStarted{
		Root:        e.Root,
		Session:     e.sessionId,
		ExecutionId: e.id,
		Panels:      immutablePanels,
		Inputs:      e.inputValues,
		Variables:   referencedVariables,
		StartTime:   startTime,
	})
	defer func() {

		e := &dashboardevents.ExecutionComplete{
			Root:        e.Root,
			Session:     e.sessionId,
			ExecutionId: e.id,
			Panels:      panels,
			Inputs:      e.inputValues,
			Variables:   referencedVariables,
			// search path elements are quoted (for consumption by postgres)
			// unquote them
			SearchPath: utils.UnquoteStringArray(searchPath),
			StartTime:  startTime,
			EndTime:    time.Now(),
		}
		workspace.PublishDashboardEvent(ctx, e)
	}()

	log.Println("[TRACE]", "begin DashboardExecutionTree.Execute")
	defer log.Println("[TRACE]", "end DashboardExecutionTree.Execute")

	if e.GetRunStatus().IsFinished() {
		// there must be no nodes to execute
		log.Println("[TRACE]", "execution tree already complete")
		return
	}

	// execute synchronously
	e.Root.Execute(cancelCtx)
}

// GetRunStatus returns the stats of the Root run
func (e *DashboardExecutionTree) GetRunStatus() dashboardtypes.RunStatus {
	return e.Root.GetRunStatus()
}

// SetError sets the error on the Root run
func (e *DashboardExecutionTree) SetError(ctx context.Context, err error) {
	e.Root.SetError(ctx, err)
}

// GetName implements DashboardParent
// use mod short name - this will be the root name for all child runs
func (e *DashboardExecutionTree) GetName() string {
	return e.workspace.Mod.ShortName
}

// GetParent implements DashboardTreeRun
func (e *DashboardExecutionTree) GetParent() dashboardtypes.DashboardParent {
	return nil
}

// GetNodeType implements DashboardTreeRun
func (*DashboardExecutionTree) GetNodeType() string {
	panic("should never call for DashboardExecutionTree")
}

func (e *DashboardExecutionTree) SetInputValues(inputValues map[string]any) {
	log.Printf("[TRACE] SetInputValues")
	e.inputLock.Lock()
	defer e.inputLock.Unlock()

	// we only support inputs if root is a dashboard (NOT a benchmark)
	runtimeDependencyPublisher, ok := e.Root.(RuntimeDependencyPublisher)
	if !ok {
		// should never happen
		log.Printf("[WARN] SetInputValues called but root Dashboard run is not a RuntimeDependencyPublisher: %s", e.Root.GetName())
		return
	}

	for name, value := range inputValues {
		log.Printf("[TRACE] DashboardExecutionTree SetInput %s = %v", name, value)
		e.inputValues[name] = value
		// publish runtime dependency
		runtimeDependencyPublisher.PublishRuntimeDependencyValue(name, &dashboardtypes.ResolvedRuntimeDependencyValue{Value: value})
	}
}

// ChildCompleteChan implements DashboardParent
func (e *DashboardExecutionTree) ChildCompleteChan() chan dashboardtypes.DashboardTreeRun {
	return e.runComplete
}

// ChildStatusChanged implements DashboardParent
func (*DashboardExecutionTree) ChildStatusChanged(context.Context) {}

func (e *DashboardExecutionTree) Cancel() {
	// if we have not completed, and already have a cancel function - cancel
	if e.GetRunStatus().IsFinished() || e.cancel == nil {
		log.Printf("[TRACE] DashboardExecutionTree Cancel NOT cancelling status %s cancel func %p", e.GetRunStatus(), e.cancel)
		return
	}

	log.Printf("[TRACE] DashboardExecutionTree Cancel  - calling cancel")
	e.cancel()

	// if there are any children, wait for the execution to complete
	if !e.Root.RunComplete() {
		<-e.runComplete
	}

	log.Printf("[TRACE] DashboardExecutionTree Cancel - all children complete")
}

func (e *DashboardExecutionTree) BuildSnapshotPanels() map[string]dashboardtypes.SnapshotPanel {
	// just build from e.runs
	res := map[string]dashboardtypes.SnapshotPanel{}

	for name, run := range e.runs {
		res[name] = run.(dashboardtypes.SnapshotPanel)
	}
	return res
}

// InputRuntimeDependencies returns the names of all inputs which are runtime dependencies
func (e *DashboardExecutionTree) InputRuntimeDependencies() []string {
	var deps = map[string]struct{}{}
	for _, r := range e.runs {
		if leafRun, ok := r.(*LeafRun); ok {
			for _, r := range leafRun.runtimeDependencies {
				if r.Dependency.PropertyPath.ItemType == modconfig.BlockTypeInput {
					deps[r.Dependency.SourceResourceName()] = struct{}{}
				}
			}
		}
	}
	return maps.Keys(deps)
}

// GetChildren implements DashboardParent
func (e *DashboardExecutionTree) GetChildren() []dashboardtypes.DashboardTreeRun {
	return []dashboardtypes.DashboardTreeRun{e.Root}
}

// ChildrenComplete implements DashboardParent
func (e *DashboardExecutionTree) ChildrenComplete() bool {
	return e.Root.RunComplete()
}

// Tactical: Empty implementations of DashboardParent functions
// TODO remove need for this

func (e *DashboardExecutionTree) Initialise(ctx context.Context) {
	panic("should never call for DashboardExecutionTree")
}

func (e *DashboardExecutionTree) GetTitle() string {
	panic("should never call for DashboardExecutionTree")
}

func (e *DashboardExecutionTree) GetError() error {
	panic("should never call for DashboardExecutionTree")
}

func (e *DashboardExecutionTree) SetComplete(ctx context.Context) {
	panic("should never call for DashboardExecutionTree")
}

func (e *DashboardExecutionTree) RunComplete() bool {
	panic("should never call for DashboardExecutionTree")
}

func (e *DashboardExecutionTree) GetInputsDependingOn(s string) []string {
	panic("should never call for DashboardExecutionTree")
}

func (*DashboardExecutionTree) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	panic("should never call for DashboardExecutionTree")
}

func (*DashboardExecutionTree) GetResource() modconfig.DashboardLeafNode {
	panic("should never call for DashboardExecutionTree")
}
