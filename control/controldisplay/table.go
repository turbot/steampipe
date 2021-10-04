package controldisplay

import (
	"bytes"

	"github.com/turbot/steampipe/control/controlexecute"
)

type TableRenderer struct {
	resultTree *controlexecute.ExecutionTree

	// screen width
	width             int
	maxFailedControls int
	maxTotalControls  int
}

func NewTableRenderer(resultTree *controlexecute.ExecutionTree, width int) *TableRenderer {
	return &TableRenderer{
		resultTree:        resultTree,
		width:             width,
		maxFailedControls: resultTree.Root.Summary.Status.FailedCount(),
		maxTotalControls:  resultTree.Root.Summary.Status.TotalCount(),
	}
}

func (r TableRenderer) Render() string {
	// the buffer to put the output data in
	outbuf := bytes.NewBufferString("")

	outbuf.WriteString(r.renderResult())
	outbuf.WriteString("\n")
	outbuf.WriteString(r.renderSummary())

	return outbuf.String()
}

func (r TableRenderer) renderSummary() string {
	return NewSummaryRenderer(r.resultTree, r.width).Render()
}

func (r TableRenderer) renderResult() string {
	return NewGroupRenderer(r.resultTree.Root, nil, r.maxFailedControls, r.maxTotalControls, r.resultTree, r.width).Render()
}
