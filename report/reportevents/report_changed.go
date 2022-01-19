package reportevents

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ReportChanged struct {
	ChangedReports    []*modconfig.ReportTreeItemDiffs
	ChangedContainers []*modconfig.ReportTreeItemDiffs
	ChangedTexts      []*modconfig.ReportTreeItemDiffs
	ChangedTables     []*modconfig.ReportTreeItemDiffs
	ChangedCounters   []*modconfig.ReportTreeItemDiffs
	ChangedCharts     []*modconfig.ReportTreeItemDiffs

	NewReports    []*modconfig.ReportContainer
	NewContainers []*modconfig.ReportContainer
	NewTexts      []*modconfig.ReportText
	NewTables     []*modconfig.ReportTable
	NewCounters   []*modconfig.ReportCounter
	NewCharts     []*modconfig.ReportChart

	DeletedReports    []*modconfig.ReportContainer
	DeletedContainers []*modconfig.ReportContainer
	DeletedTexts      []*modconfig.ReportText
	DeletedTables     []*modconfig.ReportTable
	DeletedCounters   []*modconfig.ReportCounter
	DeletedCharts     []*modconfig.ReportChart
}

// IsReportEvent implements ReportEvent interface
func (*ReportChanged) IsReportEvent() {}

func (c *ReportChanged) HasChanges() bool {
	return len(c.ChangedCounters)+
		len(c.ChangedReports)+
		len(c.ChangedContainers)+
		len(c.ChangedTexts)+
		len(c.ChangedTables)+
		len(c.ChangedCounters)+
		len(c.ChangedCharts)+
		len(c.NewReports)+
		len(c.NewContainers)+
		len(c.NewTexts)+
		len(c.NewTables)+
		len(c.NewCounters)+
		len(c.NewCharts)+
		len(c.DeletedReports)+
		len(c.DeletedContainers)+
		len(c.DeletedTexts)+
		len(c.DeletedTables)+
		len(c.DeletedCounters)+
		len(c.DeletedCharts) > 0
}
