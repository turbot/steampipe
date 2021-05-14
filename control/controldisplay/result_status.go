package controldisplay

import (
	"fmt"
)

type ResultStatusRenderer struct {
	status string
}

func NewResultStatusRenderer(status string) *ResultStatusRenderer {
	return &ResultStatusRenderer{
		status: status,
	}
}

// Render returns the id, truncated to the max length if necessary
func (d ResultStatusRenderer) Render() (string, int) {
	// get the color for our status
	colorFunc, ok := statusColors[d.status]
	// length is the length of the longest status - ERROR
	length := 6
	if !ok {
		// for unrecognised status, just return unformatted - we should be validating elsewhere
		return fmt.Sprintf("%-6s", d.status), length
	}
	return fmt.Sprintf("%-6s", colorFunc(d.status)), length
}
