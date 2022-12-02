package dashboardevents

import (
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type DashboardChanged struct {
	ChangedDashboards  []*modconfig.DashboardTreeItemDiffs
	ChangedContainers  []*modconfig.DashboardTreeItemDiffs
	ChangedControls    []*modconfig.DashboardTreeItemDiffs
	ChangedBenchmarks  []*modconfig.DashboardTreeItemDiffs
	ChangedCategories  []*modconfig.DashboardTreeItemDiffs
	ChangedCards       []*modconfig.DashboardTreeItemDiffs
	ChangedCharts      []*modconfig.DashboardTreeItemDiffs
	ChangedFlows       []*modconfig.DashboardTreeItemDiffs
	ChangedGraphs      []*modconfig.DashboardTreeItemDiffs
	ChangedHierarchies []*modconfig.DashboardTreeItemDiffs
	ChangedImages      []*modconfig.DashboardTreeItemDiffs
	ChangedInputs      []*modconfig.DashboardTreeItemDiffs
	ChangedTables      []*modconfig.DashboardTreeItemDiffs
	ChangedTexts       []*modconfig.DashboardTreeItemDiffs
	ChangedNodes       []*modconfig.DashboardTreeItemDiffs
	ChangedEdges       []*modconfig.DashboardTreeItemDiffs

	NewDashboards  []*modconfig.Dashboard
	NewContainers  []*modconfig.DashboardContainer
	NewControls    []*modconfig.Control
	NewBenchmarks  []*modconfig.Benchmark
	NewCards       []*modconfig.DashboardCard
	NewCategories  []*modconfig.DashboardCategory
	NewCharts      []*modconfig.DashboardChart
	NewFlows       []*modconfig.DashboardFlow
	NewGraphs      []*modconfig.DashboardGraph
	NewHierarchies []*modconfig.DashboardHierarchy
	NewImages      []*modconfig.DashboardImage
	NewInputs      []*modconfig.DashboardInput
	NewTables      []*modconfig.DashboardTable
	NewTexts       []*modconfig.DashboardText
	NewNodes       []*modconfig.DashboardNode
	NewEdges       []*modconfig.DashboardEdge

	DeletedDashboards  []*modconfig.Dashboard
	DeletedContainers  []*modconfig.DashboardContainer
	DeletedControls    []*modconfig.Control
	DeletedBenchmarks  []*modconfig.Benchmark
	DeletedCards       []*modconfig.DashboardCard
	DeletedCategories  []*modconfig.DashboardCategory
	DeletedCharts      []*modconfig.DashboardChart
	DeletedFlows       []*modconfig.DashboardFlow
	DeletedGraphs      []*modconfig.DashboardGraph
	DeletedHierarchies []*modconfig.DashboardHierarchy
	DeletedImages      []*modconfig.DashboardImage
	DeletedInputs      []*modconfig.DashboardInput
	DeletedTables      []*modconfig.DashboardTable
	DeletedTexts       []*modconfig.DashboardText
	DeletedNodes       []*modconfig.DashboardNode
	DeletedEdges       []*modconfig.DashboardEdge
}

// IsDashboardEvent implements DashboardEvent interface
func (*DashboardChanged) IsDashboardEvent() {}

func (c *DashboardChanged) HasChanges() bool {
	return len(c.ChangedDashboards)+
		len(c.ChangedContainers)+
		len(c.ChangedBenchmarks)+
		len(c.ChangedControls)+
		len(c.ChangedCards)+
		len(c.ChangedCategories)+
		len(c.ChangedCharts)+
		len(c.ChangedFlows)+
		len(c.ChangedGraphs)+
		len(c.ChangedHierarchies)+
		len(c.ChangedImages)+
		len(c.ChangedInputs)+
		len(c.ChangedTables)+
		len(c.ChangedTexts)+
		len(c.ChangedNodes)+
		len(c.ChangedEdges)+
		len(c.NewDashboards)+
		len(c.NewContainers)+
		len(c.NewBenchmarks)+
		len(c.NewControls)+
		len(c.NewCards)+
		len(c.NewCategories)+
		len(c.NewCharts)+
		len(c.NewFlows)+
		len(c.NewGraphs)+
		len(c.NewHierarchies)+
		len(c.NewImages)+
		len(c.NewInputs)+
		len(c.NewTables)+
		len(c.NewTexts)+
		len(c.NewNodes)+
		len(c.NewEdges)+
		len(c.DeletedDashboards)+
		len(c.DeletedContainers)+
		len(c.DeletedBenchmarks)+
		len(c.DeletedControls)+
		len(c.DeletedCards)+
		len(c.DeletedCategories)+
		len(c.DeletedCharts)+
		len(c.DeletedFlows)+
		len(c.DeletedGraphs)+
		len(c.DeletedHierarchies)+
		len(c.DeletedImages)+
		len(c.DeletedInputs)+
		len(c.DeletedTables)+
		len(c.DeletedTexts)+
		len(c.DeletedNodes)+
		len(c.DeletedEdges) > 0
}

func (c *DashboardChanged) WalkChangedResources(resourceFunc func(item modconfig.ModTreeItem) (bool, error)) error {
	for _, r := range c.ChangedDashboards {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedContainers {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedControls {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedCards {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedCategories {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedCharts {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedFlows {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedGraphs {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedHierarchies {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedImages {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedInputs {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedTables {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.ChangedTexts {
		if continueWalking, err := resourceFunc(r.Item); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewDashboards {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewContainers {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewControls {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewCards {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewCategories {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewCharts {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewFlows {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewGraphs {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewHierarchies {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewImages {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewInputs {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewTables {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.NewTexts {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedContainers {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedControls {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedCards {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedCategories {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedCharts {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedFlows {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedGraphs {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedHierarchies {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedImages {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedInputs {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedTables {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range c.DeletedTexts {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}

	return nil
}

func (c *DashboardChanged) SetParentsChanged(item modconfig.ModTreeItem) {
	parents := item.GetParents()
	for _, parent := range parents {
		c.AddChanged(parent)
		c.SetParentsChanged(parent)
	}
}

func (c *DashboardChanged) diffsContain(diffs []*modconfig.DashboardTreeItemDiffs, item modconfig.ModTreeItem) bool {
	for _, d := range diffs {
		if d.Item.Name() == item.Name() {
			return true
		}
	}
	return false
}

func (c *DashboardChanged) AddChanged(item modconfig.ModTreeItem) {
	diff := &modconfig.DashboardTreeItemDiffs{
		Name:              item.Name(),
		Item:              item,
		ChangedProperties: []string{"Children"},
	}
	switch item.(type) {
	case *modconfig.Dashboard:
		if !c.diffsContain(c.ChangedDashboards, item) {
			c.ChangedDashboards = append(c.ChangedDashboards, diff)
		}
	case *modconfig.DashboardContainer:
		if !c.diffsContain(c.ChangedContainers, item) {
			c.ChangedContainers = append(c.ChangedContainers, diff)
		}
	case *modconfig.Control:
		if !c.diffsContain(c.ChangedControls, item) {
			c.ChangedControls = append(c.ChangedControls, diff)
		}
	case *modconfig.Benchmark:
		if !c.diffsContain(c.ChangedBenchmarks, item) {
			c.ChangedBenchmarks = append(c.ChangedBenchmarks, diff)
		}
	case *modconfig.DashboardCard:
		if !c.diffsContain(c.ChangedCards, item) {
			c.ChangedCards = append(c.ChangedCards, diff)
		}
	case *modconfig.DashboardCategory:
		if !c.diffsContain(c.ChangedCategories, item) {
			c.ChangedCategories = append(c.ChangedCategories, diff)
		}
	case *modconfig.DashboardChart:
		if !c.diffsContain(c.ChangedCharts, item) {
			c.ChangedCharts = append(c.ChangedCharts, diff)
		}
	case *modconfig.DashboardHierarchy:
		if !c.diffsContain(c.ChangedHierarchies, item) {
			c.ChangedHierarchies = append(c.ChangedHierarchies, diff)
		}

	case *modconfig.DashboardImage:
		if !c.diffsContain(c.ChangedImages, item) {
			c.ChangedImages = append(c.ChangedImages, diff)
		}

	case *modconfig.DashboardInput:
		if !c.diffsContain(c.ChangedInputs, item) {
			c.ChangedInputs = append(c.ChangedInputs, diff)
		}

	case *modconfig.DashboardTable:
		if !c.diffsContain(c.ChangedTables, item) {
			c.ChangedTables = append(c.ChangedTables, diff)
		}
	case *modconfig.DashboardText:
		if !c.diffsContain(c.ChangedTexts, item) {
			c.ChangedTexts = append(c.ChangedTexts, diff)
		}
	}
}
