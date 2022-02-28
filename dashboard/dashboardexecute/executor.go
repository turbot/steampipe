package dashboardexecute

import (
	"context"
	"fmt"
	"sync"

	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/workspace"
)

type DashboardExecutor struct {
	executions    map[string]*DashboardExecutionTree
	executionLock sync.Mutex
}

func newDashboardExecutor() *DashboardExecutor {
	return &DashboardExecutor{
		executions: make(map[string]*DashboardExecutionTree),
	}
}

var Executor = newDashboardExecutor()

func (e *DashboardExecutor) ExecuteDashboard(ctx context.Context, sessionId, dashboardName string, inputs map[string]interface{}, workspace *workspace.Workspace, client db_common.Client) error {
	// reset any existing executions for this session
	e.ClearDashboard(ctx, sessionId)

	// now create a new execution
	executionTree, err := NewDashboardExecutionTree(dashboardName, sessionId, client, workspace)
	if err != nil {
		return err
	}
	// add to execution map
	e.setExecution(sessionId, executionTree)

	// if inputs have been passed, set them first
	if len(inputs) > 0 {
		if err := executionTree.SetInputs(inputs); err != nil {
			return err
		}
	}

	go executionTree.Execute(ctx)

	return nil
}

func (e *DashboardExecutor) SetDashboardInputs(ctx context.Context, sessionId string, inputs map[string]interface{}, changedInput string) error {
	// find the execution
	executionTree, found := e.executions[sessionId]
	if !found {
		return fmt.Errorf("no dashboard running for session %s", sessionId)
	}

	// first see if any other inputs rely on the one which was just changed
	clearedInputs := e.clearDependentInputs(executionTree.Root, changedInput, inputs)
	if len(clearedInputs) > 0 {
		event := &dashboardevents.InputValuesCleared{
			ClearedInputs: clearedInputs,
			Session:       executionTree.sessionId,
		}
		executionTree.workspace.PublishDashboardEvent(event)
	}
	// oif there are any dependent inputs, set their value to nil and send an event to the UI
	// if the dashboard run is complete, just re-execute
	if executionTree.GetRunStatus() == dashboardinterfaces.DashboardRunComplete {
		return e.ExecuteDashboard(
			ctx,
			sessionId,
			executionTree.dashboardName,
			inputs,
			executionTree.workspace,
			executionTree.client)
	}

	// set the inputs
	if err := executionTree.SetInputs(inputs); err != nil {
		return err
	}

	return nil
}

func (e *DashboardExecutor) clearDependentInputs(dashboard *DashboardRun, changedInput string, inputs map[string]interface{}) []string {
	dependentInputs := dashboard.GetInputsDependingOn(changedInput)
	clearedInputs := dependentInputs
	if len(dependentInputs) > 0 {
		for _, inputName := range dependentInputs {
			// clear the input value
			inputs[inputName] = nil
			childDependentInputs := e.clearDependentInputs(dashboard, inputName, inputs)
			clearedInputs = append(clearedInputs, childDependentInputs...)
		}
	}

	return clearedInputs
}

func (e *DashboardExecutor) ClearDashboard(_ context.Context, sessionId string) {
	// find the execution
	executionTree, found := e.getExecution(sessionId)
	if !found {
		// nothing to do
		return
	}

	// cancel if in progress
	executionTree.Cancel()

	// remove from execution tree
	e.removeExecution(sessionId)
}

func (e *DashboardExecutor) getExecution(sessionId string) (*DashboardExecutionTree, bool) {
	e.executionLock.Lock()
	defer e.executionLock.Unlock()

	executionTree, found := e.executions[sessionId]
	return executionTree, found
}

func (e *DashboardExecutor) setExecution(sessionId string, executionTree *DashboardExecutionTree) {
	e.executionLock.Lock()
	defer e.executionLock.Unlock()

	e.executions[sessionId] = executionTree
}

func (e *DashboardExecutor) removeExecution(sessionId string) {
	e.executionLock.Lock()
	defer e.executionLock.Unlock()

	delete(e.executions, sessionId)
}
