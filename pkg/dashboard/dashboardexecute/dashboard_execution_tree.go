package dashboardexecute

import (
	"context"
	"fmt"
	"golang.org/x/exp/maps"
	"log"
	"sync"
	"time"

	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/workspace"
)

// DashboardExecutionTree is a structure representing the control result hierarchy
type DashboardExecutionTree struct {
	Root dashboardtypes.DashboardNodeRun

	dashboardName string
	sessionId     string
	client        db_common.Client
	runs          map[string]dashboardtypes.DashboardNodeRun
	workspace     *workspace.Workspace
	runComplete   chan dashboardtypes.DashboardNodeRun

	inputLock sync.Mutex
	// store subscribers as a map of maps for simple unsubscription
	inputDataSubscriptions map[string]map[*chan bool]struct{}
	cancel                 context.CancelFunc
	inputValues            map[string]interface{}
	id                     string
}

func NewDashboardExecutionTree(rootName string, sessionId string, client db_common.Client, workspace *workspace.Workspace) (*DashboardExecutionTree, error) {
	// now populate the DashboardExecutionTree
	executionTree := &DashboardExecutionTree{
		dashboardName:          rootName,
		sessionId:              sessionId,
		client:                 client,
		runs:                   make(map[string]dashboardtypes.DashboardNodeRun),
		workspace:              workspace,
		runComplete:            make(chan dashboardtypes.DashboardNodeRun, 1),
		inputDataSubscriptions: make(map[string]map[*chan bool]struct{}),
		inputValues:            make(map[string]interface{}),
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

func (e *DashboardExecutionTree) createRootItem(rootName string) (dashboardtypes.DashboardNodeRun, error) {
	parsedName, err := modconfig.ParseResourceName(rootName)
	if err != nil {
		return nil, err
	}
	if parsedName.ItemType == "" {
		return nil, fmt.Errorf("root item is not valid named resource")
	}
	// if no mod is specified, assume the workspace mod
	if parsedName.Mod == "" {
		parsedName.Mod = e.workspace.Mod.ShortName
		rootName = parsedName.ToFullName()
	}
	switch parsedName.ItemType {
	case modconfig.BlockTypeDashboard:
		dashboard, ok := e.workspace.GetResourceMaps().Dashboards[rootName]
		if !ok {
			return nil, fmt.Errorf("dashboard '%s' does not exist in workspace", rootName)
		}
		return NewDashboardRun(dashboard, e, e)
	case modconfig.BlockTypeBenchmark:
		benchmark, ok := e.workspace.GetResourceMaps().Benchmarks[rootName]
		if !ok {
			return nil, fmt.Errorf("benchmark '%s' does not exist in workspace", rootName)
		}
		return NewCheckRun(benchmark, e, e)
	case modconfig.BlockTypeQuery:
		// wrap in a table
		query, ok := e.workspace.GetResourceMaps().Queries[rootName]
		if !ok {
			return nil, fmt.Errorf("query '%s' does not exist in workspace", rootName)
		}
		// wrap this in a chart and a dashboard
		dashboard, err := modconfig.NewQueryDashboard(query)
		if err != nil {
			return nil, err
		}
		return NewDashboardRun(dashboard, e, e)
	case modconfig.BlockTypeControl:
		// wrap in a table
		control, ok := e.workspace.GetResourceMaps().Controls[rootName]
		if !ok {
			return nil, fmt.Errorf("query '%s' does not exist in workspace", rootName)
		}
		// wrap this in a chart and a dashboard
		dashboard, err := modconfig.NewQueryDashboard(control)
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

	// store context
	cancelCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	workspace := e.workspace

	// perform any necessary initialisation
	// (e.g. check run creates the control execution tree)
	e.Root.Initialise(ctx)
	if e.Root.GetError() != nil {
		return
	}

	panels := e.BuildSnapshotPanels()
	workspace.PublishDashboardEvent(&dashboardevents.ExecutionStarted{
		Root:        e.Root,
		Session:     e.sessionId,
		ExecutionId: e.id,
		Panels:      panels,
	})
	defer func() {
		// build map of those variables referenced by the dashboard run
		referencedVariables := GetReferencedVariables(e.Root, e.workspace)
		e := &dashboardevents.ExecutionComplete{
			Root:        e.Root,
			Session:     e.sessionId,
			ExecutionId: e.id,
			Panels:      panels,
			Inputs:      e.inputValues,
			Variables:   referencedVariables,
			// search path elements are quoted (for consumption by postgres)
			// unquote them
			SearchPath: utils.UnquoteStringArray(e.client.GetRequiredSessionSearchPath()),
			StartTime:  startTime,
			EndTime:    time.Now(),
		}
		workspace.PublishDashboardEvent(e)
	}()

	log.Println("[TRACE]", "begin DashboardExecutionTree.Execute")
	defer log.Println("[TRACE]", "end DashboardExecutionTree.Execute")

	if e.GetRunStatus() == dashboardtypes.DashboardRunComplete {
		// there must be no nodes to execute
		log.Println("[TRACE]", "execution tree already complete")
		return
	}

	// execute synchronously
	e.Root.Execute(cancelCtx)
}

// GetRunStatus returns the stats of the Root run
func (e *DashboardExecutionTree) GetRunStatus() dashboardtypes.DashboardRunStatus {
	return e.Root.GetRunStatus()
}

// SetError sets the error on the Root run
func (e *DashboardExecutionTree) SetError(ctx context.Context, err error) {
	e.Root.SetError(ctx, err)
}

// GetName implements DashboardNodeParent
// use mod chort name - this will be the root name for all child runs
func (e *DashboardExecutionTree) GetName() string {
	return e.workspace.Mod.ShortName
}

func (e *DashboardExecutionTree) SetInputs(inputValues map[string]interface{}) {
	for name, value := range inputValues {
		e.inputValues[name] = value
		// now see if anyone needs to be notified about this input
		e.notifyInputAvailable(name)
	}
}

// ChildCompleteChan implements DashboardNodeParent
func (e *DashboardExecutionTree) ChildCompleteChan() chan dashboardtypes.DashboardNodeRun {
	return e.runComplete
}

func (e *DashboardExecutionTree) Cancel() {
	// if we have not completed, and already have a cancel function - cancel
	if e.GetRunStatus() != dashboardtypes.DashboardRunReady || e.cancel == nil {
		return
	}
	e.cancel()

	// if there are any children, wait for the execution to complete
	if !e.Root.RunComplete() {
		<-e.runComplete
	}
}

func (e *DashboardExecutionTree) GetInputValue(name string) interface{} {
	return e.inputValues[name]
}

func (e *DashboardExecutionTree) BuildSnapshotPanels() map[string]dashboardtypes.SnapshotPanel {
	res := map[string]dashboardtypes.SnapshotPanel{}
	// if this node is a snapshot node, add to map
	if snapshotNode, ok := e.Root.(dashboardtypes.SnapshotPanel); ok {
		res[e.Root.GetName()] = snapshotNode
	}
	return e.buildSnapshotPanelsUnder(e.Root, res)
}

// RuntimeDependencies returns the runtime depedencies for all leaf nodes
func (e *DashboardExecutionTree) RuntimeDependencies() []string {
	var deps = map[string]struct{}{}
	for _, r := range e.runs {
		if leafRun, ok := r.(*LeafRun); ok {
			for _, v := range leafRun.runtimeDependencies {
				deps[v.dependency.SourceResource.GetUnqualifiedName()] = struct{}{}
			}
		}
	}
	return maps.Keys(deps)
}

func (e *DashboardExecutionTree) waitForRuntimeDependency(ctx context.Context, dependency *modconfig.RuntimeDependency) error {
	depChan := make(chan bool, 1)

	e.subscribeToInput(dependency.SourceResource.GetUnqualifiedName(), &depChan)
	defer e.unsubscribeToInput(dependency.SourceResource.GetUnqualifiedName(), &depChan)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-depChan:
		return nil
	}
}

func (e *DashboardExecutionTree) subscribeToInput(inputName string, depChan *chan bool) {
	e.inputLock.Lock()
	defer e.inputLock.Unlock()
	subscriptions := e.inputDataSubscriptions[inputName]
	if subscriptions == nil {
		subscriptions = make(map[*chan bool]struct{})
	}
	subscriptions[depChan] = struct{}{}
	e.inputDataSubscriptions[inputName] = subscriptions
}

func (e *DashboardExecutionTree) notifyInputAvailable(inputName string) {
	e.inputLock.Lock()
	defer e.inputLock.Unlock()

	for c := range e.inputDataSubscriptions[inputName] {
		*c <- true
	}
}

// remove a subscriber from the map of subscribers for this input
func (e *DashboardExecutionTree) unsubscribeToInput(inputName string, depChan *chan bool) {
	e.inputLock.Lock()
	defer e.inputLock.Unlock()

	subscribers := e.inputDataSubscriptions[inputName]
	if len(subscribers) == 0 {
		return
	}
	delete(subscribers, depChan)
}

func (e *DashboardExecutionTree) buildSnapshotPanelsUnder(parent dashboardtypes.DashboardNodeRun, res map[string]dashboardtypes.SnapshotPanel) map[string]dashboardtypes.SnapshotPanel {
	if checkRun, ok := parent.(*CheckRun); ok {
		return checkRun.BuildSnapshotPanels(res)
	}
	for _, c := range parent.GetChildren() {
		// if this node is a snapshot node, add to map
		if snapshotNode, ok := c.(dashboardtypes.SnapshotPanel); ok {
			res[c.GetName()] = snapshotNode
		}
		res = e.buildSnapshotPanelsUnder(c, res)
	}

	return res
}
