package controldisplay

import (
	"fmt"
	"log"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/control/execute"
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
	log.Println("[TRACE] begin error render")
	defer log.Println("[TRACE] end error render")

	status := NewResultStatusRenderer(execute.ControlError)
	statusString := status.Render()
	statusWidth := helpers.PrintableLength(statusString)
	nestingString := r.indent
	nestingWidth := helpers.PrintableLength(nestingString)

	// figure out how much width we have available for the error message
	availableWidth := r.width - statusWidth - nestingWidth
	errorMessage := helpers.TruncateString(r.error.Error(), availableWidth)
	errorString := fmt.Sprintf("%s", ControlColors.StatusError(errorMessage))

	// now put these all together
	str := fmt.Sprintf("%s%s%s", nestingString, statusString, errorString)
	return str
}
