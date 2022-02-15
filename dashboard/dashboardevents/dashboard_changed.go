package dashboardevents

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type DashboardChanged struct {
	ChangedDashboards []*modconfig.DashboardTreeItemDiffs
	ChangedContainers []*modconfig.DashboardTreeItemDiffs
	ChangedControls    []*modconfig.DashboardTreeItemDiffs
	ChangedBenchmarks  []*modconfig.DashboardTreeItemDiffs
	ChangedCards       []*modconfig.DashboardTreeItemDiffs
	ChangedCharts      []*modconfig.DashboardTreeItemDiffs
	ChangedHierarchies []*modconfig.DashboardTreeItemDiffs
	ChangedImages      []*modconfig.DashboardTreeItemDiffs
	ChangedInputs      []*modconfig.DashboardTreeItemDiffs
	ChangedTables      []*modconfig.DashboardTreeItemDiffs
	ChangedTexts       []*modconfig.DashboardTreeItemDiffs

	NewDashboards []*modconfig.DashboardContainer
	NewContainers []*modconfig.DashboardContainer
	NewControls    []*modconfig.Control
	NewBenchmarks  []*modconfig.Benchmark
	NewCards       []*modconfig.DashboardCard
	NewCharts      []*modconfig.DashboardChart
	NewHierarchies []*modconfig.DashboardHierarchy
	NewImages      []*modconfig.DashboardImage
	NewInputs      []*modconfig.DashboardInput
	NewTables      []*modconfig.DashboardTable
	NewTexts       []*modconfig.DashboardText

	DeletedDashboards []*modconfig.DashboardContainer
	DeletedContainers []*modconfig.DashboardContainer
	DeletedControls    []*modconfig.Control
	DeletedBenchmarks  []*modconfig.Benchmark
	DeletedCards       []*modconfig.DashboardCard
	DeletedCharts      []*modconfig.DashboardChart
	DeletedHierarchies []*modconfig.DashboardHierarchy
	DeletedImages      []*modconfig.DashboardImage
	DeletedInputs      []*modconfig.DashboardInput
	DeletedTables      []*modconfig.DashboardTable
	DeletedTexts       []*modconfig.DashboardText
}

// IsDashboardEvent implements DashboardEvent interface
func (*DashboardChanged) IsDashboardEvent() {}

func (c *DashboardChanged) HasChanges() bool {
	return len(c.ChangedDashboards)+
		len(c.ChangedContainers)+
		len(c.ChangedBenchmarks)+
		len(c.ChangedControls)+
		len(c.ChangedCards)+
		len(c.ChangedCharts)+
		len(c.ChangedHierarchies)+
		len(c.ChangedImages)+
		len(c.ChangedInputs)+
		len(c.ChangedTables)+
		len(c.ChangedTexts)+
		len(c.NewDashboards)+
		len(c.NewContainers)+
		len(c.NewBenchmarks)+
		len(c.NewControls)+
		len(c.NewCards)+
		len(c.NewCharts)+
		len(c.NewHierarchies)+
		len(c.NewImages)+
		len(c.NewInputs)+
		len(c.NewTables)+
		len(c.NewTexts)+
		len(c.DeletedDashboards)+
		len(c.DeletedContainers)+
		len(c.DeletedBenchmarks)+
		len(c.DeletedControls)+
		len(c.DeletedCards)+
		len(c.DeletedCharts)+
		len(c.DeletedHierarchies)+
		len(c.DeletedImages)+
		len(c.DeletedInputs)+
		len(c.DeletedTables)+
		len(c.DeletedTexts) > 0
}
