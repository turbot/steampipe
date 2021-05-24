package reportevents

import "github.com/turbot/steampipe/report/reportexecute"

type PanelComplete struct {
	Panel *reportexecute.PanelRun
}

// IsReportEvent implements ReportEvent interface
func (*PanelComplete) IsReportEvent() {}
