package reportevents

import "github.com/turbot/steampipe/report/reportinterfaces"

type PanelError struct {
	Panel reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*PanelError) IsReportEvent() {}
