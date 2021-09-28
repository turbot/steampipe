package controldisplay

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/turbot/steampipe/control/controlexecute"
)

type SummaryRenderer struct {
	resultTree *controlexecute.ExecutionTree

	// screen width
	width             int
	maxFailedControls int
	maxTotalControls  int
}

func NewSummaryRenderer(resultTree *controlexecute.ExecutionTree, width int) *SummaryRenderer {
	return &SummaryRenderer{
		resultTree:        resultTree,
		width:             width,
		maxFailedControls: resultTree.Root.Summary.Status.FailedCount(),
		maxTotalControls:  resultTree.Root.Summary.Status.TotalCount(),
	}
}

func (r SummaryRenderer) Render() string {
	outbuf := bytes.NewBufferString("")
	table := tablewriter.NewWriter(outbuf)
	table.SetHeader([]string{})                // no headers please
	table.SetAlignment(tablewriter.ALIGN_LEFT) // we align to the left
	table.SetAutoWrapText(false)               // let's not wrap the text
	table.SetBorder(true)                      // there needs to be a border to give the dialog feel
	table.AppendBulk([][]string{
		{r.renderStatus("Alarm", r.resultTree.Root.Summary.Status.Alarm, r.resultTree.Root.Summary.Status.TotalCount())},
		{r.renderStatus("Ok", r.resultTree.Root.Summary.Status.Ok, r.resultTree.Root.Summary.Status.TotalCount())},
		{r.renderStatus("Info", r.resultTree.Root.Summary.Status.Info, r.resultTree.Root.Summary.Status.TotalCount())},
		{r.renderStatus("Skip", r.resultTree.Root.Summary.Status.Skip, r.resultTree.Root.Summary.Status.TotalCount())},
		{r.renderStatus("Error", r.resultTree.Root.Summary.Status.Error, r.resultTree.Root.Summary.Status.TotalCount())},
	}) // Add Bulk Data

	table.Render()

	return outbuf.String()
}

func (r SummaryRenderer) renderStatus(status string, count int, total int) string {
	statusColorFunction := ControlColors.StatusColors[strings.ToLower(status)]
	countColorFunction := ControlColors.ReasonColors[strings.ToLower(status)]

	countString := fmt.Sprintf("%d", count)

	return fmt.Sprintf(
		"%s %s%s %s / %d",
		statusColorFunction(status),
		countColorFunction(padLeft(status, (r.width/2)-6)),
		countColorFunction(padLeft(countString, (r.width/2)-6)),
		countColorFunction(countString),
		total,
	)
}

func padLeft(str string, width int) string {
	if len(str) > width {
		return str
	}

	return strings.Repeat(".", width-len(str))
}
