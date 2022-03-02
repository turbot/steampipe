package workspace

import "github.com/turbot/steampipe/steampipeconfig/modconfig"

func (w *Workspace) GetQuery(queryName string) (*modconfig.Query, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	if query, ok := w.resourceMaps.LocalQueries[queryName]; ok {
		return query, true
	}
	if query, ok := w.resourceMaps.Queries[queryName]; ok {
		return query, true
	}
	return nil, false
}

func (w *Workspace) GetControl(controlName string) (*modconfig.Control, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	if control, ok := w.resourceMaps.LocalControls[controlName]; ok {
		return control, true
	}
	if control, ok := w.resourceMaps.Controls[controlName]; ok {
		return control, true
	}
	return nil, false
}

// GetResourceMaps implements ModResourcesProvider
func (w *Workspace) GetResourceMaps() *modconfig.ModResources {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// this will only occur for unit tests
	if w.resourceMaps == nil {
		w.populateResourceMaps()
	}

	return w.resourceMaps
}

func (w *Workspace) populateResourceMaps() {
	dependencyResourceMaps := make([]*modconfig.ModResources, len(w.Mods))
	idx := 0
	for _, m := range w.Mods {
		dependencyResourceMaps[idx] = m.GetResourceMaps()
		idx++
	}

	w.resourceMaps = w.Mod.GetResourceMaps().Merge(dependencyResourceMaps)

	// now populate references in the resource map
	w.resourceMaps.PopulateReferences()

}
