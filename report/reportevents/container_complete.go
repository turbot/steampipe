package reportevents

import "github.com/turbot/steampipe/report/reportinterfaces"

type ContainerComplete struct {
	Container reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*ContainerComplete) IsReportEvent() {}
