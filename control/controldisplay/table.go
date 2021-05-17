package controldisplay

import (
	"strings"

	"github.com/turbot/steampipe/control/execute"
)

type TableRenderer struct {
	resultTree *execute.ExecutionTree

	// screen width
	width             int
	maxFailedControls int
	maxTotalControls  int
}

func NewTableRenderer(resultTree *execute.ExecutionTree, width int) *TableRenderer {
	return &TableRenderer{
		resultTree:        resultTree,
		width:             width,
		maxFailedControls: resultTree.Root.Summary.Status.FailedCount(),
		maxTotalControls:  resultTree.Root.Summary.Status.TotalCount(),
	}
}

func (r TableRenderer) Render() string {
	// traverse tree
	str := r.renderResultGroup(r.resultTree.Root)
	return str
}

func (r TableRenderer) renderResultGroup(group *execute.ResultGroup) string {
	groupRenderer := NewGroupRenderer(
		group.Title,
		group.Summary.Status.FailedCount(),
		group.Summary.Status.TotalCount(),
		r.maxFailedControls,
		r.maxTotalControls,
		r.width)
	var tableStrings []string

	// do not render the root group
	if group.GroupId != execute.RootResultGroupName {
		// render this group
		tableStrings = append(tableStrings,
			groupRenderer.Render(),
			// newline after group
			"")
	}

	for _, run := range group.ControlRuns {
		controlRenderer := NewControlRenderer(run, r.maxFailedControls, r.maxTotalControls, r.resultTree.DimensionColorMap, r.width)
		tableStrings = append(tableStrings, controlRenderer.Render())
	}

	// render child groups
	for _, child := range group.Groups {
		tableStrings = append(tableStrings, r.renderResultGroup(child))
	}
	return strings.Join(tableStrings, "\n")

}
