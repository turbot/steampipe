package reporteventpublisher

type ReportEvent interface {
	IsReportEvent()
}
type ReportEventHandler func(ReportEvent)
