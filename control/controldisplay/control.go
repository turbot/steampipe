package controldisplay

import (
	"fmt"
	"log"
	"strings"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/control/execute"
)

type ControlRenderer struct {
	run               *execute.ControlRun
	parent            *GroupRenderer
	maxFailedControls int
	maxTotalControls  int
	// screen width
	width          int
	colorGenerator *execute.DimensionColorGenerator
	lastChild      bool
}

func NewControlRenderer(run *execute.ControlRun, parent *GroupRenderer) *ControlRenderer {
	r := &ControlRenderer{
		run:               run,
		parent:            parent,
		maxFailedControls: parent.maxFailedControls,
		maxTotalControls:  parent.maxTotalControls,
		colorGenerator:    parent.resultTree.DimensionColorGenerator,
		width:             parent.width,
	}
	r.lastChild = r.isLastChild(run)
	return r
}

// are we the last child of our parent?
// this affects the tree rendering
func (r ControlRenderer) isLastChild(run *execute.ControlRun) bool {
	if r.parent.group == nil || r.parent.group.GroupItem == nil {
		return true
	}
	siblings := r.parent.group.GroupItem.GetChildren()
	return run.Control.Name() == siblings[len(siblings)-1].Name()
}

// get the indent inherited from our parent
// - this will depend on whether we are our parents last child
func (r ControlRenderer) parentIndent() string {
	if r.lastChild {
		return r.parent.lastChildIndent()
	}
	return r.parent.childIndent()
}

// indent before first result
func (r ControlRenderer) preResultIndent() string {
	return r.parentIndent() + "| "
}

// indent before first result
func (r ControlRenderer) resultIndent() string {
	return r.parentIndent()
}

// indent after last result
func (r ControlRenderer) postResultIndent() string {
	return r.parentIndent()
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
		r.width,
		r.parent.childGroupIndent())

	// set the severity on the heading renderer
	controlHeadingRenderer.severity = typehelpers.SafeString(r.run.Control.Severity)

	// get formatted indents
	formattedPreResultIndent := fmt.Sprintf("%s", ControlColors.Indent(r.preResultIndent()))
	formattedPostResultIndent := fmt.Sprintf("%s", ControlColors.Indent(r.postResultIndent()))

	controlStrings = append(controlStrings,
		controlHeadingRenderer.Render(),
		// newline after control heading
		formattedPreResultIndent)

	// if the control is in error, render an error
	if r.run.Error != nil {
		errorRenderer := NewErrorRenderer(r.run.Error, r.width, r.parentIndent())
		controlStrings = append(controlStrings,
			errorRenderer.Render(),
			// newline after error
			formattedPostResultIndent)
	}

	// now render the results (if any)
	var resultStrings []string
	for _, row := range r.run.Rows {
		resultRenderer := NewResultRenderer(
			row.Status,
			row.Reason,
			row.Dimensions,
			r.colorGenerator,
			r.width,
			r.resultIndent())
		// the result renderer may not render the result - in quiet mode only failures are rendered
		if resultString := resultRenderer.Render(); resultString != "" {
			resultStrings = append(resultStrings, resultString)
		}
	}

	// newline after results
	if len(resultStrings) > 0 {
		controlStrings = append(controlStrings, resultStrings...)
		if len(r.run.Rows) > 0 || r.run.Error != nil {
			controlStrings = append(controlStrings, formattedPostResultIndent)
		}
	}

	return strings.Join(controlStrings, "\n")
}
