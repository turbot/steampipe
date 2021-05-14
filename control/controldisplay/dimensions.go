package controldisplay

import (
	"fmt"
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
func (d DimensionsRenderer) Render() (string, int) {
	// make array of dimension values (including trailing spaces
	var formattedDimensions = make([]string, len(d.dimensions))
	for i, d := range d.dimensions {
		formattedDimensions[i] = fmt.Sprintf(" %s", d.Value)
	}

	minDimensionWidth := 3
	var length int
	for length = dimensionsLength(formattedDimensions); length > d.width; {
		// truncate each dimension in turn to a min of const minDimensionWidth, until we satisfy the width requirement
		for i, v := range formattedDimensions {
			if len(v) > minDimensionWidth {
				// truncate the original value, not the already truncated value
				formattedDimensions[i] = helpers.TruncateString(d.dimensions[i].Value, len(v)-1)
				break
			}
			// to get here all dimensions are at the min dimension width and we are still too long - reduce min width
			if minDimensionWidth >= 2 {
				minDimensionWidth--
			} else {
				// so event with all dimensions 1 long, we still do not have enough space
				// remove a dimension from the array
				if len(formattedDimensions) > 2 {
					formattedDimensions = formattedDimensions[1:]
				} else {
					// there is only 1 dimension - nothing we can do here, give up
					return "", 0
				}
			}
		}
	}

	// ok we now have dimensions that fir in the space, color them
	for i, v := range formattedDimensions {
		// get the source dimension object
		dimension := d.dimensions[i]
		// get the color code - there must be an entry
		color := d.colorMap[dimension.Key][dimension.Value]

		formattedDimensions[i] = fmt.Sprintf("%s", aurora.Index(color, v))
	}

	return strings.Join(formattedDimensions, ""), length
}

// count the
func dimensionsLength(dimensionValues []string) int {
	var res int
	for _, v := range dimensionValues {
		res += len(v)
	}
	return res
}
