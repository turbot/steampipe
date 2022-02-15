package dashboardexecute

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/workspace"
)

// map of execution trees, keyed by dashboard name
var executions = make(map[string]*DashboardExecutionTree)

func ExecuteDashboardNode(ctx context.Context, dashboardName string, workspace *workspace.Workspace, client db_common.Client) error {
	// TODO [reports] if this report is already running - cancel ??? fail???
	if _, running := executions[dashboardName]; running {
		return fmt.Errorf("dashboard %s is already running", dashboardName)
	}

	// create context for the dashboard execution
	// (for now just disable all status messages - replace with event based?	)
	dashboardCtx := statushooks.DisableStatusHooks(ctx)
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

		if err := executionTree.Execute(dashboardCtx); err != nil {
			if executionTree.Root.GetRunStatus() != dashboardinterfaces.DashboardRunError {
				// set error state on the root node
				executionTree.Root.SetError(err)
			}
		}

	}()

	return nil
}
