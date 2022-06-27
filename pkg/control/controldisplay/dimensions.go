package controldisplay

import (
	"fmt"
	"log"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
)

type DimensionsRenderer struct {
	dimensions     []controlexecute.Dimension
	colorGenerator *controlexecute.DimensionColorGenerator
	width          int
}

func NewDimensionsRenderer(dimensions []controlexecute.Dimension, colorGenerator *controlexecute.DimensionColorGenerator, width int) *DimensionsRenderer {
	return &DimensionsRenderer{
		dimensions:     dimensions,
		colorGenerator: colorGenerator,
		width:          width,
	}
}

// Render returns the dimensions, truncated to the max length if necessary
func (r DimensionsRenderer) Render() string {
	if r.width <= 0 {
		// this should never happen, since the minimum width is set by the formatter
		log.Printf("[WARN] dimensions renderer has width of %d\n", r.width)
		return ""
	}
	if len(r.dimensions) == 0 {
		return ""
	}
	// make array of dimension values (including trailing spaces
	var formattedDimensions = make([]string, len(r.dimensions))
	for i, d := range r.dimensions {
		formattedDimensions[i] = d.Value
	}

	var length int
	for length = dimensionsLength(formattedDimensions); length > r.width; {
		// truncate the first dimension
		if helpers.PrintableLength(formattedDimensions[0]) > 0 {
			// truncate the original value, not the already truncated value
			newLength := helpers.PrintableLength(formattedDimensions[0]) - 1
			formattedDimensions[0] = helpers.TruncateString(formattedDimensions[0], newLength)
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

	// ok we now have dimensions that fit in the space, color them
	// check whether color is disabled

	for i, v := range formattedDimensions {
		// get the source dimension object
		dimension := r.dimensions[i]

		// get the color code - there must be an entry
		dimensionColorFunc := func(val interface{}) aurora.Value {
			// if current theme supports colors, apply coloring
			if ControlColors.UseColor {
				dimensionColor := r.colorGenerator.Map[dimension.Key][dimension.Value]
				return aurora.Index(dimensionColor, val)
			}
			return aurora.Reset(val)
		}

		formattedDimensions[i] = fmt.Sprintf("%s", dimensionColorFunc(v))
	}

	return strings.Join(formattedDimensions, " ")
}

// count the total length of the dimensions
func dimensionsLength(dimensionValues []string) int {
	var res int
	for _, v := range dimensionValues {
		res += len(v)
	}
	// allow for spaces between the dimensions
	res += len(dimensionValues) - 1
	return res
}
