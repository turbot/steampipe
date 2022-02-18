package dashboardexecute

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/workspace"
)

// map of execution trees, keyed by dashboard name
var executions = make(map[string]*DashboardExecutionTree)

func ExecuteDashboardNode(ctx context.Context, dashboardName string, workspace *workspace.Workspace, client db_common.Client) error {
	// TODO [reports] if this report is already running - cancel ??? fail???
	if _, running := executions[dashboardName]; running {
		return fmt.Errorf("dashboard %s is already running", dashboardName)
	}

	// TODO SET INPUTS

	executionTree, err := NewReportExecutionTree(dashboardName, client, workspace)
	if err != nil {
		return err
	}

	executions[dashboardName] = executionTree
	go func() {
		workspace.PublishDashboardEvent(&dashboardevents.ExecutionStarted{DashboardNode: executionTree.Root})
		defer func() {
			// remove tree from map of executions
			delete(executions, dashboardName)
			// send completed event
			workspace.PublishDashboardEvent(&dashboardevents.ExecutionComplete{Dashboard: executionTree.Root})
		}()

		if err := executionTree.Execute(ctx); err != nil {
			if executionTree.Root.GetRunStatus() != dashboardinterfaces.DashboardRunError {
				// set error state on the root node
				executionTree.Root.SetError(err)
			}
		}
	}()

	return nil
}

func SetDashboardInputs(_ context.Context, dashboardName string, inputs map[string]string) error {
	// find the execution
	executionTree, found := executions[dashboardName]
	if !found {
		return fmt.Errorf("dashboard %s is not running", dashboardName)
	}

	// TODO CHECK STATUS - if complete re-execute
	return executionTree.SetInputs(inputs)
}
