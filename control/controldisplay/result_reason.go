package controldisplay

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
)

type ResultReasonRenderer struct {
	status string
	reason string
	width  int
}

func NewResultReasonRenderer(status, reason string, width int) *ResultReasonRenderer {
	return &ResultReasonRenderer{
		status: status,
		reason: reason,
		width:  width,
	}
}

// Render returns the id, truncated to the max length if necessary
func (d ResultReasonRenderer) Render() (string, int) {
	// get the color for our status
	colorFunc, ok := reasonColors[d.status]

	// truncate the reason
	truncatedReason := helpers.TruncateString(d.reason, d.width)
	length := len(truncatedReason)
	// for unrecognised status, just return unformatted - we should be validating elsewhere
	if !ok {
		return truncatedReason, length
	}
	return fmt.Sprintf("%s", colorFunc(truncatedReason)), length
}
