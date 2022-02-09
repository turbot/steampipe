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

	// get the pre-load resource maps
	// NOTE: do not call GetResourceMaps - we DO NOT want to lock loadLock
	prevResourceMaps := w.resourceMaps
	// if there is an outsanding watcher error, set prevResourceMaps to empty to force refresh
	if w.watcherError != nil {
		prevResourceMaps = modconfig.NewWorkspaceResourceMaps(w.Mod)
	}

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

	// reload the resource maps
	resourceMaps := w.resourceMaps

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
	for name, prev := range prevResourceMaps.ReportCards {
		if current, ok := resourceMaps.ReportCards[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedCards = append(event.ChangedCards, diff)
			}
		} else {
			event.DeletedCards = append(event.DeletedCards, prev)
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
	for name, prev := range prevResourceMaps.Benchmarks {
		if current, ok := resourceMaps.Benchmarks[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedControls = append(event.ChangedBenchmarks, diff)
			}
		} else {
			event.DeletedBenchmarks = append(event.DeletedBenchmarks, prev)
		}
	}
	for name, prev := range prevResourceMaps.Controls {
		if current, ok := resourceMaps.Controls[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedControls = append(event.ChangedControls, diff)
			}
		} else {
			event.DeletedControls = append(event.DeletedControls, prev)
		}
	}
	for name, prev := range prevResourceMaps.ReportHierarchies {
		if current, ok := resourceMaps.ReportHierarchies[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedHierarchies = append(event.ChangedHierarchies, diff)
			}
		} else {
			event.DeletedHierarchies = append(event.DeletedHierarchies, prev)
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
	for name, prev := range prevResourceMaps.ReportInputs {
		if current, ok := resourceMaps.ReportInputs[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedInputs = append(event.ChangedInputs, diff)
			}
		} else {
			event.DeletedInputs = append(event.DeletedInputs, prev)
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
	for name, p := range resourceMaps.ReportCards {
		if _, ok := prevResourceMaps.ReportCards[name]; !ok {
			event.NewCards = append(event.NewCards, p)
		}
	}
	for name, p := range resourceMaps.ReportCharts {
		if _, ok := prevResourceMaps.ReportCharts[name]; !ok {
			event.NewCharts = append(event.NewCharts, p)
		}
	}
	for name, p := range resourceMaps.Benchmarks {
		if _, ok := prevResourceMaps.Benchmarks[name]; !ok {
			event.NewBenchmarks = append(event.NewBenchmarks, p)
		}
	}
	for name, p := range resourceMaps.Controls {
		if _, ok := prevResourceMaps.Controls[name]; !ok {
			event.NewControls = append(event.NewControls, p)
		}
	}
	for name, p := range resourceMaps.ReportHierarchies {
		if _, ok := prevResourceMaps.ReportHierarchies[name]; !ok {
			event.NewHierarchies = append(event.NewHierarchies, p)
		}
	}
	for name, p := range resourceMaps.ReportImages {
		if _, ok := prevResourceMaps.ReportImages[name]; !ok {
			event.NewImages = append(event.NewImages, p)
		}
	}
	for name, p := range resourceMaps.ReportInputs {
		if _, ok := prevResourceMaps.ReportInputs[name]; !ok {
			event.NewInputs = append(event.NewInputs, p)
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
		w.PublishReportEvent(event)
	}
}
