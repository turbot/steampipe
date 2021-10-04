package controldisplay

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/control/controlexecute"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type SummaryStatusRowRenderer struct {
	resultTree *controlexecute.ExecutionTree
	width      int
	status     string
}

func NewSummaryStatusRowRenderer(resultTree *controlexecute.ExecutionTree, width int, status string) *SummaryStatusRowRenderer {
	return &SummaryStatusRowRenderer{
		resultTree: resultTree,
		width:      width,
		status:     status,
	}
}

func (r *SummaryStatusRowRenderer) Render() string {

	colorFunction := ControlColors.StatusColors[strings.ToLower(r.status)]
	count := -1
	switch r.status {
	case "ok":
		count = r.resultTree.Root.Summary.Status.Ok
	case "skip":
		count = r.resultTree.Root.Summary.Status.Skip
	case "info":
		count = r.resultTree.Root.Summary.Status.Info
	case "alarm":
		count = r.resultTree.Root.Summary.Status.Alarm
	case "error":
		count = r.resultTree.Root.Summary.Status.Error
	}
	countString := r.getPrintableNumber(count, colorFunction)

	graph := NewCounterGraphRenderer(
		count,
		count,
		r.resultTree.Root.Summary.Status.TotalCount(),
		CounterGraphRendererOptions{
			FailedColorFunc: colorFunction,
		},
	).Render()

	statusStr := fmt.Sprintf("%s ", colorFunction(strings.ToUpper(r.status)))
	spaceAvailableForSpacer := r.width - (helpers.PrintableLength(statusStr) + helpers.PrintableLength(countString) + helpers.PrintableLength(graph))
	spacer := NewSpacerRenderer(spaceAvailableForSpacer)

	return fmt.Sprintf(
		"%s%s%s%s",
		statusStr,
		colorFunction(spacer.Render()),
		countString,
		graph,
	)
}
func (r *SummaryStatusRowRenderer) getPrintableNumber(number int, cf colorFunc) string {
	p := message.NewPrinter(language.English)
	s := p.Sprintf("%d", number)
	return fmt.Sprintf("%s ", cf(s))
}
