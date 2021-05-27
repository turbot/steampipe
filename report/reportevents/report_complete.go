package reportevents

import "github.com/turbot/steampipe/report/reportinterfaces"

type ReportComplete struct {
	Report reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*ReportComplete) IsReportEvent() {}
