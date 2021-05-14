package controlresult

import (
	"context"

	"github.com/turbot/steampipe/display"

	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
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
func NewRootResultGroup(includeControlPredicate func(string) bool, workspace *workspace.Workspace, rootItems ...modconfig.ControlTreeItem) *ResultGroup {
	root := &ResultGroup{
		GroupId: "root",
		Results: []*Result{},
		Groups:  []*ResultGroup{},
		Tags:    make(map[string]string),
	}
	for _, item := range rootItems {
		root.Groups = append(root.Groups, NewResultGroup(includeControlPredicate, workspace, item, root))
	}
	return root
}

// NewResultGroup creates a result group from a ControlTreeItem
func NewResultGroup(includeControlPredicate func(string) bool, workspace *workspace.Workspace, item modconfig.ControlTreeItem, parent *ResultGroup) *ResultGroup {
	group := &ResultGroup{
		GroupId:     item.Name(),
		Title:       item.GetTitle(),
		Description: item.GetDescription(),
		Tags:        item.GetTags(),
		parent:      parent,
		Results:     []*Result{},
		Groups:      []*ResultGroup{},
	}
	// add child groups for children which are benchmarks
	for _, c := range item.GetChildren() {
		if benchmark, ok := c.(*modconfig.Benchmark); ok {
			group.Groups = append(group.Groups, NewResultGroup(includeControlPredicate, workspace, benchmark, group))
		}
		if control, ok := c.(*modconfig.Control); ok {
			if includeControlPredicate(control.Name()) {
				group.ControlRuns = append(group.ControlRuns, NewControlRun(control, item, workspace))
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

func (r *ResultGroup) Execute(ctx context.Context, client *db.Client) int {
	spinner := display.ShowSpinner("")

	var errors = 0
	//totalControls := len(e.Controls)
	//pendingControls := totalControls
	//completeControls := 0
	//errorControls := 0
	//
	for _, controlRun := range r.ControlRuns {
		controlRun.Start(ctx, client)

		//p := c.Path()
		//display.UpdateSpinnerMessage(spinner, fmt.Sprintf("Running %d %s. (%d complete, %d pending, %d errors): executing \"%s\" (%s)", totalControls, utils.Pluralize("control", totalControls), completeControls, pendingControls, errorControls, typeHelpers.SafeString(c.Title), p))
		//
		//res := e.executeControl(ctx, c)
		//if res.GetRunStatus() == controlresult.ControlRunError {
		//	errorControls++
		//} else {
		//	completeControls++
		//}
		//pendingControls--
		//
		//e.ResultTree.AddResult(res)
		// TODO store errors

	}
	for _, child := range r.Groups {
		errors += child.Execute(ctx, client)
	}
	spinner.Stop()
	return errors
}
