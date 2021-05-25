package reportexecute

import (
	"context"
	"fmt"
	"log"

	"github.com/stevenle/topsort"

	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/workspace"
)

// ReportExecutionTree is a structure representing the control result hierarchy
type ReportExecutionTree struct {
	Root            *ReportRun
	dependencyGraph *topsort.Graph

	workspace *workspace.Workspace
	client    *db.Client
}

// NewReportExecutionTree creates a result group from a ControlTreeItem
func NewReportExecutionTree(ctx context.Context, reportName string, workspace *workspace.Workspace, client *db.Client) (*ReportExecutionTree, error) {
	_, ok := workspace.ReportMap[reportName]
	if !ok {
		return nil, fmt.Errorf("report '%s' does not exist in workspace", reportName)
	}
	// now populate the ReportExecutionTree
	reportExecutionTree := &ReportExecutionTree{
		workspace:       workspace,
		client:          client,
		dependencyGraph: topsort.NewGraph(),
	}

	// build tree of result groups, starting with a synthetic 'root' node
	//reportExecutionTree.Root = NewRootResultGroup(reportExecutionTree, rootItems...)

	return reportExecutionTree, nil
}

func (e *ReportExecutionTree) Execute(ctx context.Context) error {
	log.Println("[TRACE]", "begin ReportExecutionTree.Execute")
	defer log.Println("[TRACE]", "end ReportExecutionTree.Execute")

	if e.runStatus() == ReportRunComplete {
		log.Println("[TRACE]", "execution tree already complete")
		return nil
	}
	//get the dependency order
	executionOrder, err := e.dependencyGraph.TopSort(e.Root.Name)
	if err != nil {
		return err
	}
	fmt.Println(executionOrder)

	return nil
}

// AddDependency adds a dependency relationship to our dependency graph
// the resource has a dependency on an incomplete child resource
func (e *ReportExecutionTree) AddDependency(resource, dependency string) {
	if !e.dependencyGraph.ContainsNode(resource) {
		e.dependencyGraph.AddNode(resource)
	}
	if !e.dependencyGraph.ContainsNode(dependency) {
		e.dependencyGraph.AddNode(dependency)
	}
	// add root dependency
	e.dependencyGraph.AddEdge(resource, dependency)
}

func (e *ReportExecutionTree) runStatus() ReportRunStatus {
	return e.Root.runStatus
}
