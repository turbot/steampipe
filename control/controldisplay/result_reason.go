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
func (r ResultReasonRenderer) Render() string {
	// get the color for our status
	colorFunc, ok := reasonColors[r.status]
	if !ok {
		return ""
	}
	// truncate the reason
	// reason format is
	// ": <reason> "
	// deduct 3 from length to allow for ": " and trailing space)
	availableWidth := r.width - 3
	formattedReason := fmt.Sprintf("%s", colorFunc(helpers.TruncateString(r.reason, availableWidth)))

	return fmt.Sprintf("%s %s ", colorReasonColon(":"), formattedReason)
}
