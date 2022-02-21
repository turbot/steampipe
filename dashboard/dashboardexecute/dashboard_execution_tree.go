package dashboardexecute

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
)

// DashboardExecutionTree is a structure representing the control result hierarchy
type DashboardExecutionTree struct {
	modconfig.UniqueNameProviderBase

	Root *DashboardRun

	dashboardName string
	sessionId     string
	client        db_common.Client
	runs          map[string]dashboardinterfaces.DashboardNodeRun
	workspace     *workspace.Workspace
	runComplete   chan dashboardinterfaces.DashboardNodeRun

	inputLock              sync.Mutex
	inputDataSubscriptions map[string][]chan bool
	cancel                 context.CancelFunc
}

// NewReportExecutionTree creates a result group from a ModTreeItem
func NewReportExecutionTree(reportName string, sessionId string, client db_common.Client, workspace *workspace.Workspace) (*DashboardExecutionTree, error) {
	// now populate the DashboardExecutionTree
	reportExecutionTree := &DashboardExecutionTree{
		dashboardName:          reportName,
		sessionId:              sessionId,
		client:                 client,
		runs:                   make(map[string]dashboardinterfaces.DashboardNodeRun),
		workspace:              workspace,
		runComplete:            make(chan dashboardinterfaces.DashboardNodeRun, 1),
		inputDataSubscriptions: make(map[string][]chan bool),
	}

	// create the root run node (either a report run or a counter run)
	root, err := reportExecutionTree.createRootItem(reportName)
	if err != nil {
		return nil, err
	}

	reportExecutionTree.Root = root
	return reportExecutionTree, nil
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

func (e *DashboardExecutionTree) Execute(ctx context.Context) error {
	// store context
	cancelCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	workspace := e.workspace
	workspace.PublishDashboardEvent(&dashboardevents.ExecutionStarted{
		DashboardNode: e.Root,
		Session:       e.sessionId,
	})
	defer workspace.PublishDashboardEvent(&dashboardevents.ExecutionComplete{
		Dashboard: e.Root,
		Session:   e.sessionId,
	})

	log.Println("[TRACE]", "begin DashboardExecutionTree.Execute")
	defer log.Println("[TRACE]", "end DashboardExecutionTree.Execute")

	if e.GetRunStatus() == dashboardinterfaces.DashboardRunComplete {
		// there must be no nodes to execute
		log.Println("[TRACE]", "execution tree already complete")
		return nil
	}

	err := e.Root.Execute(cancelCtx)
	if err != nil {
		// if the tree is not already in error state, set the error
		// (this is to avoid overwriting a previous error)
		if e.GetRunStatus() != dashboardinterfaces.DashboardRunError {
			// set error state on the root node
			e.SetError(err)
		}
	}
	return err
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

func (e *DashboardExecutionTree) SetInputs(inputValues map[string]*string) error {
	for name, value := range inputValues {
		// first find a matching input
		input, found := e.Root.GetInput(name)
		if !found {
			return fmt.Errorf("no input found matchint '%s'", name)
		}
		// set the input value
		input.SetValue(value)
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

	e.subscribeToInput(dependency.SourceResource.GetUnqualifiedName(), depChan)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-depChan:
		return nil
	}
}

func (e *DashboardExecutionTree) subscribeToInput(inputName string, depChan chan bool) {
	e.inputLock.Lock()
	defer e.inputLock.Unlock()

	e.inputDataSubscriptions[inputName] = append(e.inputDataSubscriptions[inputName], depChan)
}

func (e *DashboardExecutionTree) notifyInputAvailable(inputName string) {
	e.inputLock.Lock()
	defer e.inputLock.Unlock()

	for _, c := range e.inputDataSubscriptions[inputName] {
		c <- true
	}
}

func (e *DashboardExecutionTree) Cancel() {
	// if we have not completed, and already have a cancel function - cancel
	if e.GetRunStatus() != dashboardinterfaces.DashboardRunReady || e.cancel == nil {
		return
	}
	e.cancel()
}
