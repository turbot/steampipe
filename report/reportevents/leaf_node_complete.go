package reportevents

import "github.com/turbot/steampipe/report/reportinterfaces"

type LeafNodeComplete struct {
	Node reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*LeafNodeComplete) IsReportEvent() {}
