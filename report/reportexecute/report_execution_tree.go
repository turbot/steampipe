package reportexecute

import (
	"context"
	"fmt"
	"log"

	"github.com/turbot/steampipe/utils"

	"github.com/stevenle/topsort"
	"github.com/turbot/steampipe/db/db_common"
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
	runComplete     chan (bool)
}

// NewReportExecutionTree creates a result group from a ModTreeIt
func NewReportExecutionTree(reportName string, client db_common.Client, workspace *workspace.Workspace) (*ReportExecutionTree, error) {
	// now populate the ReportExecutionTree
	reportExecutionTree := &ReportExecutionTree{
		client:          client,
		dependencyGraph: topsort.NewGraph(),
		runs:            make(map[string]reportinterfaces.ReportNodeRun),
		workspace:       workspace,
		runComplete:     make(chan (bool), 1),
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
	// TODO CAN THIS BE ANYTHING OTHER THAN A REPORT??
	var root reportinterfaces.ReportNodeRun
	switch parsedName.ItemType {
	case modconfig.BlockTypePanel:
		panel, ok := e.workspace.Panels[reportName]
		if !ok {
			return nil, fmt.Errorf("panel '%s' does not exist in workspace", reportName)
		}
		root = NewPanelRun(panel, e, e)
	case modconfig.BlockTypeReport:
		report, ok := e.workspace.Reports[reportName]
		if !ok {
			return nil, fmt.Errorf("report '%s' does not exist in workspace", reportName)
		}
		root = NewReportContainerRun(report, e, e)
	case modconfig.BlockTypeContainer:
		container, ok := e.workspace.Containers[reportName]
		if !ok {
			return nil, fmt.Errorf("report '%s' does not exist in workspace", reportName)
		}
		root = NewReportContainerRun(container, e, e)
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
	errorChan := make(chan error, len(executionOrder))
	for _, name := range executionOrder {
		runNode, ok := e.runs[name]
		if !ok {
			// should never happen
			return fmt.Errorf("'%s' not found in execution tree", name)
		}
		go func() {
			if err := runNode.Execute(ctx); err != nil {
				errorChan <- err
			}
		}()
	}

	// wait for root completion
	var errors []error
	for {
		select {
		case <-e.runComplete:
			break
		case err := <-errorChan:
			errors = append(errors, err)
			// TODO TIMEOUT??
		}
	}

	return utils.CombineErrors(errors...)
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

// GetName implements ReportNodeParent
// use mod chort name - this will be the root name for all child runs
func (e *ReportExecutionTree) GetName() string {
	return e.workspace.Mod.ShortName
}

// ChildCompleteChan implements ReportNodeParent
func (e *ReportExecutionTree) ChildCompleteChan() chan bool {
	return e.runComplete
}
