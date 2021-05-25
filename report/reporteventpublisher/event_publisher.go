package reporteventpublisher

type ReportEventPublisher struct {
	reportEventHandlers []ReportEventHandler
}

func (w *ReportEventPublisher) Register(handler ReportEventHandler) {
	w.reportEventHandlers = append(w.reportEventHandlers, handler)
}

func (w *ReportEventPublisher) Publish(e ReportEvent) {
	for _, handler := range w.reportEventHandlers {
		handler(e)
	}
}
