package reportexecutiontree

import (
	"context"
	"fmt"
	"log"

	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportexecute"
	"github.com/turbot/steampipe/workspace"

	"github.com/stevenle/topsort"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ReportExecutionTree is a structure representing the control result hierarchy
type ReportExecutionTree struct {
	Root            *reportexecute.ReportRun
	dependencyGraph *topsort.Graph
	client          *db.Client
	panels          map[string]*reportexecute.PanelRun
	reports         map[string]*reportexecute.ReportRun
	workspace       *workspace.Workspace
}

// NewReportExecutionTree creates a result group from a ModTreeItem
func NewReportExecutionTree(root *modconfig.Report, client *db.Client, workspace *workspace.Workspace) (*ReportExecutionTree, error) {

	// now populate the ReportExecutionTree
	reportExecutionTree := &ReportExecutionTree{
		client:          client,
		dependencyGraph: topsort.NewGraph(),
		panels:          make(map[string]*reportexecute.PanelRun),
		reports:         make(map[string]*reportexecute.ReportRun),
		workspace:       workspace,
	}
	reportExecutionTree.Root = reportexecute.NewReportRun(root, reportExecutionTree)

	return reportExecutionTree, nil
}

func (e *ReportExecutionTree) Execute(ctx context.Context) error {
	log.Println("[TRACE]", "begin ReportExecutionTree.Execute")
	defer log.Println("[TRACE]", "end ReportExecutionTree.Execute")

	if e.runStatus() == reportexecute.ReportRunComplete {
		// there must be no sql panels to execute
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

func (e *ReportExecutionTree) runStatus() reportexecute.ReportRunStatus {
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
			// this error will be passed up the execution tree and raised as a report error for the root node
			return fmt.Errorf("report '%s' not found in execution tree", name)
		}
		// panel should now be complete, i.e. all it's children should be complete
		if !report.ChildrenComplete() {
			// this error will be passed up the execution tree and raised as a report error for the root node
			return fmt.Errorf("panel '%s' should be complete, but it has incomplete children", report.Name)
		}
		// set complete status on report - this will raise panel complete event
		report.SetComplete()
		return nil
	}

	if parsedName.ItemType == modconfig.BlockTypePanel {
		panel, ok := e.panels[name]
		if !ok {
			// this error will be passed up the execution tree and raised as a report error for the root node
			return fmt.Errorf("panel '%s' not found in execution tree", name)
		}
		// if panel has sql execute it
		if panel.SQL != "" {
			data, err := e.executePanelSQL(ctx, panel.SQL)
			if err != nil {
				// set the error status on the panel
				panel.SetError(err)
				// raise panel error event
				e.workspace.PublishReportEvent(&reportevents.PanelError{Panel: r})

				return err
			}

			panel.Data = data
		}
		// panel should now be complete, i.e. all it's children should be complete
		if !panel.ChildrenComplete() {
			// this error will be passed up the execution tree and raised as a report error for the root node
			return fmt.Errorf("panel '%s' should be complete, but it has inomplete children", panel.Name)
		}
		// set complete status on panel - this will raise panel complete event
		panel.SetComplete()

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
