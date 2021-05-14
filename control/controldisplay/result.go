package controldisplay

import (
	"fmt"

	"github.com/turbot/steampipe/control/execute"
)

const minReasonWidth = 5

type ResultRenderer struct {
	status     string
	reason     string
	dimensions []execute.Dimension
	colorMap   execute.DimensionColorMap

	// screen width
	width int
}

func NewResultRenderer(status, reason string, dimensions []execute.Dimension, colorMap execute.DimensionColorMap, width int) *ResultRenderer {
	return &ResultRenderer{
		status:     status,
		reason:     reason,
		dimensions: dimensions,
		colorMap:   colorMap,
		width:      width,
	}
}

func (r ResultRenderer) Render() string {
	status := NewResultStatusRenderer(r.status)
	statusString, statusWidth := status.Render()

	// figure out how much width we have available for the  dimensions, allowing the minimum for the reason
	availableWidth := r.width - statusWidth

	// for now give this all to reason
	availableDimensionWidth := availableWidth - minDimensionWidth
	dimensionsString, dimensionsWidth := NewDimensionsRenderer(r.dimensions, r.colorMap, availableDimensionWidth).Render()

	availableWidth -= dimensionsWidth

	// now availableWidth is all we have - if it is not enough we need to truncate the reason
	reasonString, reasonWidth := NewResultReasonRenderer(r.status, r.reason, availableWidth).Render()

	// is there any room for a spacer
	availableWidth -= reasonWidth
	var spacerString string
	if availableWidth > 0 {
		spacerString, _ = NewSpacerRenderer(availableWidth).Render()
	}

	// now put these all together
	str := fmt.Sprintf("%s%s%s%s", statusString, reasonString, spacerString, dimensionsString)
	return str
}
