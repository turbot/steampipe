package reportevents

import "github.com/turbot/steampipe/report/reportexecute"

type ReportError struct {
	Report *reportexecute.ReportRun
}

// IsReportEvent implements ReportEvent interface
func (*ReportError) IsReportEvent() {}
