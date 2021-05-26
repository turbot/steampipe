package workspace

import (
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func (w *Workspace) PublishReportEvent(e reportevents.ReportEvent) {
	for _, handler := range w.reportEventHandlers {
		handler(e)
	}
}

func (w *Workspace) RegisterReportEventHandler(handler reportevents.ReportEventHandler) {
	w.reportEventHandlers = append(w.reportEventHandlers, handler)
}

func (w *Workspace) RaiseReportChangedEvents(panels, prevPanels map[string]*modconfig.Panel, reports, prevReports map[string]*modconfig.Report) {
	event := &reportevents.ReportChanged{}

	// first detect detect changes to existing panels/reports and removed panels and reports
	for name, prevPanel := range prevPanels {
		if currentPanel, ok := panels[name]; ok {
			diff := prevPanel.Diff(currentPanel)
			if diff.HasChanges() {
				event.ChangedPanels = append(event.ChangedPanels, diff)
			}
		} else {
			event.DeletedPanels = append(event.DeletedPanels, prevPanel)
		}
	}
	for name, prevReport := range prevReports {
		if currentReport, ok := reports[name]; ok {
			diff := prevReport.Diff(currentReport)
			if diff.HasChanges() {
				event.ChangedReports = append(event.ChangedReports, diff)
			}
		} else {
			event.DeletedReports = append(event.DeletedReports, prevReport)
		}
	}
	// now detect new panels/reports
	for name, p := range panels {
		if _, ok := prevPanels[name]; !ok {
			event.NewPanels = append(event.NewPanels, p)
		}
	}
	for name, p := range reports {
		if _, ok := prevReports[name]; !ok {
			event.NewReports = append(event.NewReports, p)
		}
	}
	if event.HasChanges() {
		w.PublishReportEvent(event)
	}
}
