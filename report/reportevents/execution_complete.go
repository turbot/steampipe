package reportevents

import "github.com/turbot/steampipe/report/reportexecute"

type ExecutionComplete struct {
	Report *reportexecute.ReportRun
}

// IsReportEvent implements ReportEvent interface
func (*ExecutionComplete) IsReportEvent() {}
