package reportevents

import "github.com/turbot/steampipe/report/reportinterfaces"

type LeafNodeError struct {
	Node reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*LeafNodeError) IsReportEvent() {}
