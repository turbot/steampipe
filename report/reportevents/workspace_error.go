package reportevents

type WorkspaceError struct {
	Error error
}

// IsReportEvent implements ReportEvent interface
func (*WorkspaceError) IsReportEvent() {}
