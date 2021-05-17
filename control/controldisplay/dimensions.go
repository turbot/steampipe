package controldisplay

import (
	"fmt"
	"log"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/turbot/steampipe/control/execute"

	"github.com/turbot/go-kit/helpers"
)

const minDimensionWidth = 3

type DimensionsRenderer struct {
	dimensions []execute.Dimension
	colorMap   execute.DimensionColorMap
	width      int
}

func NewDimensionsRenderer(dimensions []execute.Dimension, colorMap execute.DimensionColorMap, width int) *DimensionsRenderer {
	return &DimensionsRenderer{
		dimensions: dimensions,
		colorMap:   colorMap,
		width:      width,
	}
}

// Render returns the reason, truncated to the max length if necessary
func (r DimensionsRenderer) Render() string {
	if r.width <= 0 {
		log.Printf("[WARN] dimensions renderer has width of %d\n", r.width)
		return ""
	}
	if len(r.dimensions) == 0 {
		return ""
	}
	// make array of dimension values (including trailing spaces
	var formattedDimensions = make([]string, len(r.dimensions))
	for i, d := range r.dimensions {
		formattedDimensions[i] = fmt.Sprintf(" %s", d.Value)
	}

	var length int
	for length = dimensionsLength(formattedDimensions); length > r.width; {
		// truncate the first dimension

		if len(formattedDimensions[0]) > 0 {
			// truncate the original value, not the already truncated value
			formattedDimensions[0] = helpers.TruncateString(r.dimensions[0].Value, len(formattedDimensions[0])-1)
		} else {
			// so event with all dimensions 1 long, we still do not have enough space
			// remove a dimension from the array
			if len(formattedDimensions) > 2 {
				r.dimensions = r.dimensions[1:]
				formattedDimensions = formattedDimensions[1:]
			} else {
				// there is only 1 dimension - nothing we can do here, give up
				return ""
			}
		}
		// update length
		length = dimensionsLength(formattedDimensions)
	}

	// ok we now have dimensions that fir in the space, color them
	for i, v := range formattedDimensions {
		// get the source dimension object
		dimension := r.dimensions[i]
		// get the color code - there must be an entry
		color := r.colorMap[dimension.Key][dimension.Value]

		formattedDimensions[i] = fmt.Sprintf("%s", aurora.Index(color, v))
	}

	return strings.Join(formattedDimensions, "")
}

// count the
func dimensionsLength(dimensionValues []string) int {
	var res int
	for _, v := range dimensionValues {
		res += len(v)
	}
	return res
}
