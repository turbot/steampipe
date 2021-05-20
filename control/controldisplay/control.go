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
	width          int
	run            *execute.ControlRun
	colorGenerator *execute.DimensionColorGenerator
}

func NewControlRenderer(run *execute.ControlRun, maxFailed, maxTotal int, colorGenerator *execute.DimensionColorGenerator, width int) *ControlRenderer {
	return &ControlRenderer{
		run:               run,
		maxFailedControls: maxFailed,
		maxTotalControls:  maxTotal,
		colorGenerator:    colorGenerator,
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

	// set the severity on the heading renderer
	controlHeadingRenderer.severity = typehelpers.SafeString(r.run.Control.Severity)

	controlStrings = append(controlStrings,
		controlHeadingRenderer.Render(),
		// newline after control heading
		"")

	// if the control is in error, render an error
	if r.run.Error != nil {
		errorRenderer := NewErrorRenderer(r.run.Error, r.width)
		controlStrings = append(controlStrings,
			errorRenderer.Render(),
			// newline after error
			"")
	}

	// now render the results (if any)
	var resultStrings []string
	for _, row := range r.run.Rows {
		resultRenderer := NewResultRenderer(row.Status, row.Reason, row.Dimensions, r.colorGenerator, r.width)
		// the result renderer may not render the result - in quiet mode only failures are rendered
		if resultString := resultRenderer.Render(); resultString != "" {
			resultStrings = append(resultStrings, resultString)
		}
	}

	// newline after results
	if len(resultStrings) > 0 {
		controlStrings = append(controlStrings, resultStrings...)
		if len(r.run.Rows) > 0 || r.run.Error != nil {
			controlStrings = append(controlStrings, "")
		}
	}

	return strings.Join(controlStrings, "\n")
}
