package controldisplay

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
)

type SummarySeverityRenderer struct {
	resultTree *controlexecute.ExecutionTree
	width      int
}

func NewSummarySeverityRenderer(resultTree *controlexecute.ExecutionTree, width int) *SummarySeverityRenderer {
	return &SummarySeverityRenderer{
		resultTree: resultTree,
		width:      width,
	}
}

func (r *SummarySeverityRenderer) Render() []string {
	availableWidth := r.width

	// render the critical line
	criticalSeverityRow := NewSummarySeverityRowRenderer(r.resultTree, availableWidth, "critical").Render()
	criticalWidth := helpers.PrintableLength(criticalSeverityRow)
	// if there is a critical line, use this to set the max width
	if criticalWidth > 0 {
		availableWidth = criticalWidth
	}

	// render the high line
	highSeverityRow := NewSummarySeverityRowRenderer(r.resultTree, availableWidth, "high").Render()
	highWidth := helpers.PrintableLength(highSeverityRow)

	// build the severity block
	var strs []string
	if criticalWidth > 0 {
		strs = append(strs, criticalSeverityRow)
	}
	if highWidth > 0 {
		strs = append(strs, highSeverityRow)
	}
	return strs
}
