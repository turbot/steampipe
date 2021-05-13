package tabledisplay

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

// String returns the id, truncated to the max length if necessary
func (d ResultReasonRenderer) String() string {
	// get the color for our status
	colorFunc, ok := reasonColors[d.status]

	// truncate the reason
	truncatedReason := helpers.TruncateString(d.reason, d.width)
	// for unrecognised status, just return unformatted - we should be validating elsewhere
	if !ok {
		return truncatedReason
	}
	return fmt.Sprintf("%s", colorFunc(truncatedReason))
}
