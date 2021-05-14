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

func (t TableRenderer) renderResultGroup(node *execute.ResultGroup) string {
	groupRenderer := NewGroupRenderer(node, t.maxFailedControls, t.maxTotalControls, t.width)
	var tableStrings []string

	// do not render the root node
	if node.GroupId != execute.RootResultGroupName {
		// render this group
		tableStrings = append(tableStrings,
			groupRenderer.Render(),
			// newline after group
			"")
	}

	// render results
	resultsRendered := 0
	for _, run := range node.ControlRuns {
		for _, row := range run.Result.Rows {
			// TODO dimensions
			resultRenderer := NewResultRenderer(row.Status, row.Reason, t.width)
			tableStrings = append(tableStrings, resultRenderer.Render())
			resultsRendered++
		}
	}
	// newline after results
	if resultsRendered > 0 {
		tableStrings = append(tableStrings, "")
	}

	// render child groups
	for _, child := range node.Groups {
		tableStrings = append(tableStrings, t.renderResultGroup(child))
	}
	return strings.Join(tableStrings, "\n")

}
