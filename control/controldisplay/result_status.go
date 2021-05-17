package controldisplay

import (
	"fmt"
	"strings"
)

type ResultStatusRenderer struct {
	status string
}

func NewResultStatusRenderer(status string) *ResultStatusRenderer {
	return &ResultStatusRenderer{
		status: status,
	}
}

// Render returns the status
func (r ResultStatusRenderer) Render() string {
	// pad status string to fixed width
	statusString := r.paddedStatusString()

	// get the color for our status
	colorFunc, ok := statusColors[r.status]

	if !ok {
		// for unrecognised status, just return unformatted - we should be validating elsewhere
		return statusString
	}
	return fmt.Sprintf("%-5s", colorFunc(statusString))
}

// pad out status toi length of longest status string = "ERROR" - 5 chars
func (r ResultStatusRenderer) paddedStatusString() string {
	return fmt.Sprintf("%-5s", strings.ToUpper(r.status))
}
