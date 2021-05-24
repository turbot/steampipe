package reportevents

import "github.com/turbot/steampipe/report/reportexecute"

type ReportComplete struct {
	Report *reportexecute.ReportRun
}

// IsReportEvent implements ReportEvent interface
func (*ReportComplete) IsReportEvent() {}
