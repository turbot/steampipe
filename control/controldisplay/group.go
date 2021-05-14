package controldisplay

import (
	"fmt"

	"github.com/turbot/steampipe/control/execute"
)

type GroupRenderer struct {
	title string

	failedControls    int
	totalControls     int
	maxFailedControls int
	maxTotalControls  int
	// screen width
	width int
}

func NewGroupRenderer(result *execute.ResultGroup, maxFailedControls, maxTotalControls, width int) *GroupRenderer {
	return &GroupRenderer{
		title:             result.Title,
		failedControls:    result.Summary.Status.FailedCount(),
		totalControls:     result.Summary.Status.TotalCount(),
		maxFailedControls: maxFailedControls,
		maxTotalControls:  maxTotalControls,
		width:             width,
	}
}

func (g GroupRenderer) Render() string {
	counter := NewCounterRenderer(g.failedControls, g.totalControls, g.maxFailedControls, g.maxTotalControls)
	counterString, counterWidth := counter.Render()
	graphString, graphWidth := NewCounterGraphRenderer(g.failedControls, g.totalControls, g.maxTotalControls).Render()

	// figure out how much width we have available for the title
	availableWidth := g.width - counterWidth - graphWidth

	// now availableWidth is all we have - if it is not enough we need to truncate the title
	titleString, titleWidth := NewGroupTitleRenderer(g.title, availableWidth).Render()

	// is there any room for a spacer

	spacerWidth := availableWidth - titleWidth
	var spacerString string
	if spacerWidth > 0 {
		spacerString, _ = NewSpacerRenderer(spacerWidth).Render()
	}

	// now put these all together
	str := fmt.Sprintf("%s%s%s%s", titleString, spacerString, counterString, graphString)
	return str
}
