package tabledisplay

import (
	"fmt"
)

type ResultStatusRenderer struct {
	status string
	width  int
}

func NewResultStatusRenderer(status string) *ResultStatusRenderer {
	return &ResultStatusRenderer{
		status: status,
		// width is the length of the longest status - ERROR
		width: 6,
	}
}

// String returns the id, truncated to the max length if necessary
func (d ResultStatusRenderer) String() string {
	// get the color for our status
	colorFunc, ok := statusColors[d.status]
	if !ok {
		// for unrecognised status, just return unformatted - we should be validating elsewhere
		return fmt.Sprintf("%-6s", d.status)
	}
	return fmt.Sprintf("%-6s", colorFunc(d.status))
}
