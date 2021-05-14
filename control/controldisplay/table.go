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

func (t TableRenderer) Render() string {
	// traverse tree
	str := t.renderResultGroup(t.resultTree.Root)
	return str
}

func (t TableRenderer) renderResultGroup(group *execute.ResultGroup) string {
	groupRenderer := NewGroupRenderer(
		group.Title,
		group.Summary.Status.FailedCount(),
		group.Summary.Status.TotalCount(),
		t.maxFailedControls,
		t.maxTotalControls,
		t.width)
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
		controlRenderer := NewControlRenderer(run, t.maxFailedControls, t.maxTotalControls, t.resultTree.DimensionColorMap, t.width)
		tableStrings = append(tableStrings, controlRenderer.Render())
	}

	// render child groups
	for _, child := range group.Groups {
		tableStrings = append(tableStrings, t.renderResultGroup(child))
	}
	return strings.Join(tableStrings, "\n")

}
