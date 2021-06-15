package controldisplay

import (
	"github.com/turbot/steampipe/control/execute"
)

type TableRenderer struct {
	resultTree *execute.ExecutionTree

	// screen width
	width             int
	maxFailedControls int
	maxTotalControls  int
}

func NewTableRenderer(resultTree *execute.ExecutionTree, width int) *TableRenderer {
	return &TableRenderer{
		resultTree:        resultTree,
		width:             width,
		maxFailedControls: resultTree.Root.Summary.Status.FailedCount(),
		maxTotalControls:  resultTree.Root.Summary.Status.TotalCount(),
	}
}

func (r TableRenderer) Render() string {
	// leading newline
	return NewGroupRenderer(r.resultTree.Root, nil, r.maxFailedControls, r.maxTotalControls, r.resultTree, r.width).Render()
}
