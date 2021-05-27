package workspace

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

func (w *Workspace) PublishReportEvent(e reportevents.ReportEvent) {
	for _, handler := range w.reportEventHandlers {
		handler(e)
	}
}

func (w *Workspace) RegisterReportEventHandler(handler reportevents.ReportEventHandler) {
	w.reportEventHandlers = append(w.reportEventHandlers, handler)
}

func (w *Workspace) handleFileWatcherEvent(client *db.Client, events []fsnotify.Event) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// we build a list of diffs for panels and workspaces so store the old ones
	// TODO - same for all resources??
	prevPanels := w.getPanelMap()
	prevReports := w.getReportMap()

	err := w.loadMod()
	if err != nil {
		// check the existing watcher error - if we are already in an error state, do not show error
		if w.watcherError == nil {
			fmt.Println()
			utils.ShowErrorWithMessage(err, "Failed to reload mod from file watcher")
		}
		// now set watcher error to new error
		w.watcherError = err
		// publish error event
		w.PublishReportEvent(&reportevents.WorkspaceError{Error: err})
	} else {
		// clear watcher error
		w.watcherError = nil
	}

	// todo detect differences and only refresh if necessary
	db.UpdateMetadataTables(w.GetResourceMaps(), client)

	w.raiseReportChangedEvents(w.getPanelMap(), prevPanels, w.getReportMap(), prevReports)
}

func (w *Workspace) raiseReportChangedEvents(panels, prevPanels map[string]*modconfig.Panel, reports, prevReports map[string]*modconfig.Report) {
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
