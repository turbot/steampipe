package workspace

import (
	"context"
	"fmt"
	"log"
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

	// TODO KAI THINK ABOUT LOCKING
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
	w.raiseReportChangedEvents(resourceMaps, prevResourceMaps)
}

func (w *Workspace) raiseReportChangedEvents(resourceMaps, prevResourceMaps *modconfig.WorkspaceResourceMaps) {
	event := &reportevents.ReportChanged{}

	// first detect changes to existing resources and deletions
	for name, prev := range prevResourceMaps.Reports {
		if current, ok := resourceMaps.Reports[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedReports = append(event.ChangedReports, diff)
			}
		} else {
			event.DeletedReports = append(event.DeletedReports, prev)
		}
	}
	for name, prev := range prevResourceMaps.ReportContainers {
		if current, ok := resourceMaps.ReportContainers[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedContainers = append(event.ChangedContainers, diff)
			}
		} else {
			event.DeletedContainers = append(event.DeletedContainers, prev)
		}
	}
	for name, prev := range prevResourceMaps.ReportCharts {
		if current, ok := resourceMaps.ReportCharts[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedCharts = append(event.ChangedCharts, diff)
			}
		} else {
			event.DeletedCharts = append(event.DeletedCharts, prev)
		}
	}
	for name, prev := range prevResourceMaps.ReportCounters {
		if current, ok := resourceMaps.ReportCounters[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedCounters = append(event.ChangedCounters, diff)
			}
		} else {
			event.DeletedCounters = append(event.DeletedCounters, prev)
		}
	}
	for name, prev := range prevResourceMaps.ReportImages {
		if current, ok := resourceMaps.ReportImages[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedImages = append(event.ChangedImages, diff)
			}
		} else {
			event.DeletedImages = append(event.DeletedImages, prev)
		}
	}
	for name, prev := range prevResourceMaps.ReportTables {
		if current, ok := resourceMaps.ReportTables[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedTables = append(event.ChangedTables, diff)
			}
		} else {
			event.DeletedTables = append(event.DeletedTables, prev)
		}
	}
	for name, prev := range prevResourceMaps.ReportTexts {
		if current, ok := resourceMaps.ReportTexts[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedTexts = append(event.ChangedTexts, diff)
			}
		} else {
			event.DeletedTexts = append(event.DeletedTexts, prev)
		}
	}

	// now detect new resources
	for name, p := range resourceMaps.Reports {
		if _, ok := prevResourceMaps.Reports[name]; !ok {
			event.NewReports = append(event.NewReports, p)
		}
	}
	for name, p := range resourceMaps.ReportContainers {
		if _, ok := prevResourceMaps.ReportContainers[name]; !ok {
			event.NewContainers = append(event.NewContainers, p)
		}
	}
	for name, p := range resourceMaps.ReportCharts {
		if _, ok := prevResourceMaps.ReportCharts[name]; !ok {
			event.NewCharts = append(event.NewCharts, p)
		}
	}
	for name, p := range resourceMaps.ReportCounters {
		if _, ok := prevResourceMaps.ReportCounters[name]; !ok {
			event.NewCounters = append(event.NewCounters, p)
		}
	}
	for name, p := range resourceMaps.ReportImages {
		if _, ok := prevResourceMaps.ReportImages[name]; !ok {
			event.NewImages = append(event.NewImages, p)
		}
	}
	for name, p := range resourceMaps.ReportTables {
		if _, ok := prevResourceMaps.ReportTables[name]; !ok {
			event.NewTables = append(event.NewTables, p)
		}
	}
	for name, p := range resourceMaps.ReportTexts {
		if _, ok := prevResourceMaps.ReportTexts[name]; !ok {
			event.NewTexts = append(event.NewTexts, p)
		}
	}

	if event.HasChanges() {
		log.Printf("[WARN] **************** ReportChanged EVENT ***************\n")
		w.PublishReportEvent(event)
	}
}
