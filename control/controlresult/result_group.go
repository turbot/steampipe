package controlresult

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

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
	Results     []*Result         `json:"controls"`

	parent *ResultGroup
}

type GroupSummary struct {
	Status StatusSummary `json:"status"`
}

// NewRootResultGroup creates a ResultGroup to act as the root node of a control execution tree
func NewRootResultGroup(rootItems []modconfig.ControlTreeItem) *ResultGroup {
	root := &ResultGroup{
		GroupId: "root",
	}
	for _, item := range rootItems {
		root.Groups = append(root.Groups, NewResultGroup(item, root))
	}
	return root
}

// NewResultGroup creates a result group from a ControlTreeItem
func NewResultGroup(item modconfig.ControlTreeItem, parent *ResultGroup) *ResultGroup {
	group := &ResultGroup{
		GroupId:     item.Name(),
		Title:       item.GetTitle(),
		Description: item.GetDescription(),
		Tags:        item.GetTags(),
		parent:      parent,
	}
	// add child groups for children which are benchmarks
	for _, c := range item.GetChildren() {
		if benchmark, ok := c.(*modconfig.Benchmark); ok {
			group.Groups = append(group.Groups, NewResultGroup(benchmark, group))
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
	r.Results = append(r.Results, run.Result)
	r.updateSummary(run.Summary)
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
