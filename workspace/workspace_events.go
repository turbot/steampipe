package workspace

import (
	"github.com/turbot/steampipe/report/reporteventpublisher"
)

func (w *Workspace) PublishReportEvent(e reporteventpublisher.ReportEvent) {
	w.reportEventPublisher.Publish(e)
}

func (w *Workspace) RegisterReportEventHandler(handler reporteventpublisher.ReportEventHandler) {
	w.reportEventPublisher.Register(handler)
}
