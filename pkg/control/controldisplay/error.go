package controldisplay

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
)

type ErrorRenderer struct {
	error error

	// screen width
	width  int
	indent string
}

func NewErrorRenderer(err error, width int, indent string) *ErrorRenderer {
	return &ErrorRenderer{
		error:  err,
		width:  width,
		indent: indent,
	}
}

func (r ErrorRenderer) Render() string {
	status := NewResultStatusRenderer(constants.ControlError)
	statusString := status.Render()
	statusWidth := helpers.PrintableLength(statusString)
	formattedIndent := fmt.Sprintf("%s", ControlColors.Indent(r.indent))
	indentWidth := helpers.PrintableLength(formattedIndent)

	// figure out how much width we have available for the error message
	availableWidth := r.width - statusWidth - indentWidth
	errorMessage := helpers.TruncateString(r.error.Error(), availableWidth)
	errorString := fmt.Sprintf("%s", ControlColors.StatusError(errorMessage))

	// now put these all together
	str := fmt.Sprintf("%s%s%s", formattedIndent, statusString, errorString)
	return str
}
