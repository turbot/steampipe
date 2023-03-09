package controldisplay

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
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
	severityRows := NewSummarySeverityRenderer(r.resultTree, availableWidth).Render()
	// now get the length of the longest row from the severity block (if any)
	severityWidth := 0
	for _, row := range severityRows {
		if w := helpers.PrintableLength(row); w > severityWidth {
			severityWidth = w
		}
	}
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
	if len(severityRows) > 0 {
		summaryLines = append(summaryLines, "") // blank line
		summaryLines = append(summaryLines, severityRows...)
	}
	// now add the summary
	summaryLines = append(summaryLines,
		"", // blank line
		// summary row
		summaryRow,
	)

	return strings.Join(summaryLines, "\n")
}
