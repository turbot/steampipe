package controldisplay

import (
	"fmt"
	"strings"

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
	availableWidth := r.width

	// first build the severity block - if it exists, it will be used to dictate the max width
	severityBlock := NewSummarySeverityRenderer(r.resultTree, availableWidth, "ok").Render()
	severityWidth := helpers.PrintableLength(severityBlock)
	if severityWidth > 0 {
		availableWidth = severityWidth
	}

	// now do the summary row - this is the next longest row
	summaryRow := NewSummaryTotalRowRenderer(r.resultTree, availableWidth).Render()
	// if there is no severity block, use the summary row to dictate the max width
	if severityWidth == 0 {
		availableWidth = helpers.PrintableLength(summaryRow)
	}

	okStatusRow := NewSummaryStatusRowRenderer(r.resultTree, availableWidth, "ok").Render()
	skipStatusRow := NewSummaryStatusRowRenderer(r.resultTree, availableWidth, "skip").Render()
	infoStatusRow := NewSummaryStatusRowRenderer(r.resultTree, availableWidth, "info").Render()
	alarmStatusRow := NewSummaryStatusRowRenderer(r.resultTree, availableWidth, "alarm").Render()
	errorStatusRow := NewSummaryStatusRowRenderer(r.resultTree, availableWidth, "error").Render()

	titleLine := fmt.Sprintf("%s\n", ControlColors.GroupTitle("Summary"))

	// build the summary

	var summaryLines = []string{
		titleLine,
		// status summaries
		okStatusRow,
		skipStatusRow,
		infoStatusRow,
		alarmStatusRow,
		errorStatusRow,
	}
	// if there is a severity block, add it
	if severityWidth > 0 {
		summaryLines = append(summaryLines,
			"", // blank line
			severityBlock)
	}
	// now add the summary
	summaryLines = append(summaryLines,
		"", // blank line
		// summary row
		summaryRow,
	)

	return strings.Join(summaryLines, "\n")
}
