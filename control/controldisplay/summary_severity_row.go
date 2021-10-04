package controldisplay

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/control/controlexecute"
)

type SummarySeverityRowRenderer struct {
	resultTree *controlexecute.ExecutionTree
	width      int
	severity   string
}

func NewSummarySeverityRowRenderer(resultTree *controlexecute.ExecutionTree, width int, severity string) *SummarySeverityRowRenderer {
	return &SummarySeverityRowRenderer{
		resultTree: resultTree,
		width:      width,
		severity:   severity,
	}
}

func (r *SummarySeverityRowRenderer) Render() string {
	var severitySummary controlexecute.StatusSummary
	if val, exists := r.resultTree.Root.Severity[r.severity]; exists {
		severitySummary = val
	}

	colorFunc := ControlColors.Severity
	severityStr := fmt.Sprintf("%s ", colorFunc(strings.ToUpper(r.severity)))

	count := NewCounterRenderer(
		severitySummary.FailedCount(),
		severitySummary.TotalCount(),
		r.resultTree.Root.Summary.Status.FailedCount(), // not sure what this is
		r.resultTree.Root.Summary.Status.TotalCount(),
		false,
	).Render()

	graph := NewCounterGraphRenderer(
		severitySummary.FailedCount(),
		severitySummary.TotalCount(),
		r.resultTree.Root.Summary.Status.TotalCount(),
		CounterGraphRendererOptions{
			FailedColorFunc: ControlColors.CountGraphFail,
		},
	).Render()

	spaceAvailable := r.width - (helpers.PrintableLength(severityStr) + helpers.PrintableLength(count) + helpers.PrintableLength(graph))
	space := ""
	if r.severity == "critical" {
		space = NewSpacerRenderer(4).Render()
	} else {
		space = NewSpacerRenderer(spaceAvailable).Render()
	}

	return fmt.Sprintf(
		"%s%s%s%s",
		severityStr,
		space,
		colorFunc(count),
		colorFunc(graph),
	)
}
