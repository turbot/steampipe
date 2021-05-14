package controldisplay

import (
	"strings"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/control/execute"
)

type ControlRenderer struct {
	maxFailedControls int
	maxTotalControls  int
	// screen width
	width    int
	run      *execute.ControlRun
	colorMap execute.DimensionColorMap
}

func NewControlRenderer(run *execute.ControlRun, maxFailed, maxTotal int, colorMap execute.DimensionColorMap, width int) *ControlRenderer {
	return &ControlRenderer{
		run:               run,
		maxFailedControls: maxFailed,
		maxTotalControls:  maxTotal,
		colorMap:          colorMap,
		width:             width,
	}
}

func (c ControlRenderer) Render() string {

	var controlStrings []string
	// use group renderer to render the control title and counts
	controlRenderer := NewGroupRenderer(typehelpers.SafeString(c.run.Control.Title),
		c.run.Summary.FailedCount(),
		c.run.Summary.TotalCount(),
		c.maxFailedControls,
		c.maxTotalControls,
		c.width)
	controlStrings = append(controlStrings,
		controlRenderer.Render(),
		// newline after group
		"")

	// now render the results
	for _, row := range c.run.Result.Rows {

		resultRenderer := NewResultRenderer(row.Status, row.Reason, row.Dimensions, c.colorMap, c.width)
		controlStrings = append(controlStrings, resultRenderer.Render())

	}
	// newline after results
	if len(c.run.Result.Rows) > 0 {
		controlStrings = append(controlStrings, "")
	}
	return strings.Join(controlStrings, "\n")
}
