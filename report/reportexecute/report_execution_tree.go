package reportexecute

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
)

// ReportExecutionTree is a structure representing the control result hierarchy
type ReportExecutionTree struct {
	modconfig.UniqueNameProviderBase

	Root        *ReportContainerRun
	client      db_common.Client
	runs        map[string]reportinterfaces.ReportNodeRun
	workspace   *workspace.Workspace
	runComplete chan reportinterfaces.ReportNodeRun
	inputLock   sync.Mutex

	inputDataSubscriptions map[string][]chan bool
}

// NewReportExecutionTree creates a result group from a ModTreeItem
func NewReportExecutionTree(reportName string, client db_common.Client, workspace *workspace.Workspace) (*ReportExecutionTree, error) {
	// now populate the ReportExecutionTree
	reportExecutionTree := &ReportExecutionTree{
		client:                 client,
		runs:                   make(map[string]reportinterfaces.ReportNodeRun),
		workspace:              workspace,
		runComplete:            make(chan reportinterfaces.ReportNodeRun, 1),
		inputDataSubscriptions: make(map[string][]chan bool),
	}

	// create the root run node (either a report run or a counter run)
	root, err := reportExecutionTree.createRootItem(reportName)
	if err != nil {
		return nil, err
	}

	reportExecutionTree.Root = root
	return reportExecutionTree, nil
}

func (e *ReportExecutionTree) createRootItem(reportName string) (*ReportContainerRun, error) {
	parsedName, err := modconfig.ParseResourceName(reportName)
	if err != nil {
		return nil, err
	}

	if parsedName.ItemType != modconfig.BlockTypeReport {
		return nil, fmt.Errorf("reporting type %s cannot be executed directly - only reports may be executed", parsedName.ItemType)
	}
	report, ok := e.workspace.Reports[reportName]
	if !ok {
		return nil, fmt.Errorf("report '%s' does not exist in workspace", reportName)
	}
	return NewReportContainerRun(report, e, e)

}

func (e *ReportExecutionTree) Execute(ctx context.Context) error {
	log.Println("[TRACE]", "begin ReportExecutionTree.Execute")
	defer log.Println("[TRACE]", "end ReportExecutionTree.Execute")

	if e.runStatus() == reportinterfaces.ReportRunComplete {
		// there must be no nodes to execute
		log.Println("[TRACE]", "execution tree already complete")
		return nil
	}

	return e.Root.Execute(ctx)
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
func (e *ReportExecutionTree) ChildCompleteChan() chan reportinterfaces.ReportNodeRun {
	return e.runComplete
}

func (e *ReportExecutionTree) waitForRuntimeDependency(ctx context.Context, dependency *modconfig.RuntimeDependency) error {
	depChan := make(chan (bool), 1)
	// TOTO [reports] for now verify we only bind inputs to args (somewhere)

	e.subscribeToInput(dependency.PropertyPath.Name, depChan)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-depChan:
		return nil
	}
}

func (e *ReportExecutionTree) subscribeToInput(inputName string, depChan chan bool) {
	e.inputLock.Lock()
	defer e.inputLock.Unlock()

	e.inputDataSubscriptions[inputName] = append(e.inputDataSubscriptions[inputName], depChan)
}
