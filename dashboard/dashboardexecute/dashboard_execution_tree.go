package dashboardexecute

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
)

// DashboardExecutionTree is a structure representing the control result hierarchy
type DashboardExecutionTree struct {
	Root *DashboardRun

	dashboardName string
	sessionId     string
	client        db_common.Client
	runs          map[string]dashboardinterfaces.DashboardNodeRun
	workspace     *workspace.Workspace
	runComplete   chan dashboardinterfaces.DashboardNodeRun

	inputLock sync.Mutex
	// store subscribers as a map of maps for simple unsubscription
	inputDataSubscriptions map[string]map[*chan bool]struct{}
	cancel                 context.CancelFunc
	inputValues            map[string]interface{}
	id                     string
}

// NewDashboardExecutionTree creates a result group from a ModTreeItem
func NewDashboardExecutionTree(reportName string, sessionId string, client db_common.Client, workspace *workspace.Workspace) (*DashboardExecutionTree, error) {
	// now populate the DashboardExecutionTree
	executionTree := &DashboardExecutionTree{
		dashboardName:          reportName,
		sessionId:              sessionId,
		client:                 client,
		runs:                   make(map[string]dashboardinterfaces.DashboardNodeRun),
		workspace:              workspace,
		runComplete:            make(chan dashboardinterfaces.DashboardNodeRun, 1),
		inputDataSubscriptions: make(map[string]map[*chan bool]struct{}),
		inputValues:            make(map[string]interface{}),
	}
	executionTree.id = fmt.Sprintf("%p", executionTree)

	// create the root run node (either a report run or a counter run)
	root, err := executionTree.createRootItem(reportName)
	if err != nil {
		return nil, err
	}

	executionTree.Root = root
	return executionTree, nil
}

func (e *DashboardExecutionTree) createRootItem(reportName string) (*DashboardRun, error) {
	parsedName, err := modconfig.ParseResourceName(reportName)
	if err != nil {
		return nil, err
	}

	if parsedName.ItemType != modconfig.BlockTypeDashboard {
		return nil, fmt.Errorf("reporting type %s cannot be executed directly - only reports may be executed", parsedName.ItemType)
	}
	dashboard, ok := e.workspace.GetResourceMaps().Dashboards[reportName]
	if !ok {
		return nil, fmt.Errorf("report '%s' does not exist in workspace", reportName)
	}
	return NewDashboardRun(dashboard, e, e)

}

func (e *DashboardExecutionTree) Execute(ctx context.Context) {
	startTime := time.Now()

	// store context
	cancelCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	workspace := e.workspace
	workspace.PublishDashboardEvent(&dashboardevents.ExecutionStarted{
		Dashboard:   e.Root,
		Session:     e.sessionId,
		ExecutionId: e.id,
	})
	defer workspace.PublishDashboardEvent(&dashboardevents.ExecutionComplete{
		Dashboard:   e.Root,
		Session:     e.sessionId,
		ExecutionId: e.id,
		Inputs:      e.inputValues,
		Variables:   e.workspace.Variables,
		SearchPath:  e.client.GetRequiredSessionSearchPath(),
		StartTime:   startTime,
		EndTime:     time.Now(),
		Actor:       "",
	})

	log.Println("[TRACE]", "begin DashboardExecutionTree.Execute")
	defer log.Println("[TRACE]", "end DashboardExecutionTree.Execute")

	if e.GetRunStatus() == dashboardinterfaces.DashboardRunComplete {
		// there must be no nodes to execute
		log.Println("[TRACE]", "execution tree already complete")
		return
	}

	// execute synchronously
	e.Root.Execute(cancelCtx)
}

// GetRunStatus returns the stats of the Root run
func (e *DashboardExecutionTree) GetRunStatus() dashboardinterfaces.DashboardRunStatus {
	return e.Root.GetRunStatus()
}

// SetError sets the error on the Root run
func (e *DashboardExecutionTree) SetError(err error) {
	e.Root.SetError(err)
}

// GetName implements DashboardNodeParent
// use mod chort name - this will be the root name for all child runs
func (e *DashboardExecutionTree) GetName() string {
	return e.workspace.Mod.ShortName
}

func (e *DashboardExecutionTree) SetInputs(inputValues map[string]interface{}) error {
	for name, value := range inputValues {
		e.inputValues[name] = value
		// now see if anyone needs to be notified about this input
		e.notifyInputAvailable(name)
	}
	return nil
}

// ChildCompleteChan implements DashboardNodeParent
func (e *DashboardExecutionTree) ChildCompleteChan() chan dashboardinterfaces.DashboardNodeRun {
	return e.runComplete
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

func (e *DashboardExecutionTree) Cancel() {
	// if we have not completed, and already have a cancel function - cancel
	if e.GetRunStatus() != dashboardinterfaces.DashboardRunReady || e.cancel == nil {
		return
	}
	e.cancel()

	// if there are any children, wait for the execution to complete
	if len(e.Root.Children) > 0 {
		<-e.runComplete
	}
}

func (e *DashboardExecutionTree) GetInputValue(name string) interface{} {
	return e.inputValues[name]
}
