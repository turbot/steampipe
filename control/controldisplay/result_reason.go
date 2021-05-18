package controldisplay

import (
	"fmt"
	"log"

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
// NOTE: adds a trailing space
func (r ResultReasonRenderer) Render() string {
	log.Println("[TRACE] begin result reason render")
	defer log.Println("[TRACE] end result reason render")

	// get the color for our status
	colorFunc, ok := ControlColors.ReasonColors[r.status]
	if !ok {
		return ""
	}
	// truncate the reason (allow for trailing space)
	availableWidth := r.width - 1
	formattedReason := fmt.Sprintf("%s ", colorFunc(helpers.TruncateString(r.reason, availableWidth)))
	return formattedReason
}
