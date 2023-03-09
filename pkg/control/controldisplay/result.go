package controldisplay

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
)

const minReasonWidth = 10

type ResultRenderer struct {
	status         string
	reason         string
	dimensions     []controlexecute.Dimension
	colorGenerator *controlexecute.DimensionColorGenerator

	// screen width
	width int
	// if true, only display failed results
	errorsOnly bool
	indent     string
}

func NewResultRenderer(status, reason string, dimensions []controlexecute.Dimension, colorGenerator *controlexecute.DimensionColorGenerator, width int, indent string) *ResultRenderer {
	return &ResultRenderer{
		status:         status,
		reason:         reason,
		dimensions:     dimensions,
		colorGenerator: colorGenerator,
		width:          width,
		errorsOnly:     viper.GetString(constants.ArgOutput) == "brief",
		indent:         indent,
	}
}

func (r ResultRenderer) Render() string {
	// in quiet mode, only render failures
	if r.errorsOnly && !helpers.StringSliceContains([]string{string(constants.ControlAlarm), string(constants.ControlError)}, r.status) {
		return ""
	}

	status := NewResultStatusRenderer(r.status)
	statusString := status.Render()
	statusWidth := helpers.PrintableLength(statusString)

	formattedIndent := fmt.Sprintf("%s", ControlColors.Indent(r.indent))
	indentWidth := helpers.PrintableLength(formattedIndent)

	// figure out how much width we have available for the  dimensions, allowing the minimum for the reason
	availableWidth := r.width - statusWidth - indentWidth

	// for now give this all to reason
	availableDimensionWidth := availableWidth - minReasonWidth
	var dimensionsString string
	var dimensionWidth int
	if availableDimensionWidth > 0 {
		dimensionsString = NewDimensionsRenderer(r.dimensions, r.colorGenerator, availableDimensionWidth).Render()
		dimensionWidth = helpers.PrintableLength(dimensionsString)
		availableWidth -= dimensionWidth
	}

	// now availableWidth is all we have - if it is not enough we need to truncate the reason
	reasonString := NewResultReasonRenderer(r.status, r.reason, availableWidth).Render()
	reasonWidth := helpers.PrintableLength(reasonString)

	// is there any room for a spacer
	availableWidth -= reasonWidth
	var spacerString string
	if availableWidth > 0 && r.dimensions != nil {
		spacerString = NewSpacerRenderer(availableWidth).Render()
	}

	// now put these all together
	str := fmt.Sprintf("%s%s%s%s%s", formattedIndent, statusString, reasonString, spacerString, dimensionsString)
	return str
}
