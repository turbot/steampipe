package controlexecute

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"golang.org/x/sync/semaphore"
)

const RootResultGroupName = "root_result_group"

// ResultGroup is a struct representing a grouping of control results
// It may correspond to a Benchmark, or some other arbitrary grouping
type ResultGroup struct {
	GroupId       string            `json:"name" csv:"group_id"`
	Title         string            `json:"title,omitempty" csv:"title"`
	Description   string            `json:"description,omitempty" csv:"description"`
	Tags          map[string]string `json:"tags,omitempty"`
	Documentation string            `json:"documentation,omitempty"`
	Display       string            `json:"display,omitempty"`
	Type          string            `json:"type,omitempty"`

	// the overall summary of the group
	Summary *GroupSummary `json:"summary"`
	// child result groups
	Groups []*ResultGroup `json:"-"`
	// child control runs
	ControlRuns []*ControlRun `json:"-"`
	// list of children stored as controlexecute.ExecutionTreeNode
	Children []ExecutionTreeNode                    `json:"-"`
	Severity map[string]controlstatus.StatusSummary `json:"-"`
	// "benchmark"
	NodeType string `json:"panel_type"`
	// the control tree item associated with this group(i.e. a mod/benchmark)
	GroupItem modconfig.ModTreeItem `json:"-"`
	Parent    *ResultGroup          `json:"-"`
	Duration  time.Duration         `json:"-"`
	// a list of distinct dimension keys from descendant controls
	DimensionKeys []string `json:"-"`

	// lock to prevent multiple control_runs updating this
	updateLock *sync.Mutex
}

type GroupSummary struct {
	Status   controlstatus.StatusSummary            `json:"status"`
	Severity map[string]controlstatus.StatusSummary `json:"-"`
}

func NewGroupSummary() *GroupSummary {
	return &GroupSummary{Severity: make(map[string]controlstatus.StatusSummary)}
}

// NewRootResultGroup creates a ResultGroup to act as the root node of a control execution tree
func NewRootResultGroup(ctx context.Context, executionTree *ExecutionTree, rootItem modconfig.ModTreeItem) *ResultGroup {
	root := &ResultGroup{
		GroupId:    RootResultGroupName,
		Groups:     []*ResultGroup{},
		Tags:       make(map[string]string),
		Summary:    NewGroupSummary(),
		Severity:   make(map[string]controlstatus.StatusSummary),
		updateLock: new(sync.Mutex),
		NodeType:   modconfig.BlockTypeBenchmark,
		Title:      rootItem.GetTitle(),
	}

	// if root item is a benchmark, create new result group with root as parent
	if control, ok := rootItem.(*modconfig.Control); ok {
		// if root item is a control, add control run
		executionTree.AddControl(ctx, control, root)
	} else {
		// create a result group for this item
		itemGroup := NewResultGroup(ctx, executionTree, rootItem, root)
		root.addResultGroup(itemGroup)
	}

	return root
}

// NewResultGroup creates a result group from a ModTreeItem
func NewResultGroup(ctx context.Context, executionTree *ExecutionTree, treeItem modconfig.ModTreeItem, parent *ResultGroup) *ResultGroup {
	// only show qualified group names for controls from dependent mods
	groupId := treeItem.Name()
	if mod := treeItem.GetMod(); mod != nil && mod.Name() == executionTree.Workspace.Mod.Name() {
		// TODO: We should be able to use the unqualified name for the Root Mod.
		// https://github.com/turbot/steampipe/issues/1301
		groupId = modconfig.UnqualifiedResourceName(groupId)
	}

	group := &ResultGroup{
		GroupId:     treeItem.Name(),
		Title:       treeItem.GetTitle(),
		Description: treeItem.GetDescription(),
		Tags:        treeItem.GetTags(),
		GroupItem:   treeItem,
		Parent:      parent,
		Groups:      []*ResultGroup{},
		Summary:     NewGroupSummary(),
		Severity:    make(map[string]controlstatus.StatusSummary),
		updateLock:  new(sync.Mutex),
		NodeType:    modconfig.BlockTypeBenchmark,
	}

	// populate additional properties (this avoids adding GetDocumentation, GetDisplay and GetType to all ModTreeItems)
	switch t := treeItem.(type) {
	case *modconfig.Benchmark:
		group.Documentation = t.GetDocumentation()
		group.Display = t.GetDisplay()
		group.Type = t.GetType()
	case *modconfig.Control:
		group.Documentation = t.GetDocumentation()
		group.Display = t.GetDisplay()
		group.Type = t.GetType()
	}
	// add child groups for children which are benchmarks
	for _, c := range treeItem.GetChildren() {
		if benchmark, ok := c.(*modconfig.Benchmark); ok {
			// create a result group for this item
			benchmarkGroup := NewResultGroup(ctx, executionTree, benchmark, group)
			// if the group has any control runs, add to tree
			if benchmarkGroup.ControlRunCount() > 0 {
				// create a new result group with 'group' as the parent
				group.addResultGroup(benchmarkGroup)
			}
		}
		if control, ok := c.(*modconfig.Control); ok {
			executionTree.AddControl(ctx, control, group)
		}
	}

	return group
}

func (r *ResultGroup) AllTagKeys() []string {
	tags := []string{}
	for k := range r.Tags {
		tags = append(tags, k)
	}
	for _, child := range r.Groups {
		tags = append(tags, child.AllTagKeys()...)
	}
	for _, run := range r.ControlRuns {
		for k := range run.Control.Tags {
			tags = append(tags, k)
		}
	}
	tags = helpers.StringSliceDistinct(tags)
	sort.Strings(tags)
	return tags
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

// IsSnapshotPanel implements SnapshotPanel
func (*ResultGroup) IsSnapshotPanel() {}

// IsExecutionTreeNode implements ExecutionTreeNode
func (*ResultGroup) IsExecutionTreeNode() {}

// GetChildren implements ExecutionTreeNode
func (r *ResultGroup) GetChildren() []ExecutionTreeNode { return r.Children }

// GetName implements ExecutionTreeNode
func (r *ResultGroup) GetName() string { return r.GroupId }

// AsTreeNode implements ExecutionTreeNode
func (r *ResultGroup) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	res := &dashboardtypes.SnapshotTreeNode{
		Name:     r.GroupId,
		Children: make([]*dashboardtypes.SnapshotTreeNode, len(r.Children)),
		NodeType: r.NodeType,
	}
	for i, c := range r.Children {
		res.Children[i] = c.AsTreeNode()
	}
	return res
}

// add result group into our list, and also add a tree node into our child list
func (r *ResultGroup) addResultGroup(group *ResultGroup) {
	r.Groups = append(r.Groups, group)
	r.Children = append(r.Children, group)
}

// add control into our list, and also add a tree node into our child list
func (r *ResultGroup) addControl(controlRun *ControlRun) {
	r.ControlRuns = append(r.ControlRuns, controlRun)
	r.Children = append(r.Children, controlRun)
}

func (r *ResultGroup) addDimensionKeys(keys ...string) {
	r.updateLock.Lock()
	defer r.updateLock.Unlock()
	r.DimensionKeys = append(r.DimensionKeys, keys...)
	if r.Parent != nil {
		r.Parent.addDimensionKeys(keys...)
	}
	r.DimensionKeys = helpers.StringSliceDistinct(r.DimensionKeys)
	sort.Strings(r.DimensionKeys)
}

func (r *ResultGroup) addDuration(d time.Duration) {
	r.updateLock.Lock()
	defer r.updateLock.Unlock()

	r.Duration += d.Round(time.Millisecond)
	if r.Parent != nil {
		r.Parent.addDuration(d.Round(time.Millisecond))
	}
}

func (r *ResultGroup) updateSummary(summary *controlstatus.StatusSummary) {
	r.updateLock.Lock()
	defer r.updateLock.Unlock()

	r.Summary.Status.Skip += summary.Skip
	r.Summary.Status.Alarm += summary.Alarm
	r.Summary.Status.Info += summary.Info
	r.Summary.Status.Ok += summary.Ok
	r.Summary.Status.Error += summary.Error

	if r.Parent != nil {
		r.Parent.updateSummary(summary)
	}
}

func (r *ResultGroup) updateSeverityCounts(severity string, summary *controlstatus.StatusSummary) {
	r.updateLock.Lock()
	defer r.updateLock.Unlock()

	val, exists := r.Severity[severity]
	if !exists {
		val = controlstatus.StatusSummary{}
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

func (r *ResultGroup) execute(ctx context.Context, client db_common.Client, parallelismLock *semaphore.Weighted) {
	log.Printf("[TRACE] begin ResultGroup.Execute: %s\n", r.GroupId)
	defer log.Printf("[TRACE] end ResultGroup.Execute: %s\n", r.GroupId)

	for _, controlRun := range r.ControlRuns {
		if utils.IsContextCancelled(ctx) {
			controlRun.setError(ctx, ctx.Err())
			continue
		}

		if viper.GetBool(constants.ArgDryRun) {
			controlRun.skip(ctx)
			continue
		}

		err := parallelismLock.Acquire(ctx, 1)
		if err != nil {
			controlRun.setError(ctx, err)
			continue
		}

		go executeRun(ctx, controlRun, parallelismLock, client)
	}
	for _, child := range r.Groups {
		child.execute(ctx, client, parallelismLock)
	}
}

func executeRun(ctx context.Context, run *ControlRun, parallelismLock *semaphore.Weighted, client db_common.Client) {
	defer func() {
		if r := recover(); r != nil {
			// if the Execute panic'ed, set it as an error
			run.setError(ctx, helpers.ToError(r))
		}
		// Release in defer, so that we don't retain the lock even if there's a panic inside
		parallelismLock.Release(1)
	}()

	run.execute(ctx, client)

}
