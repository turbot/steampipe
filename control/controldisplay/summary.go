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
	fmt.Println("Width:", width)
	return &SummaryRenderer{
		resultTree: resultTree,
		width:      width,
	}
}

func (r SummaryRenderer) Render() string {

	criticalSeverityRow := NewSummarySeverityRowRenderer(r.resultTree, r.width, "critical").Render()
	derivedWidth := helpers.PrintableLength(criticalSeverityRow) ///- 3 // why the magic number?

	highSeverityRow := NewSummarySeverityRowRenderer(r.resultTree, derivedWidth, "high").Render()
	fmt.Println("high:", helpers.PrintableLength(highSeverityRow))

	okStatusRow := NewSummaryStatusRowRenderer(r.resultTree, derivedWidth, "ok").Render()
	skipStatusRow := NewSummaryStatusRowRenderer(r.resultTree, derivedWidth, "skip").Render()
	infoStatusRow := NewSummaryStatusRowRenderer(r.resultTree, derivedWidth, "info").Render()
	alarmStatusRow := NewSummaryStatusRowRenderer(r.resultTree, derivedWidth, "alarm").Render()
	errorStatusRow := NewSummaryStatusRowRenderer(r.resultTree, derivedWidth, "error").Render()

	fmt.Println(r.shouldRenderSeverities(r.resultTree.Root))

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

		NewSummaryTotalRowRenderer(r.resultTree, derivedWidth).Render(),
	)
}

func (r SummaryRenderer) shouldRenderSeverities(group *controlexecute.ResultGroup) bool {
	for _, subGroup := range group.Groups {
		return r.shouldRenderSeverities(subGroup)
	}
	for _, run := range group.ControlRuns {
		if len(run.Severity) != 0 {
			return true
		}
	}
	return false
}
