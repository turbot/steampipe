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

func (r *SummarySeverityRowRenderer) countWithSeverity(group *controlexecute.ResultGroup) (int, int) {
	runsWithThisSeverity := 0
	runsWithSeverityValue := 0

	for _, subGroup := range group.Groups {
		v, any := r.countWithSeverity(subGroup)
		runsWithThisSeverity += v
		runsWithSeverityValue += any
	}

	for _, run := range group.ControlRuns {
		if len(run.Severity) == 0 {
			continue
		}

		runsWithSeverityValue += 1
		if strings.EqualFold(run.Severity, r.severity) {
			runsWithThisSeverity += 1
		}
	}

	return runsWithThisSeverity, runsWithSeverityValue
}

func (r *SummarySeverityRowRenderer) Render() string {
	colorFunc := ControlColors.Severity

	severityStr := fmt.Sprintf("%s ", colorFunc(strings.ToUpper(r.severity)))
	sevCount, withSevValueCount := r.countWithSeverity(r.resultTree.Root)

	count := NewCounterRenderer(
		sevCount,
		withSevValueCount,
		r.resultTree.Root.Summary.Status.FailedCount(), // not sure what this is
		r.resultTree.Root.Summary.Status.TotalCount(),
	).Render()

	graph := NewCounterGraphRenderer(
		sevCount,
		withSevValueCount,
		r.resultTree.Root.Summary.Status.TotalCount(),
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
