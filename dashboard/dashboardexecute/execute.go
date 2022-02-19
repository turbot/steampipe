package dashboardexecute

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/workspace"
)

// TODO [repprts] we probably need locking
// map of execution trees, keyed by dashboard name
var executions = make(map[string]*DashboardExecutionTree)

func ExecuteDashboard(ctx context.Context, dashboardName string, workspace *workspace.Workspace, client db_common.Client) error {
	// TODO SET INPUTS
	// TODO [reports] if this report is already running - cancel ??? fail???
	var executionTree *DashboardExecutionTree
	executionTree, found := executions[dashboardName]
	if found {
		// there is alread an execution tree - rerun
		if executionTree.GetRunStatus() == dashboardinterfaces.DashboardRunReady {
			return fmt.Errorf("dashboard %s is already running", dashboardName)
		}
	} else {
		// there is no execution tree - reset
		var err error
		executionTree, err = NewReportExecutionTree(dashboardName, client, workspace)
		if err != nil {
			return err
		}
		executions[dashboardName] = executionTree
	}

	// TODO [reports] for now leave execution in tree?
	// for now we leave the tree in the map

	go executionTree.Execute(ctx)

	return nil
}

func SetDashboardInputs(ctx context.Context, dashboardName string, inputs map[string]*string) error {
	// find the execution
	executionTree, found := executions[dashboardName]
	if !found {
		return fmt.Errorf("dashboard %s is not running", dashboardName)
	}

	// check the dashboard run status - if it is complete,  re-execute
	if executionTree.GetRunStatus() == dashboardinterfaces.DashboardRunComplete {
		go executionTree.Execute(ctx)
	}

	return executionTree.SetInputs(inputs)
}

func ResetDashboard(_ context.Context, dashboardName string) {
	// find the execution
	executionTree, found := executions[dashboardName]
	if !found {
		// nothing to do
		return
	}

	// cancel if in progress
	executionTree.Cancel()

	// remove from execution tree
	delete(executions, dashboardName)
}

func CancelDashboard(_ context.Context, dashboardName string) {
	// find the execution
	executionTree, found := executions[dashboardName]
	if !found {
		// nothing to do
		return
	}

	// cancel if in progress
	executionTree.Cancel()
}
