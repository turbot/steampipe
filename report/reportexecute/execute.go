package reportexecute

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/workspace"
)

// map of execution trees, keyed by report name
var executions = make(map[string]*ReportExecutionTree)

func ExecuteReportNode(ctx context.Context, reportName string, workspace *workspace.Workspace, client db_common.Client) error {
	// TODO KAI if this report is already running - cancel ??? fail???
	if _, running := executions[reportName]; running {
		return fmt.Errorf("report %s is already running", reportName)
	}

	// create context for the report execution
	// (for now just disable all status messages - replace with event based?	)
	reportCtx := statushooks.DisableStatusHooks(ctx)
	executionTree, err := NewReportExecutionTree(reportName, client, workspace)
	if err != nil {
		return err
	}

	executions[reportName] = executionTree
	go func() {
		workspace.PublishReportEvent(&reportevents.ExecutionStarted{ReportNode: executionTree.Root})
		defer func() {
			// remove tree from map of executions
			delete(executions, reportName)
			// send completed event
			workspace.PublishReportEvent(&reportevents.ExecutionComplete{Report: executionTree.Root})
		}()

		if err := executionTree.Execute(reportCtx); err != nil {
			if executionTree.Root.GetRunStatus() != reportinterfaces.ReportRunError {
				// set error state on the root node
				executionTree.Root.SetError(err)
			}
		}

	}()

	return nil
}
