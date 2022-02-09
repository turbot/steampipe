package reportevents

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ReportChanged struct {
	ChangedReports     []*modconfig.ReportTreeItemDiffs
	ChangedContainers  []*modconfig.ReportTreeItemDiffs
	ChangedControls    []*modconfig.ReportTreeItemDiffs
	ChangedBenchmarks  []*modconfig.ReportTreeItemDiffs
	ChangedCards       []*modconfig.ReportTreeItemDiffs
	ChangedCharts      []*modconfig.ReportTreeItemDiffs
	ChangedHierarchies []*modconfig.ReportTreeItemDiffs
	ChangedImages      []*modconfig.ReportTreeItemDiffs
	ChangedInputs      []*modconfig.ReportTreeItemDiffs
	ChangedTables      []*modconfig.ReportTreeItemDiffs
	ChangedTexts       []*modconfig.ReportTreeItemDiffs

	NewReports     []*modconfig.ReportContainer
	NewContainers  []*modconfig.ReportContainer
	NewControls    []*modconfig.Control
	NewBenchmarks  []*modconfig.Benchmark
	NewCards       []*modconfig.ReportCard
	NewCharts      []*modconfig.ReportChart
	NewHierarchies []*modconfig.ReportHierarchy
	NewImages      []*modconfig.ReportImage
	NewInputs      []*modconfig.ReportInput
	NewTables      []*modconfig.ReportTable
	NewTexts       []*modconfig.ReportText

	DeletedReports     []*modconfig.ReportContainer
	DeletedContainers  []*modconfig.ReportContainer
	DeletedControls    []*modconfig.Control
	DeletedBenchmarks  []*modconfig.Benchmark
	DeletedCards       []*modconfig.ReportCard
	DeletedCharts      []*modconfig.ReportChart
	DeletedHierarchies []*modconfig.ReportHierarchy
	DeletedImages      []*modconfig.ReportImage
	DeletedInputs      []*modconfig.ReportInput
	DeletedTables      []*modconfig.ReportTable
	DeletedTexts       []*modconfig.ReportText
}

// IsReportEvent implements ReportEvent interface
func (*ReportChanged) IsReportEvent() {}

func (c *ReportChanged) HasChanges() bool {
	return len(c.ChangedReports)+
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
		len(c.NewReports)+
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
		len(c.DeletedReports)+
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
