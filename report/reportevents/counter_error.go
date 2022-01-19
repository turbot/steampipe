package reportevents

import "github.com/turbot/steampipe/report/reportinterfaces"

type CounterError struct {
	Counter reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*CounterError) IsReportEvent() {}
