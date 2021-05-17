package controldisplay

import (
	"fmt"
	"log"

	"github.com/turbot/go-kit/helpers"
)

type GroupHeadingRenderer struct {
	title             string
	severity          string
	failedControls    int
	totalControls     int
	maxFailedControls int
	maxTotalControls  int
	// screen width
	width int
}

func NewGroupHeadingRenderer(title string, failed, total, maxFailed, maxTotal, width int) *GroupHeadingRenderer {
	return &GroupHeadingRenderer{
		title:             title,
		failedControls:    failed,
		totalControls:     total,
		maxFailedControls: maxFailed,
		maxTotalControls:  maxTotal,
		width:             width,
	}
}

func (r GroupHeadingRenderer) Render() string {
	log.Println("[TRACE] begin group heading render")
	defer log.Println("[TRACE] end group heading render")

	if r.width <= 0 {
		log.Printf("[WARN] group renderer has width of %d\n", r.width)
		return ""
	}

	severityString := NewSeverityRenderer(r.severity).Render()
	severityWidth := helpers.PrintableLength(severityString)

	counterString := NewCounterRenderer(r.failedControls, r.totalControls, r.maxFailedControls, r.maxTotalControls).Render()
	counterWidth := helpers.PrintableLength(counterString)

	graphString := NewCounterGraphRenderer(r.failedControls, r.totalControls, r.maxTotalControls).Render()
	graphWidth := helpers.PrintableLength(graphString)

	// figure out how much width we have available for the title
	availableWidth := r.width - counterWidth - graphWidth - severityWidth

	// now availableWidth is all we have - if it is not enough we need to truncate the title
	titleString := NewGroupTitleRenderer(r.title, availableWidth).Render()
	titleWidth := helpers.PrintableLength(titleString)

	// is there any room for a spacer
	spacerWidth := availableWidth - titleWidth
	var spacerString string
	if spacerWidth > 0 {
		spacerString = NewSpacerRenderer(spacerWidth).Render()
	}

	// now put these all together
	str := fmt.Sprintf("%s%s%s%s%s", titleString, spacerString, severityString, counterString, graphString)
	return str
}
