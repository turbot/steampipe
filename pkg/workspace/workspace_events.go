package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func (w *Workspace) PublishDashboardEvent(e dashboardevents.DashboardEvent) {
	if w.dashboardEventChan != nil {
		// send an event onto the event bus
		w.dashboardEventChan <- e
	}
}

// RegisterDashboardEventHandler starts the event handler goroutine if necessary and
// adds the event handler to our list
func (w *Workspace) RegisterDashboardEventHandler(handler dashboardevents.DashboardEventHandler) {
	// if no event channel has been created we need to start the event handler goroutine
	if w.dashboardEventChan == nil {
		w.dashboardEventChan = make(chan dashboardevents.DashboardEvent, 20)
		go w.handleDashboardEvent()
	}
	// now add the handler to our list
	w.dashboardEventHandlers = append(w.dashboardEventHandlers, handler)
}

// UnregisterDashboardEventHandlers clears all event handlers
// used when generating multiple snapshots
func (w *Workspace) UnregisterDashboardEventHandlers() {
	w.dashboardEventHandlers = nil
}

// this function is run as a goroutine to call registered event handlers for all received events
func (w *Workspace) handleDashboardEvent() {
	for {
		e := <-w.dashboardEventChan
		if e == nil {
			w.dashboardEventChan = nil
			return
		}

		for _, handler := range w.dashboardEventHandlers {
			handler(e)
		}
	}
}

func (w *Workspace) handleFileWatcherEvent(ctx context.Context, client db_common.Client, ev []fsnotify.Event) {
	prevResourceMaps, resourceMaps, err := w.reloadResourceMaps(ctx)
	if err != nil {
		// publish error event
		w.PublishDashboardEvent(&dashboardevents.WorkspaceError{Error: err})
		return
	}
	// if resources have changed, update introspection tables and prepared statements
	if !prevResourceMaps.Equals(resourceMaps) {
		res := client.RefreshSessions(context.Background())
		if res.Error != nil || len(res.Warnings) > 0 {
			fmt.Println()
			error_helpers.ShowErrorWithMessage(ctx, res.Error, "error when refreshing session data")
			error_helpers.ShowWarning(strings.Join(res.Warnings, "\n"))
		}
		if w.onFileWatcherEventMessages != nil {
			w.onFileWatcherEventMessages()
		}
	}
	w.raiseDashboardChangedEvents(resourceMaps, prevResourceMaps)
}

func (w *Workspace) reloadResourceMaps(ctx context.Context) (*modconfig.ResourceMaps, *modconfig.ResourceMaps, error) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// get the pre-load resource maps
	// NOTE: do not call GetResourceMaps - we DO NOT want to lock loadLock
	prevResourceMaps := w.Mod.ResourceMaps
	// if there is an outstanding watcher error, set prevResourceMaps to empty to force refresh
	if w.watcherError != nil {
		prevResourceMaps = modconfig.NewModResources(w.Mod)
	}

	// now reload the workspace
	err := w.loadWorkspaceMod(ctx)
	if err != nil {
		// check the existing watcher error - if we are already in an error state, do not show error
		if w.watcherError == nil {
			w.fileWatcherErrorHandler(ctx, error_helpers.PrefixError(err, "failed to reload workspace"))
		}
		// now set watcher error to new error
		w.watcherError = err

		return nil, nil, err
	}

	// clear watcher error
	w.watcherError = nil

	// reload the resource maps
	resourceMaps := w.Mod.ResourceMaps

	return prevResourceMaps, resourceMaps, nil

}

func (w *Workspace) raiseDashboardChangedEvents(resourceMaps, prevResourceMaps *modconfig.ResourceMaps) {
	event := &dashboardevents.DashboardChanged{}

	// TODO reports can we use a ResourceMaps diff function to do all of this - we are duplicating logic

	// first detect changes to existing resources and deletions
	for name, prev := range prevResourceMaps.Dashboards {
		if current, ok := resourceMaps.Dashboards[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedDashboards = append(event.ChangedDashboards, diff)
			}
		} else {
			event.DeletedDashboards = append(event.DeletedDashboards, prev)
		}
	}
	for name, prev := range prevResourceMaps.DashboardContainers {
		if current, ok := resourceMaps.DashboardContainers[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedContainers = append(event.ChangedContainers, diff)
			}
		} else {
			event.DeletedContainers = append(event.DeletedContainers, prev)
		}
	}
	for name, prev := range prevResourceMaps.DashboardCards {
		if current, ok := resourceMaps.DashboardCards[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedCards = append(event.ChangedCards, diff)
			}
		} else {
			event.DeletedCards = append(event.DeletedCards, prev)
		}
	}
	for name, prev := range prevResourceMaps.DashboardCharts {
		if current, ok := resourceMaps.DashboardCharts[name]; ok {
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
				event.ChangedBenchmarks = append(event.ChangedBenchmarks, diff)
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
	for name, prev := range prevResourceMaps.DashboardFlows {
		if current, ok := resourceMaps.DashboardFlows[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedFlows = append(event.ChangedFlows, diff)
			}
		} else {
			event.DeletedFlows = append(event.DeletedFlows, prev)
		}
	}
	for name, prev := range prevResourceMaps.DashboardGraphs {
		if current, ok := resourceMaps.DashboardGraphs[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedGraphs = append(event.ChangedGraphs, diff)
			}
		} else {
			event.DeletedGraphs = append(event.DeletedGraphs, prev)
		}
	}
	for name, prev := range prevResourceMaps.DashboardHierarchies {
		if current, ok := resourceMaps.DashboardHierarchies[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedHierarchies = append(event.ChangedHierarchies, diff)
			}
		} else {
			event.DeletedHierarchies = append(event.DeletedHierarchies, prev)
		}
	}
	for name, prev := range prevResourceMaps.DashboardImages {
		if current, ok := resourceMaps.DashboardImages[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedImages = append(event.ChangedImages, diff)
			}
		} else {
			event.DeletedImages = append(event.DeletedImages, prev)
		}
	}
	for name, prev := range prevResourceMaps.GlobalDashboardInputs {
		if current, ok := resourceMaps.GlobalDashboardInputs[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedInputs = append(event.ChangedInputs, diff)
			}
		} else {
			event.DeletedInputs = append(event.DeletedInputs, prev)
		}
	}
	for name, prevInputsForDashboard := range prevResourceMaps.DashboardInputs {
		if currentInputsForDashboard, ok := resourceMaps.DashboardInputs[name]; ok {
			for name, prev := range prevInputsForDashboard {
				if current, ok := currentInputsForDashboard[name]; ok {
					diff := prev.Diff(current)
					if diff.HasChanges() {
						event.ChangedInputs = append(event.ChangedInputs, diff)
					}
				} else {
					event.DeletedInputs = append(event.DeletedInputs, prev)
				}
			}
		} else {
			for _, prev := range prevInputsForDashboard {
				event.DeletedInputs = append(event.DeletedInputs, prev)
			}
		}
	}
	for name, prev := range prevResourceMaps.DashboardTables {
		if current, ok := resourceMaps.DashboardTables[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedTables = append(event.ChangedTables, diff)
			}
		} else {
			event.DeletedTables = append(event.DeletedTables, prev)
		}
	}
	for name, prev := range prevResourceMaps.DashboardCategories {
		if current, ok := resourceMaps.DashboardCategories[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedCategories = append(event.ChangedCategories, diff)
			}
		} else {
			event.DeletedCategories = append(event.DeletedCategories, prev)
		}
	}
	for name, prev := range prevResourceMaps.DashboardTexts {
		if current, ok := resourceMaps.DashboardTexts[name]; ok {
			diff := prev.Diff(current)
			if diff.HasChanges() {
				event.ChangedTexts = append(event.ChangedTexts, diff)
			}
		} else {
			event.DeletedTexts = append(event.DeletedTexts, prev)
		}
	}

	// now detect new resources
	for name, p := range resourceMaps.Dashboards {
		if _, ok := prevResourceMaps.Dashboards[name]; !ok {
			event.NewDashboards = append(event.NewDashboards, p)
		}
	}
	for name, p := range resourceMaps.DashboardContainers {
		if _, ok := prevResourceMaps.DashboardContainers[name]; !ok {
			event.NewContainers = append(event.NewContainers, p)
		}
	}
	for name, p := range resourceMaps.DashboardCards {
		if _, ok := prevResourceMaps.DashboardCards[name]; !ok {
			event.NewCards = append(event.NewCards, p)
		}
	}
	for name, p := range resourceMaps.DashboardCategories {
		if _, ok := prevResourceMaps.DashboardCategories[name]; !ok {
			event.NewCategories = append(event.NewCategories, p)
		}
	}
	for name, p := range resourceMaps.DashboardCharts {
		if _, ok := prevResourceMaps.DashboardCharts[name]; !ok {
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
	for name, p := range resourceMaps.DashboardFlows {
		if _, ok := prevResourceMaps.DashboardFlows[name]; !ok {
			event.NewFlows = append(event.NewFlows, p)
		}
	}
	for name, p := range resourceMaps.DashboardGraphs {
		if _, ok := prevResourceMaps.DashboardGraphs[name]; !ok {
			event.NewGraphs = append(event.NewGraphs, p)
		}
	}
	for name, p := range resourceMaps.DashboardHierarchies {
		if _, ok := prevResourceMaps.DashboardHierarchies[name]; !ok {
			event.NewHierarchies = append(event.NewHierarchies, p)
		}
	}
	for name, p := range resourceMaps.DashboardImages {
		if _, ok := prevResourceMaps.DashboardImages[name]; !ok {
			event.NewImages = append(event.NewImages, p)
		}
	}
	for name, p := range resourceMaps.GlobalDashboardInputs {
		if _, ok := prevResourceMaps.GlobalDashboardInputs[name]; !ok {
			event.NewInputs = append(event.NewInputs, p)
		}
	}

	for name, currentInputsForDashboard := range resourceMaps.DashboardInputs {
		if prevInputsForDashboard, ok := prevResourceMaps.DashboardInputs[name]; ok {
			for name, current := range currentInputsForDashboard {
				if _, ok := prevInputsForDashboard[name]; !ok {
					event.NewInputs = append(event.NewInputs, current)
				}
			}
		} else {
			// all new
			for _, current := range currentInputsForDashboard {
				event.NewInputs = append(event.NewInputs, current)
			}
		}
	}

	for name, p := range resourceMaps.DashboardTables {
		if _, ok := prevResourceMaps.DashboardTables[name]; !ok {
			event.NewTables = append(event.NewTables, p)
		}
	}
	for name, p := range resourceMaps.DashboardTexts {
		if _, ok := prevResourceMaps.DashboardTexts[name]; !ok {
			event.NewTexts = append(event.NewTexts, p)
		}
	}

	if event.HasChanges() {
		// for every changed resource, set parents as changed, up the tree
		f := func(item modconfig.ModTreeItem) (bool, error) {
			event.SetParentsChanged(item)
			return true, nil
		}
		event.WalkChangedResources(f)
		w.PublishDashboardEvent(event)
	}
}
