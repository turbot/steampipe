package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/turbot/steampipe/db/db_common"
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

func (w *Workspace) handleFileWatcherEvent(ctx context.Context, client db_common.Client, events []fsnotify.Event) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// store prev resources so we can detect diffs
	prevPanels := w.getPanelMap()
	prevReports := w.getReportMap()
	prevResourceMaps := w.GetResourceMaps()

	// now reload the workspace
	err := w.loadWorkspaceMod(ctx)
	if err != nil {
		// check the existing watcher error - if we are already in an error state, do not show error
		if w.watcherError == nil {
			w.fileWatcherErrorHandler(ctx, utils.PrefixError(err, "Failed to reload workspace"))
		}
		// now set watcher error to new error
		w.watcherError = err
		// publish error event
		w.PublishReportEvent(&reportevents.WorkspaceError{Error: err})
		return
	}

	// clear watcher error
	w.watcherError = nil
	resourceMaps := w.GetResourceMaps()
	// if resources have changed, update introspection tables and prepared statements
	if !prevResourceMaps.Equals(resourceMaps) {
		res := client.RefreshSessions(context.Background())
		if res.Error != nil || len(res.Warnings) > 0 {
			fmt.Println()
			utils.ShowErrorWithMessage(ctx, res.Error, "error when refreshing session data")
			utils.ShowWarning(strings.Join(res.Warnings, "\n"))
			if w.onFileWatcherEventMessages != nil {
				w.onFileWatcherEventMessages()
			}
		}
	}
	w.raiseReportChangedEvents(w.getPanelMap(), prevPanels, w.getReportMap(), prevReports)
}

func (w *Workspace) raiseReportChangedEvents(panels, prevPanels map[string]*modconfig.Panel, reports, prevReports map[string]*modconfig.ReportContainer) {
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
