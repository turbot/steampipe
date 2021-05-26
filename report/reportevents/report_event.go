package reportevents

type ReportEvent interface {
	IsReportEvent()
}
type ReportEventHandler func(ReportEvent)
