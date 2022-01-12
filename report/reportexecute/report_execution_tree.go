package reportexecute

import (
	"context"
	"fmt"
	"log"

	"github.com/stevenle/topsort"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
)

// ReportExecutionTree is a structure representing the control result hierarchy
type ReportExecutionTree struct {
	Root            reportinterfaces.ReportNodeRun
	dependencyGraph *topsort.Graph
	client          db_common.Client
	runs            map[string]reportinterfaces.ReportNodeRun
	workspace       *workspace.Workspace
}

// NewReportExecutionTree creates a result group from a ModTreeIt
func NewReportExecutionTree(reportName string, client db_common.Client, workspace *workspace.Workspace) (*ReportExecutionTree, error) {
	// now populate the ReportExecutionTree
	reportExecutionTree := &ReportExecutionTree{
		client:          client,
		dependencyGraph: topsort.NewGraph(),
		runs:            make(map[string]reportinterfaces.ReportNodeRun),
		workspace:       workspace,
	}

	// create the root run node (either a report run or a panel run)
	root, err := reportExecutionTree.createRootItem(reportName)
	if err != nil {
		return nil, err
	}

	reportExecutionTree.Root = root
	return reportExecutionTree, nil
}

func (e *ReportExecutionTree) createRootItem(reportName string) (reportinterfaces.ReportNodeRun, error) {
	parsedName, err := modconfig.ParseResourceName(reportName)
	if err != nil {
		return nil, err
	}

	rootParentName := e.workspace.Mod.ShortName
	var root reportinterfaces.ReportNodeRun
	switch parsedName.ItemType {
	case modconfig.BlockTypePanel:
		panel, ok := e.workspace.Panels[reportName]
		if !ok {
			return nil, fmt.Errorf("panel '%s' does not exist in workspace", reportName)
		}
		root = NewPanelRun(panel, rootParentName, e)
	case modconfig.BlockTypeReport:
		report, ok := e.workspace.Reports[reportName]
		if !ok {
			return nil, fmt.Errorf("report '%s' does not exist in workspace", reportName)
		}
		root = NewReportContainerRun(report, rootParentName, e)
	case modconfig.BlockTypeContainer:
		container, ok := e.workspace.Containers[reportName]
		if !ok {
			return nil, fmt.Errorf("report '%s' does not exist in workspace", reportName)
		}
		root = NewReportContainerRun(container, rootParentName, e)
	default:
		return nil, fmt.Errorf("invalid block type '%s' passed to ExecuteReport", reportName)
	}
	return root, nil
}

func (e *ReportExecutionTree) Execute(ctx context.Context) error {
	log.Println("[TRACE]", "begin ReportExecutionTree.Execute")
	defer log.Println("[TRACE]", "end ReportExecutionTree.Execute")

	if e.runStatus() == reportinterfaces.ReportRunComplete {
		// there must be no sql panels to execute
		log.Println("[TRACE]", "execution tree already complete")
		return nil
	}
	//get the dependency order
	executionOrder, err := e.dependencyGraph.TopSort(e.Root.GetName())
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

func (e *ReportExecutionTree) runStatus() reportinterfaces.ReportRunStatus {
	return e.Root.GetRunStatus()
}

func (e *ReportExecutionTree) ExecuteNode(ctx context.Context, name string) error {
	runNode, ok := e.runs[name]
	if !ok {
		// this error will be passed up the execution tree and raised as a report error for the root node
		return fmt.Errorf("'%s' not found in execution tree", name)
	}
	switch run := runNode.(type) {
	case *ReportContainerRun:
		// panel should now be complete, i.e. all it's children should be complete
		if !run.ChildrenComplete() {
			// this error will be passed up the execution tree and raised as a report error for the root node
			return fmt.Errorf("'%s' should be complete, but it has incomplete children", run.Name)
		}
		// set complete status on report - this will raise panel complete event
		run.SetComplete()
		return nil
	case *PanelRun:
		// if panel has sql execute it
		if run.SQL != "" {
			data, err := e.executePanelSQL(ctx, run.SQL)
			if err != nil {
				// set the error status on the panel - this will raise panel error event
				run.SetError(err)

				return err
			}

			run.Data = data
		}
		// set complete status on panel - this will raise panel complete event
		run.SetComplete()
		return nil
	default:
		return fmt.Errorf("invalid block type '%s' passed to ReportExecutionTree.ExecuteNode", name)
	}
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
