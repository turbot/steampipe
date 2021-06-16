package controldisplay

import (
	"fmt"
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
	lastChild         bool
	parent            *GroupRenderer
}

func NewGroupRenderer(group *execute.ResultGroup, parent *GroupRenderer, maxFailedControls, maxTotalControls int, resultTree *execute.ExecutionTree, width int) *GroupRenderer {
	r := &GroupRenderer{
		group:             group,
		parent:            parent,
		resultTree:        resultTree,
		maxFailedControls: maxFailedControls,
		maxTotalControls:  maxTotalControls,
		width:             width,
	}
	r.lastChild = r.isLastChild(group)
	return r
}

// are we the last child of our parent?
// this affects the tree rendering
func (r GroupRenderer) isLastChild(group *execute.ResultGroup) bool {
	if group.Parent == nil || group.Parent.GroupItem == nil {
		return true
	}
	siblings := group.Parent.GroupItem.GetChildren()
	// get the name of the last sibling which has controls (or is a control)
	var finalSiblingName string
	for _, s := range siblings {
		if b, ok := s.(*modconfig.Benchmark); ok {
			// find the result group for this benchmark and see if it has controls
			resultGroup := r.resultTree.Root.GetChildGroupByName(b.Name())
			// if the result group has not controls, we will not find it in the result tree
			if resultGroup == nil || resultGroup.ControlRunCount() == 0 {
				continue
			}
		}
		// store the name of this sibling
		finalSiblingName = s.Name()
	}

	res := group.GroupItem.Name() == finalSiblingName

	return res
}

// the indent for blank lines
// same as for (not last) children
func (r GroupRenderer) blankLineIndent() string {
	return r.childIndent()
}

// the indent got group heading
func (r GroupRenderer) headingIndent() string {
	// if this is the first displayed node, no indent
	if r.parent == nil || r.parent.group.GroupId == execute.RootResultGroupName {
		return ""
	}
	// as our parent for the indent for a group
	i := r.parent.childGroupIndent()
	return i
}

// the indent for child groups/controls (which are not the final child)
// include the tree '|'
func (r GroupRenderer) childIndent() string {
	return r.parentIndent() + "| "
}

// the indent for the FINAL child groups/controls
// just a space
func (r GroupRenderer) lastChildIndent() string {
	return r.parentIndent() + "  "
}

// the indent for child groups - our parent indent with the group expander "+ "
func (r GroupRenderer) childGroupIndent() string {
	return r.parentIndent() + "+ "
}

// get the indent inherited from our parent
// - this will depend on whether we are our parents last child
func (r GroupRenderer) parentIndent() string {
	if r.parent == nil || r.parent.group.GroupId == execute.RootResultGroupName {
		return ""
	}
	if r.lastChild {
		return r.parent.lastChildIndent()
	}
	return r.parent.childIndent()
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
		r.width,
		r.headingIndent())

	// render this group header
	tableStrings := append([]string{},
		groupHeadingRenderer.Render(),
		// newline after group
		fmt.Sprintf("%s", ControlColors.Indent(r.blankLineIndent())))

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
		groupRenderer := NewGroupRenderer(group, &r, r.maxFailedControls, r.maxTotalControls, r.resultTree, r.width)
		resultStrings[i] = groupRenderer.Render()
	}
	for i, run := range r.group.ControlRuns {
		controlRenderer := NewControlRenderer(run, &r)
		resultStrings[i] = controlRenderer.Render()
	}
	return strings.Join(resultStrings, "\n")
}

// render the children of this group, in the order they are specified in the hcl
func (r GroupRenderer) renderChildren() []string {
	children := r.group.GroupItem.GetChildren()
	var childStrings []string

	for _, child := range children {
		if control, ok := child.(*modconfig.Control); ok {
			// get Result group with a matching name
			if run := r.group.GetControlRunByName(control.Name()); run != nil {
				controlRenderer := NewControlRenderer(run, &r)
				childStrings = append(childStrings, controlRenderer.Render())
			}
		} else {
			if childGroup := r.group.GetGroupByName(child.Name()); childGroup != nil {
				groupRenderer := NewGroupRenderer(childGroup, &r, r.maxFailedControls, r.maxTotalControls, r.resultTree, r.width)
				childStrings = append(childStrings, groupRenderer.Render())
			}
		}
	}

	return childStrings
}
