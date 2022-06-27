package controldisplay

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
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
	severitySummary, exists := r.resultTree.Root.Summary.Severity[r.severity]
	// if there are no items for this severity level, return empty string
	if !exists {
		return ""
	}
	colorFunc := ControlColors.Severity
	severityStr := fmt.Sprintf("%s ", colorFunc(strings.ToUpper(r.severity)))

	count := NewCounterRenderer(
		severitySummary.FailedCount(),
		severitySummary.TotalCount(),
		r.resultTree.Root.Summary.Status.FailedCount(),
		r.resultTree.Root.Summary.Status.TotalCount(),
		CounterRendererOptions{
			AddLeadingSpace: false,
		},
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
		count,
		graph,
	)
}
