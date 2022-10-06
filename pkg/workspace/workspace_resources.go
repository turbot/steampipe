package workspace

import "github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"

func (w *Workspace) GetQuery(queryName string) (*modconfig.Query, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	if query, ok := w.Mod.ResourceMaps.LocalQueries[queryName]; ok {
		return query, true
	}
	if query, ok := w.Mod.ResourceMaps.Queries[queryName]; ok {
		return query, true
	}
	return nil, false
}

func (w *Workspace) GetControl(controlName string) (*modconfig.Control, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	if control, ok := w.Mod.ResourceMaps.LocalControls[controlName]; ok {
		return control, true
	}
	if control, ok := w.Mod.ResourceMaps.Controls[controlName]; ok {
		return control, true
	}
	return nil, false
}

// GetResourceMaps implements ResourceMapsProvider
func (w *Workspace) GetResourceMaps() *modconfig.ResourceMaps {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// if this a source snapshot workspace, create a ResourceMaps containing ONLY source snapshot paths
	if len(w.SourceSnapshots) != 0 {
		return modconfig.NewSourceSnapshotModResources(w.SourceSnapshots)
	}
	return w.Mod.ResourceMaps
}
