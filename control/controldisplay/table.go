package controldisplay

import (
	"strings"

	typehelpers "github.com/turbot/go-kit/types"

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

	// render controls
	resultsRendered := 0
	for _, run := range group.ControlRuns {
		// use group renderer to render the control title and counts
		controlRenderer := NewGroupRenderer(typehelpers.SafeString(run.Control.Title),
			run.Summary.FailedCount(),
			run.Summary.TotalCount(),
			t.maxFailedControls,
			t.maxTotalControls,
			t.width)
		tableStrings = append(tableStrings,
			controlRenderer.Render(),
			// newline after group
			"")

		// now render the results
		for _, row := range run.Result.Rows {
			// TODO dimensions
			resultRenderer := NewResultRenderer(row.Status, row.Reason, t.width)
			tableStrings = append(tableStrings, resultRenderer.Render())
			resultsRendered++
		}
		// newline after results
		if resultsRendered > 0 {
			tableStrings = append(tableStrings, "")
		}
	}

	// render child groups
	for _, child := range group.Groups {
		tableStrings = append(tableStrings, t.renderResultGroup(child))
	}
	return strings.Join(tableStrings, "\n")

}
