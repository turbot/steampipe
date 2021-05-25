package reportexecute

import (
	"context"

	"github.com/turbot/steampipe/workspace"
)

func ExecuteReport(ctx context.Context, reportName string, workspace *workspace.Workspace) {
	//// run all queries
	//failures := 0
	//for i, q := range queries {
	//	if err := executeQuery(ctx, q, client); err != nil {
	//		failures++
	//		utils.ShowWarning(fmt.Sprintf("executeQueries: query %d of %d failed: %v", i+1, len(queries), err))
	//	}
	//	if showBlankLineBetweenResults() {
	//		fmt.Println()
	//	}
	//}
	//
	//return failures
}
