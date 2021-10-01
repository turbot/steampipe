package controldisplay

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/control/controlexecute"
)

type SummaryRenderer struct {
	resultTree *controlexecute.ExecutionTree
	width      int
}

func NewSummaryRenderer(resultTree *controlexecute.ExecutionTree, width int) *SummaryRenderer {
	return &SummaryRenderer{
		resultTree: resultTree,
		width:      width,
	}
}

func (r SummaryRenderer) Render() string {

	criticalSeverityRow := NewSummarySeverityRowRenderer(r.resultTree, r.width, "critical").Render()
	availableWidth := helpers.PrintableLength(criticalSeverityRow) ///- 3 // why the magic number?

	highSeverityRow := NewSummarySeverityRowRenderer(r.resultTree, availableWidth, "high").Render()

	okStatusRow := NewSummaryStatusRowRenderer(r.resultTree, availableWidth, "ok").Render()
	skipStatusRow := NewSummaryStatusRowRenderer(r.resultTree, availableWidth, "skip").Render()
	infoStatusRow := NewSummaryStatusRowRenderer(r.resultTree, availableWidth, "info").Render()
	alarmStatusRow := NewSummaryStatusRowRenderer(r.resultTree, availableWidth, "alarm").Render()
	errorStatusRow := NewSummaryStatusRowRenderer(r.resultTree, availableWidth, "error").Render()

	return fmt.Sprintf(
		`
 %s
 
 %s
 %s
 %s
 %s
 %s
 
 %s
 %s
 
 %s
		`,
		ControlColors.GroupTitle("Summary"),

		okStatusRow,
		skipStatusRow,
		infoStatusRow,
		alarmStatusRow,
		errorStatusRow,

		highSeverityRow,
		criticalSeverityRow,

		NewSummaryTotalRowRenderer(r.resultTree, availableWidth).Render(),
	)
}
