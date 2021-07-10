package reportevents

import "github.com/turbot/steampipe/report/reportinterfaces"

type PanelComplete struct {
	Panel reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*PanelComplete) IsReportEvent() {}
