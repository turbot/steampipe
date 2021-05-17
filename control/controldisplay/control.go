package controldisplay

import (
	"log"
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

func (r ControlRenderer) Render() string {
	log.Println("[TRACE] begin control render")
	defer log.Println("[TRACE] end control render")

	var controlStrings []string
	// use group heading renderer to render the control title and counts
	controlHeadingRenderer := NewGroupHeadingRenderer(typehelpers.SafeString(r.run.Control.Title),
		r.run.Summary.FailedCount(),
		r.run.Summary.TotalCount(),
		r.maxFailedControls,
		r.maxTotalControls,
		r.width)
	controlStrings = append(controlStrings,
		controlHeadingRenderer.Render(),
		// newline after control heading
		"")

	// now render the results
	for _, row := range r.run.Result.Rows {
		resultRenderer := NewResultRenderer(row.Status, row.Reason, row.Dimensions, r.colorMap, r.width)
		controlStrings = append(controlStrings, resultRenderer.Render())

	}
	// newline after results
	if len(r.run.Result.Rows) > 0 {
		controlStrings = append(controlStrings, "")
	}
	return strings.Join(controlStrings, "\n")
}
