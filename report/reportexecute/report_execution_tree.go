package reportexecute

import (
	"context"
	"fmt"
	"log"

	"github.com/stevenle/topsort"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
)

// ReportExecutionTree is a structure representing the control result hierarchy
type ReportExecutionTree struct {
	Root            *ReportRun
	dependencyGraph *topsort.Graph

	workspace *workspace.Workspace
	client    *db.Client
	panels    map[string]*PanelRun
	reports   map[string]*ReportRun
}

// NewReportExecutionTree creates a result group from a ControlTreeItem
func NewReportExecutionTree(reportName string, workspace *workspace.Workspace, client *db.Client) (*ReportExecutionTree, error) {
	report, ok := workspace.ReportMap[reportName]
	if !ok {
		return nil, fmt.Errorf("report '%s' does not exist in workspace", reportName)
	}
	// now populate the ReportExecutionTree
	reportExecutionTree := &ReportExecutionTree{
		workspace:       workspace,
		client:          client,
		dependencyGraph: topsort.NewGraph(),
		panels:          make(map[string]*PanelRun),
		reports:         make(map[string]*ReportRun),
	}
	reportExecutionTree.Root = NewReportRun(report, reportExecutionTree)

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
	for _, name := range executionOrder {
		err = e.ExecuteNode(ctx, name)
		if err != nil {
			return err
		}
	}

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

func (e *ReportExecutionTree) ExecuteNode(ctx context.Context, name string) error {
	parsedName, err := modconfig.ParseResourceName(name)
	if err != nil {
		return err
	}

	if parsedName.ItemType == modconfig.BlockTypeReport {
		report, ok := e.reports[name]
		if !ok {
			return fmt.Errorf("report '%s' not found in execution tree", name)
		}
		// panel should now be complete, i.e. all it's children should be complete
		if !report.ChildrenComplete() {
			return fmt.Errorf("panel '%s' should be complete, but it has inomplete children", report.Name)
		}
		report.runStatus = ReportRunComplete
		return nil
	}

	if parsedName.ItemType == modconfig.BlockTypePanel {
		panel, ok := e.panels[name]
		if !ok {
			return fmt.Errorf("panel '%s' not found in execution tree", name)
		}
		// if panel has sql execute it
		if panel.SQL != "" {
			data, err := e.executePanelSQL(ctx, panel.SQL)
			if err != nil {
				return err
			}
			panel.Data = data
		}
		// panel should now be complete, i.e. all it's children should be complete
		if !panel.ChildrenComplete() {
			return fmt.Errorf("panel '%s' should be complete, but it has inomplete children", panel.Name)
		}
		panel.runStatus = ReportRunComplete
		return nil
	}
	return fmt.Errorf("invalid block type '%s' passed to ReportExecutionTree.ExecuteNode", name)
}

func (e *ReportExecutionTree) executePanelSQL(ctx context.Context, query string) ([][]interface{}, error) {
	queryResult, err := e.client.ExecuteSync(ctx, query)
	if err != nil {
		return nil, err
	}
	var res = make([][]interface{}, len(queryResult.Rows)+1)
	var columns = make([]interface{}, len(queryResult.ColTypes))
	for i, c := range queryResult.ColTypes {
		columns[i] = c.Name()
	}
	res[0] = columns
	for i, row := range queryResult.Rows {
		rowData := make([]interface{}, len(queryResult.ColTypes))
		for j, columnVal := range row.(*queryresult.RowResult).Data {
			rowData[j] = columnVal
		}
		res[i+1] = rowData
	}

	return res, nil
}
