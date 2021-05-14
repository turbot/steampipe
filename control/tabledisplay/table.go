package tabledisplay

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
	var tableStrings = []string{
		// render this group
		groupRenderer.Render(),
	}
	// render results
	//for _, result:= range node.Results{
	//	// TODO
	//}
	// render child groups
	for _, child := range node.Groups {
		tableStrings = append(tableStrings, t.renderResultGroup(child))
	}
	return strings.Join(tableStrings, "\n")

}
