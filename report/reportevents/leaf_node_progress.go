package reportevents

import "github.com/turbot/steampipe/report/reportinterfaces"

type LeafNodeProgress struct {
	Node reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*LeafNodeProgress) IsReportEvent() {}
