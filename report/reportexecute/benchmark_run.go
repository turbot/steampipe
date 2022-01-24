package reportexecute

import (
	"context"
	"log"

	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// BenchmarkRun is a struct representing the execution of a leaf reporting node
type BenchmarkRun struct {
	Name string `json:"name"`

	Title                string               `json:"title,omitempty"`
	Width                int                  `json:"width,omitempty"`
	Data                 *LeafData            `json:"data,omitempty"`
	Error                error                `json:"error,omitempty"`
	benchmark            *modconfig.Benchmark `json:"properties"`
	NodeType             string               `json:"node_type"`
	Path                 []string             `json:"-"`
	parent               reportinterfaces.ReportNodeParent
	runStatus            reportinterfaces.ReportRunStatus
	executionTree        *ReportExecutionTree
	ControlExecutionTree *controlexecute.ExecutionTree
}

func NewBenchmarkRun(benchmark *modconfig.Benchmark, parent reportinterfaces.ReportNodeParent, executionTree *ReportExecutionTree) (*BenchmarkRun, error) {
	r := &BenchmarkRun{
		Name:          benchmark.Name(),
		Title:         benchmark.GetTitle(),
		Width:         benchmark.GetWidth(),
		Path:          benchmark.GetPaths()[0],
		benchmark:     benchmark,
		executionTree: executionTree,
		parent:        parent,

		// set to complete, optimistically
		// if any children have SQL we will set this to ReportRunReady instead
		runStatus: reportinterfaces.ReportRunComplete,
	}

	parsedName, err := modconfig.ParseResourceName(benchmark.Name())
	if err != nil {
		return nil, err
	}
	r.NodeType = parsedName.ItemType
	//  set status to ready
	r.runStatus = reportinterfaces.ReportRunReady

	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r, nil
}

// Execute implements ReportRunNode
func (r *BenchmarkRun) Execute(ctx context.Context) error {
	executionTree, err := controlexecute.NewExecutionTree(ctx, r.executionTree.workspace, r.executionTree.client, r.benchmark.Name())
	if err != nil {
		log.Printf("[WARN] %s Benchmark execution error %v", r.Name, err)
		// set the error status on the counter - this will raise counter error event
		r.SetError(err)
		return err
	}
	executionTree.Execute(ctx)
	r.ControlExecutionTree = executionTree

	log.Printf("[WARN] %s SetComplete", r.Name)
	// set complete status on counter - this will raise counter complete event
	r.SetComplete()

	log.Printf("[WARN] %s Execute DONE", r.Name)
	return nil
}

// GetName implements ReportNodeRun
func (r *BenchmarkRun) GetName() string {
	return r.Name
}

// GetPath implements ReportNodeRun
func (r *BenchmarkRun) GetPath() modconfig.NodePath {
	return r.Path
}

// GetRunStatus implements ReportNodeRun
func (r *BenchmarkRun) GetRunStatus() reportinterfaces.ReportRunStatus {
	return r.runStatus
}

// SetError implements ReportNodeRun
func (r *BenchmarkRun) SetError(err error) {
	r.Error = err
	r.runStatus = reportinterfaces.ReportRunError
	// raise counter error event
	r.executionTree.workspace.PublishReportEvent(&reportevents.LeafNodeError{Node: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r

}

// SetComplete implements ReportNodeRun
func (r *BenchmarkRun) SetComplete() {
	r.runStatus = reportinterfaces.ReportRunComplete
	// raise counter complete event
	log.Printf("[WARN] **************** BenchmarkRun DONE EVENT %s ***************", r.Name)
	r.executionTree.workspace.PublishReportEvent(&reportevents.LeafNodeComplete{Node: r})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements ReportNodeRun
func (r *BenchmarkRun) RunComplete() bool {
	return r.runStatus == reportinterfaces.ReportRunComplete || r.runStatus == reportinterfaces.ReportRunError
}

// ChildrenComplete implements ReportNodeRun
func (r *BenchmarkRun) ChildrenComplete() bool {
	return r.RunComplete()
}
