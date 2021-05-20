package controldisplay

import (
	"log"
	"strings"

	"github.com/turbot/steampipe/control/execute"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type GroupRenderer struct {
	group *execute.ResultGroup
	// screen width
	width             int
	maxFailedControls int
	maxTotalControls  int
	resultTree        *execute.ExecutionTree
}

func NewGroupRenderer(group *execute.ResultGroup, maxFailedControls, maxTotalControls int, resultTree *execute.ExecutionTree, width int) *GroupRenderer {
	return &GroupRenderer{
		group:             group,
		resultTree:        resultTree,
		maxFailedControls: maxFailedControls,
		maxTotalControls:  maxTotalControls,
		width:             width,
	}
}

func (r GroupRenderer) Render() string {
	log.Printf("[TRACE] begin group render '%s'\n", r.group.GroupId)
	defer log.Printf("[TRACE] end table render'%s'\n", r.group.GroupId)

	if r.group.GroupId == execute.RootResultGroupName {
		return r.renderRootResultGroup()
	}

	groupHeadingRenderer := NewGroupHeadingRenderer(
		r.group.Title,
		r.group.Summary.Status.FailedCount(),
		r.group.Summary.Status.TotalCount(),
		r.maxFailedControls,
		r.maxTotalControls,
		r.width)

	// render this group header
	tableStrings := append([]string{},
		groupHeadingRenderer.Render(),
		// newline after group
		"")

	// now render the group children, in the order they are specified
	childStrings := r.renderChildren()
	tableStrings = append(tableStrings, childStrings...)
	return strings.Join(tableStrings, "\n")
}

// for root result group, there will either be one or more groups, or one or more control runs
// there will be no order specified so just lop through them
func (r GroupRenderer) renderRootResultGroup() string {
	var resultStrings = make([]string, len(r.group.Groups)+len(r.group.ControlRuns))
	for i, group := range r.group.Groups {
		groupRenderer := NewGroupRenderer(group, r.maxFailedControls, r.maxTotalControls, r.resultTree, r.width)
		resultStrings[i] = groupRenderer.Render()
	}
	for i, run := range r.group.ControlRuns {
		controlRenderer := NewControlRenderer(run, r.maxFailedControls, r.maxTotalControls, r.resultTree.DimensionColorGenerator, r.width)
		resultStrings[i] = controlRenderer.Render()
	}
	return strings.Join(resultStrings, "\n")
}

// render the children of this group, in the order they are specified in the hcl
func (r GroupRenderer) renderChildren() []string {
	children := r.group.GroupItem.GetChildren()
	var childStrings = make([]string, len(children))

	for i, child := range children {
		if control, ok := child.(*modconfig.Control); ok {
			// get Result group with a matching name
			if run := r.group.GetControlRunByName(control.Name()); run != nil {
				controlRenderer := NewControlRenderer(run, r.maxFailedControls, r.maxTotalControls, r.resultTree.DimensionColorGenerator, r.width)
				childStrings[i] = controlRenderer.Render()
			}
		} else {
			if childGroup := r.group.GetGroupByName(child.Name()); childGroup != nil {
				groupRenderer := NewGroupRenderer(childGroup, r.maxFailedControls, r.maxTotalControls, r.resultTree, r.width)
				childStrings[i] = groupRenderer.Render()
			}
		}
	}

	return childStrings
}
