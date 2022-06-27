package controldisplay

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
)

type SummaryTotalRowRenderer struct {
	resultTree *controlexecute.ExecutionTree
	width      int
}

func NewSummaryTotalRowRenderer(resultTree *controlexecute.ExecutionTree, width int) *SummaryTotalRowRenderer {
	return &SummaryTotalRowRenderer{
		resultTree: resultTree,
		width:      width,
	}
}

func (r *SummaryTotalRowRenderer) Render() string {

	head := fmt.Sprintf("%s ", ControlColors.GroupTitle("TOTAL"))
	count := NewCounterRenderer(
		r.resultTree.Root.Summary.Status.FailedCount(),
		r.resultTree.Root.Summary.Status.TotalCount(),
		r.resultTree.Root.Summary.Status.FailedCount(),
		r.resultTree.Root.Summary.Status.TotalCount(),
		CounterRendererOptions{
			AddLeadingSpace: false,
		},
	).Render()

	graph := NewCounterGraphRenderer(
		r.resultTree.Root.Summary.Status.FailedCount(),
		r.resultTree.Root.Summary.Status.TotalCount(),
		r.resultTree.Root.Summary.Status.TotalCount(),
		CounterGraphRendererOptions{
			FailedColorFunc: ControlColors.CountGraphFail,
		},
	).Render()

	spaceWidth := r.width - (helpers.PrintableLength(head) + helpers.PrintableLength(count) + helpers.PrintableLength(graph))

	spacer := NewSpacerRenderer(spaceWidth)

	return fmt.Sprintf(
		"%s%s%s%s",
		head,
		spacer.Render(),
		count,
		graph,
	)
}
