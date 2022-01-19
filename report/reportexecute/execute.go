package reportexecute

import (
	"context"
	"log"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/workspace"
)

func ExecuteReportNode(ctx context.Context, reportName string, workspace *workspace.Workspace, client db_common.Client) error {
	log.Printf("[WARN] **************** ExecuteReportNode ***************\n")
	// create context for the report execution
	// (for now just disable all status messages - replace with event based?	)
	reportCtx := statushooks.DisableStatusHooks(ctx)
	executionTree, err := NewReportExecutionTree(reportName, client, workspace)
	if err != nil {
		return err
	}

	go func() {
		workspace.PublishReportEvent(&reportevents.ExecutionStarted{ReportNode: executionTree.Root})

		if err := executionTree.Execute(reportCtx); err != nil {
			if executionTree.Root.GetRunStatus() != reportinterfaces.ReportRunError {
				// set error state on the root node
				executionTree.Root.SetError(err)
			}
		}
		workspace.PublishReportEvent(&reportevents.ExecutionComplete{Report: executionTree.Root})
	}()

	return nil
}
