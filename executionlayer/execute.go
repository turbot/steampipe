package executionlayer

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/report/reportexecutiontree"

	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/workspace"
)

func ExecuteReport(ctx context.Context, reportName string, workspace *workspace.Workspace, client *db.Client) error {
	report, ok := workspace.ReportMap[reportName]
	if !ok {
		return fmt.Errorf("report '%s' does not exist in workspace", reportName)
	}
	executionTree, err := reportexecutiontree.NewReportExecutionTree(report, client, workspace)
	if err != nil {
		return err
	}

	go func() {
		workspace.PublishReportEvent(&reportevents.ExecutionStarted{Report: executionTree.Root})

		if err := executionTree.Execute(ctx); err != nil {
			if executionTree.Root.Error == nil {
				executionTree.Root.SetError(err)
			}
		}
		workspace.PublishReportEvent(&reportevents.ExecutionComplete{Report: executionTree.Root})
	}()

	return nil
}
