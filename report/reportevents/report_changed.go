package reportevents

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ReportChanged struct {
	ChangedReports     []*modconfig.ReportTreeItemDiffs
	ChangedContainers  []*modconfig.ReportTreeItemDiffs
	ChangedCharts      []*modconfig.ReportTreeItemDiffs
	ChangedControls    []*modconfig.ReportTreeItemDiffs
	ChangedCounters    []*modconfig.ReportTreeItemDiffs
	ChangedHierarchies []*modconfig.ReportTreeItemDiffs
	ChangedImages      []*modconfig.ReportTreeItemDiffs
	ChangedTables      []*modconfig.ReportTreeItemDiffs
	ChangedTexts       []*modconfig.ReportTreeItemDiffs

	NewReports     []*modconfig.ReportContainer
	NewContainers  []*modconfig.ReportContainer
	NewCharts      []*modconfig.ReportChart
	NewControls    []*modconfig.ReportControl
	NewCounters    []*modconfig.ReportCounter
	NewHierarchies []*modconfig.ReportHierarchy
	NewImages      []*modconfig.ReportImage
	NewTables      []*modconfig.ReportTable
	NewTexts       []*modconfig.ReportText

	DeletedReports     []*modconfig.ReportContainer
	DeletedContainers  []*modconfig.ReportContainer
	DeletedCharts      []*modconfig.ReportChart
	DeletedControls    []*modconfig.ReportControl
	DeletedCounters    []*modconfig.ReportCounter
	DeletedHierarchies []*modconfig.ReportHierarchy
	DeletedImages      []*modconfig.ReportImage
	DeletedTables      []*modconfig.ReportTable
	DeletedTexts       []*modconfig.ReportText
}

// IsReportEvent implements ReportEvent interface
func (*ReportChanged) IsReportEvent() {}

func (c *ReportChanged) HasChanges() bool {
	return len(c.ChangedReports)+
		len(c.ChangedContainers)+
		len(c.ChangedCharts)+
		len(c.ChangedControls)+
		len(c.ChangedCounters)+
		len(c.ChangedHierarchies)+
		len(c.ChangedImages)+
		len(c.ChangedTables)+
		len(c.ChangedTexts)+
		len(c.NewReports)+
		len(c.NewContainers)+
		len(c.NewCharts)+
		len(c.NewControls)+
		len(c.NewCounters)+
		len(c.NewHierarchies)+
		len(c.NewImages)+
		len(c.NewTables)+
		len(c.NewTexts)+
		len(c.DeletedReports)+
		len(c.DeletedContainers)+
		len(c.DeletedCharts)+
		len(c.DeletedControls)+
		len(c.DeletedCounters)+
		len(c.DeletedHierarchies)+
		len(c.DeletedImages)+
		len(c.DeletedTables)+
		len(c.DeletedTexts) > 0
}
