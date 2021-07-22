package controldisplay

import (
	"fmt"
	"log"
	"strings"

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
	formattedIndent := fmt.Sprintf("%s", ControlColors.Indent(r.indent))
	indentWidth := helpers.PrintableLength(formattedIndent)

	// figure out how much width we have available for the error message
	availableWidth := r.width - statusWidth - indentWidth
	errStrings := strings.Split(r.error.Error(), "\n")
	for i, str := range errStrings {
		errStrings[i] = helpers.TruncateString(str, availableWidth)
	}
	errorMessage := strings.Join(errStrings, "\n")
	errorString := fmt.Sprintf("%s", ControlColors.StatusError(errorMessage))

	// now put these all together
	str := fmt.Sprintf("%s%s%s", formattedIndent, statusString, errorString)
	return str
}
