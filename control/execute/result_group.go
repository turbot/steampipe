package execute

import (
	"context"

	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

const RootResultGroupName = "root_result_group"

// ResultGroup is a struct representing a grouping of control results
// It may correspond to a Benchmark, or some other arbitrary grouping
type ResultGroup struct {
	GroupId     string            `json:"group_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        map[string]string `json:"tags"`
	Summary     GroupSummary      `json:"summary"`
	Groups      []*ResultGroup    `json:"groups"`
	ControlRuns []*ControlRun     `json:"-"`
	// the control tree item associated with this group(i.e. a mod/benchmark
	GroupItem modconfig.ControlTreeItem
	parent    *ResultGroup
}

type GroupSummary struct {
	Status StatusSummary `json:"status"`
}

// NewRootResultGroup creates a ResultGroup to act as the root node of a control execution tree
func NewRootResultGroup(executionTree *ExecutionTree, rootItems ...modconfig.ControlTreeItem) *ResultGroup {
	root := &ResultGroup{
		GroupId: RootResultGroupName,
		Groups:  []*ResultGroup{},
		Tags:    make(map[string]string),
	}
	for _, item := range rootItems {
		// create new result group with root as parent
		root.Groups = append(root.Groups, NewResultGroup(executionTree, item, root))
	}
	return root
}

// NewResultGroup creates a result group from a ControlTreeItem
func NewResultGroup(executionTree *ExecutionTree, treeItem modconfig.ControlTreeItem, parent *ResultGroup) *ResultGroup {
	group := &ResultGroup{
		GroupId:     treeItem.Name(),
		Title:       treeItem.GetTitle(),
		Description: treeItem.GetDescription(),
		Tags:        treeItem.GetTags(),
		GroupItem:   treeItem,
		parent:      parent,
		Groups:      []*ResultGroup{},
	}
	// add child groups for children which are benchmarks
	for _, c := range treeItem.GetChildren() {
		if benchmark, ok := c.(*modconfig.Benchmark); ok {
			// create a new result group with 'group' as the parent
			group.Groups = append(group.Groups, NewResultGroup(executionTree, benchmark, group))
		}
		if control, ok := c.(*modconfig.Control); ok {
			if executionTree.ShouldIncludeControl(control.Name()) {
				// create new ControlRun with treeItem as the parent
				controlRun := NewControlRun(control, group, executionTree)
				// add it into the group
				group.ControlRuns = append(group.ControlRuns, controlRun)
				// also add it into the tree
				executionTree.controlRuns = append(executionTree.controlRuns, controlRun)
			}
		}
	}
	return group
}

// PopulateGroupMap mutates the passed in a map to return all child result groups
func (r *ResultGroup) PopulateGroupMap(groupMap map[string]*ResultGroup) {
	if groupMap == nil {
		groupMap = make(map[string]*ResultGroup)
	}
	// add self
	groupMap[r.GroupId] = r
	for _, g := range r.Groups {
		g.PopulateGroupMap(groupMap)
	}
}

// AddResult adds a result to the list, updates the summary status
// (this also updates the status of our parent, all the way up the tree)
func (r *ResultGroup) AddResult(run *ControlRun) {
	r.ControlRuns = append(r.ControlRuns, run)

}

func (r *ResultGroup) updateSummary(summary StatusSummary) {
	r.Summary.Status.Skip += summary.Skip
	r.Summary.Status.Alarm += summary.Alarm
	r.Summary.Status.Info += summary.Info
	r.Summary.Status.Ok += summary.Ok
	r.Summary.Status.Error += summary.Error
	if r.parent != nil {
		r.parent.updateSummary(summary)
	}
}

func (r *ResultGroup) Execute(ctx context.Context, client *db.Client) int {
	var errors = 0
	for _, controlRun := range r.ControlRuns {
		controlRun.Start(ctx, client)
	}
	for _, child := range r.Groups {
		errors += child.Execute(ctx, client)
	}
	return errors
}
