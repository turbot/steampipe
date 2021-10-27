package controlexecute

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/turbot/steampipe/db/db_common"
	"golang.org/x/sync/semaphore"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

const RootResultGroupName = "root_result_group"

// ResultGroup is a struct representing a grouping of control results
// It may correspond to a Benchmark, or some other arbitrary grouping
type ResultGroup struct {
	GroupId     string            `json:"group_id" csv:"group_id"`
	Title       string            `json:"title" csv:"title"`
	Description string            `json:"description" csv:"description"`
	Tags        map[string]string `json:"tags"`
	Summary     *GroupSummary     `json:"summary"`
	Groups      []*ResultGroup    `json:"groups"`
	ControlRuns []*ControlRun     `json:"controls"`

	Severity map[string]StatusSummary `json:"-"`

	summaryUpdateLock *sync.Mutex

	// the control tree item associated with this group(i.e. a mod/benchmark)
	GroupItem modconfig.ModTreeItem `json:"-"`
	Parent    *ResultGroup          `json:"-"`
	Duration  time.Duration         `json:"-"`
}

type GroupSummary struct {
	Status   StatusSummary            `json:"status"`
	Severity map[string]StatusSummary `json:"-"`
}

func NewGroupSummary() *GroupSummary {
	return &GroupSummary{Severity: make(map[string]StatusSummary)}
}

// NewRootResultGroup creates a ResultGroup to act as the root node of a control execution tree
func NewRootResultGroup(executionTree *ExecutionTree, rootItems ...modconfig.ModTreeItem) *ResultGroup {
	root := &ResultGroup{
		GroupId:           RootResultGroupName,
		Groups:            []*ResultGroup{},
		Tags:              make(map[string]string),
		Summary:           NewGroupSummary(),
		summaryUpdateLock: new(sync.Mutex),
	}
	for _, item := range rootItems {
		// if root item is a benchmark, create new result group with root as parent
		if control, ok := item.(*modconfig.Control); ok {
			// if root item is a control, add control run
			executionTree.AddControl(control, root)
		} else {
			// create a result group for this item
			itemGroup := NewResultGroup(executionTree, item, root)
			root.Groups = append(root.Groups, itemGroup)
		}
	}
	return root
}

// NewResultGroup creates a result group from a ModTreeItem
func NewResultGroup(executionTree *ExecutionTree, treeItem modconfig.ModTreeItem, parent *ResultGroup) *ResultGroup {
	group := &ResultGroup{
		GroupId:           treeItem.Name(),
		Title:             treeItem.GetTitle(),
		Description:       treeItem.GetDescription(),
		Tags:              treeItem.GetTags(),
		GroupItem:         treeItem,
		Parent:            parent,
		Groups:            []*ResultGroup{},
		Summary:           NewGroupSummary(),
		summaryUpdateLock: new(sync.Mutex),
	}
	// add child groups for children which are benchmarks
	for _, c := range treeItem.GetChildren() {
		if benchmark, ok := c.(*modconfig.Benchmark); ok {
			// create a result group for this item
			benchmarkGroup := NewResultGroup(executionTree, benchmark, group)
			// if the group has any control runs, add to tree
			if benchmarkGroup.ControlRunCount() > 0 {
				// create a new result group with 'group' as the parent
				group.Groups = append(group.Groups, benchmarkGroup)
			}
		}
		if control, ok := c.(*modconfig.Control); ok {
			executionTree.AddControl(control, group)
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
	r.summaryUpdateLock.Lock()
	defer r.summaryUpdateLock.Unlock()

	r.Summary.Status.Skip += summary.Skip
	r.Summary.Status.Alarm += summary.Alarm
	r.Summary.Status.Info += summary.Info
	r.Summary.Status.Ok += summary.Ok
	r.Summary.Status.Error += summary.Error
	if r.Parent != nil {
		r.Parent.updateSummary(summary)
	}
}

func (r *ResultGroup) updateSeverityCounts(severity string, summary StatusSummary) {
	r.summaryUpdateLock.Lock()
	defer r.summaryUpdateLock.Unlock()

	if r.Severity == nil {
		r.Severity = make(map[string]StatusSummary)
	}
	val, exists := r.Severity[severity]
	if !exists {
		val = StatusSummary{}
	}
	val.Alarm += summary.Alarm
	val.Error += summary.Error
	val.Info += summary.Info
	val.Ok += summary.Ok
	val.Skip += summary.Skip

	r.Summary.Severity[severity] = val
	if r.Parent != nil {
		r.Parent.updateSeverityCounts(severity, summary)
	}
}

func (r *ResultGroup) Execute(ctx context.Context, client db_common.Client, parallelismLock *semaphore.Weighted) int {
	log.Printf("[TRACE] begin ResultGroup.Execute: %s\n", r.GroupId)
	defer log.Printf("[TRACE] end ResultGroup.Execute: %s\n", r.GroupId)

	// TODO consider executing in order specified in hcl?
	// it may not matter, as we display results in order
	// it is only an issue if there are dependencies, in which case we must run in dependency order

	startTime := time.Now()

	var failures int = 0

	for _, controlRun := range r.ControlRuns {
		if ctx.Err() != nil {
			controlRun.SetError(ctx.Err())
			continue
		}

		if viper.GetBool(constants.ArgDryRun) {
			controlRun.Skip()
			continue
		}

		err := parallelismLock.Acquire(ctx, 1)
		if err != nil {
			controlRun.SetError(err)
			continue
		}

		go func(run *ControlRun) {
			defer func() {
				// Release in defer, so that we don't retain the lock even if there's a panic inside
				parallelismLock.Release(1)
			}()
			run.Execute(ctx, client)

			run.group.summaryUpdateLock.Lock()
			failures += (run.Summary.Alarm + run.Summary.Error)
			run.group.summaryUpdateLock.Unlock()
		}(controlRun)
	}
	for _, child := range r.Groups {
		failures += child.Execute(ctx, client, parallelismLock)
	}

	r.Duration = time.Since(startTime)

	return failures
}

// GetGroupByName finds an immediate child ResultGroup with a specific name
func (r *ResultGroup) GetGroupByName(name string) *ResultGroup {
	for _, group := range r.Groups {
		if group.GroupId == name {
			return group
		}
	}
	return nil
}

// GetChildGroupByName finds a nested child ResultGroup with a specific name
func (r *ResultGroup) GetChildGroupByName(name string) *ResultGroup {
	for _, group := range r.Groups {
		if group.GroupId == name {
			return group
		}
		if child := group.GetChildGroupByName(name); child != nil {
			return child
		}
	}
	return nil
}

// GetControlRunByName finds a child ControlRun with a specific control name
func (r *ResultGroup) GetControlRunByName(name string) *ControlRun {
	for _, run := range r.ControlRuns {
		if run.Control.Name() == name {
			return run
		}
	}
	return nil
}

func (r *ResultGroup) ControlRunCount() int {
	count := len(r.ControlRuns)
	for _, g := range r.Groups {
		count += g.ControlRunCount()
	}
	return count
}
