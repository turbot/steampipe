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

// Render returns the reason, truncated to the max length if necessary
func (d ResultReasonRenderer) Render() (string, int) {
	// truncate the reason (deduct 2 from length to allow for ": ")
	availableWidth := d.width - 2
	formattedReason := helpers.TruncateString(d.reason, availableWidth)
	length := len(formattedReason) + 2
	// get the color for our status
	if colorFunc, ok := reasonColors[d.status]; ok {
		formattedReason = fmt.Sprintf("%s", colorFunc(formattedReason))
	}

	return fmt.Sprintf("%s %s", colorReasonColon(":"), formattedReason), length
}
