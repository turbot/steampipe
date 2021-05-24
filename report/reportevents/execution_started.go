package reportevents

import "github.com/turbot/steampipe/report/reportexecute"

type ExecutionStarted struct {
	Report *reportexecute.ReportRun
}

// IsReportEvent implements ReportEvent interface
func (*ExecutionStarted) IsReportEvent() {}
