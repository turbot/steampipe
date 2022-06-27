package controldisplay

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
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
	txtColorFunction := ControlColors.StatusColors[r.status]
	graphColorFunction := ControlColors.GraphColors[r.status]

	count := -1
	switch r.status {
	case constants.ControlOk:
		count = r.resultTree.Root.Summary.Status.Ok
	case constants.ControlSkip:
		count = r.resultTree.Root.Summary.Status.Skip
	case constants.ControlInfo:
		count = r.resultTree.Root.Summary.Status.Info
	case constants.ControlAlarm:
		count = r.resultTree.Root.Summary.Status.Alarm
	case constants.ControlError:
		count = r.resultTree.Root.Summary.Status.Error
	default:
		// we can safely panic here, since the status enum check should have been
		// done by the executor. this is here for unit tests mostly
		panic(fmt.Sprintf("unknown status: %s", r.status))
	}
	countString := r.getPrintableNumber(count, txtColorFunction)

	graph := NewCounterGraphRenderer(
		count,
		count,
		r.resultTree.Root.Summary.Status.TotalCount(),
		CounterGraphRendererOptions{
			FailedColorFunc: graphColorFunction,
		},
	).Render()

	statusStr := fmt.Sprintf("%s ", txtColorFunction(strings.ToUpper(r.status)))
	spaceAvailableForSpacer := r.width - (helpers.PrintableLength(statusStr) + helpers.PrintableLength(countString) + helpers.PrintableLength(graph))
	spacer := NewSpacerRenderer(spaceAvailableForSpacer)

	return fmt.Sprintf(
		"%s%s%s%s",
		statusStr,
		spacer.Render(),
		countString,
		graph,
	)
}

func (r *SummaryStatusRowRenderer) getPrintableNumber(number int, cf colorFunc) string {
	p := message.NewPrinter(language.English)
	s := p.Sprintf("%d", number)
	return fmt.Sprintf("%s ", cf(s))
}
