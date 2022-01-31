package reportevents

import "github.com/turbot/steampipe/report/reportinterfaces"

type ContainerError struct {
	Container reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*ContainerError) IsReportEvent() {}
