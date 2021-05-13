package tabledisplay

import (
	"fmt"

	"github.com/turbot/steampipe/control/controlresult"
)

type GroupRenderer struct {
	id string

	failedControls    int
	totalControls     int
	maxFailedControls int
	maxTotalControls  int
	// screen width
	width int
}

func NewGroupRenderer(result *controlresult.ResultGroup, maxFailedControls, maxTotalControls, width int) *GroupRenderer {
	return &GroupRenderer{
		id:                result.GroupId,
		failedControls:    result.Summary.Status.FailedCount(),
		totalControls:     result.Summary.Status.TotalCount(),
		maxFailedControls: maxFailedControls,
		maxTotalControls:  maxTotalControls,
		width:             width,
	}
}

func (g GroupRenderer) String() string {
	counter := NewCounterRenderer(g.failedControls, g.totalControls, g.maxFailedControls, g.maxTotalControls)
	counterString := counter.String()
	graphString := NewCounterGraphRenderer(g.failedControls, g.totalControls, g.maxTotalControls).String()

	// figure out how much width we have available for the id
	availableWidth := g.width - len(counterString) - len(graphString)
	// if the id is longer than the available width, try the short counter string
	// (if the fail count is zero, there is a short version of the counter)
	if availableWidth < len(g.id) {
		counterString = counter.ShortString()
		availableWidth = g.width - len(counterString) - len(graphString)
	}

	// now availableWidth is all we have - if it is not enough we need to truncate the id
	groupIdString := NewGroupIdRenderer(g.id, availableWidth).String()

	// is there any room for a spacer
	spacerWidth := availableWidth - len(groupIdString)
	var spacerString string
	if spacerWidth > 0 {
		spacerString = NewSpacerRenderer(spacerWidth).String()
	}

	// now put these all together
	str := fmt.Sprintf("%s %s %s %s", groupIdString, spacerString, counterString, graphString)
	//fmt.Println(str)
	return str
}
