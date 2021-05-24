package reportexecute

import (
	"context"
	"log"

	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/workspace"
)

// ReportExecutionTree is a structure representing the control result hierarchy
type ReportExecutionTree struct {
	Root *PanelRun

	workspace *workspace.Workspace
	client    *db.Client
}

// NewReportExecutionTree creates a result group from a ControlTreeItem
func NewReportExecutionTree(ctx context.Context, workspace *workspace.Workspace, client *db.Client, arg string) (*ReportExecutionTree, error) {
	// now populate the ReportExecutionTree
	reportExecutionTree := &ReportExecutionTree{
		workspace: workspace,
		client:    client,
	}

	// build tree of result groups, starting with a synthetic 'root' node
	//reportExecutionTree.Root = NewRootResultGroup(reportExecutionTree, rootItems...)

	return reportExecutionTree, nil
}

func (e *ReportExecutionTree) Execute(ctx context.Context, client *db.Client) int {
	log.Println("[TRACE]", "begin ReportExecutionTree.Execute")
	defer log.Println("[TRACE]", "end ReportExecutionTree.Execute")

	return 0
}
