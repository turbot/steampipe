package reportevents

import "github.com/turbot/steampipe/report/reportinterfaces"

type CounterComplete struct {
	Counter reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*CounterComplete) IsReportEvent() {}
