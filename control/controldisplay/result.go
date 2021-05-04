package controldisplay

import (
	"fmt"
	"log"

	"github.com/turbot/go-kit/helpers"

	"github.com/turbot/steampipe/control/execute"
)

const minReasonWidth = 10

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
	log.Println("[TRACE] begin result render")
	defer log.Println("[TRACE] end result render")

	status := NewResultStatusRenderer(r.status)
	statusString := status.Render()
	statusWidth := helpers.PrintableLength(statusString)

	// figure out how much width we have available for the  dimensions, allowing the minimum for the reason
	availableWidth := r.width - statusWidth

	// for now give this all to reason
	availableDimensionWidth := availableWidth - minReasonWidth
	var dimensionsString string
	var dimensionWidth int
	if availableDimensionWidth > 0 {
		dimensionsString = NewDimensionsRenderer(r.dimensions, r.colorMap, availableDimensionWidth).Render()
		dimensionWidth = helpers.PrintableLength(dimensionsString)
		availableWidth -= dimensionWidth
	}

	// now availableWidth is all we have - if it is not enough we need to truncate the reason
	reasonString := NewResultReasonRenderer(r.status, r.reason, availableWidth).Render()
	reasonWidth := helpers.PrintableLength(reasonString)

	// is there any room for a spacer
	availableWidth -= reasonWidth
	var spacerString string
	if availableWidth > 0 {
		spacerString = NewSpacerRenderer(availableWidth).Render()
	}

	// now put these all together
	str := fmt.Sprintf("%s%s%s%s", statusString, reasonString, spacerString, dimensionsString)
	return str
}
