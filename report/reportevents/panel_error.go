package reportevents

import "github.com/turbot/steampipe/report/reportexecute"

type PanelError struct {
	Panel *reportexecute.PanelRun
}

// IsReportEvent implements ReportEvent interface
func (*PanelError) IsReportEvent() {}
