package workspace

import "github.com/turbot/steampipe/steampipeconfig/modconfig"

func (w *Workspace) GetQuery(queryName string) (*modconfig.Query, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	if query, ok := w.LocalQueries[queryName]; ok {
		return query, true
	}
	if query, ok := w.Queries[queryName]; ok {
		return query, true
	}
	return nil, false
}

func (w *Workspace) GetControl(controlName string) (*modconfig.Control, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	if control, ok := w.LocalControls[controlName]; ok {
		return control, true
	}
	if control, ok := w.Controls[controlName]; ok {
		return control, true
	}
	return nil, false
}

// GetResourceMaps implements ResourceMapsProvider
func (w *Workspace) GetResourceMaps() *modconfig.WorkspaceResourceMaps {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	return w.resourceMaps
}

func (w *Workspace) populateResourceMaps() {
	w.resourceMaps = &modconfig.WorkspaceResourceMaps{
		Mod:                  w.Mod,
		Mods:                 make(map[string]*modconfig.Mod),
		LocalQueries:         w.LocalQueries,
		Queries:              w.Queries,
		Controls:             w.Controls,
		LocalControls:        w.LocalControls,
		Benchmarks:           w.Benchmarks,
		LocalBenchmarks:      w.LocalBenchmarks,
		Variables:            w.Variables,
		Dashboards:           w.Dashboards,
		DashboardContainers:  w.DashboardContainers,
		DashboardCards:       w.DashboardCards,
		DashboardCharts:      w.DashboardCharts,
		DashboardHierarchies: w.DashboardHierarchies,
		DashboardImages:      w.DashboardImages,
		DashboardInputs:      w.DashboardInputs,
		DashboardTables:      w.DashboardTables,
		DashboardTexts:       w.DashboardTexts,
	}
	w.resourceMaps.PopulateReferences()

	if !w.Mod.IsDefaultMod() {
		w.resourceMaps.Mods[w.Mod.Name()] = w.Mod
	}
}

// resource map building
func (w *Workspace) buildQueryMap(modMap modconfig.ModMap) (map[string]*modconfig.Query, map[string]*modconfig.Query) {
	//  build a list of long and short names for these queries
	var queryMap = make(map[string]*modconfig.Query)
	var localQueryMap = make(map[string]*modconfig.Query)

	for _, q := range w.Mod.Queries {
		localQueryMap[q.UnqualifiedName] = q
		queryMap[q.Name()] = q
	}

	// for mod dependencies, add resources keyed by long name only
	for _, mod := range modMap {
		for _, q := range mod.Queries {
			// if this mod is a direct dependency of the workspace mod, add it to the map _without_ a verison
			queryMap[q.Name()] = q

		}
	}
	return queryMap, localQueryMap
}

func (w *Workspace) buildControlMap(modMap modconfig.ModMap) (map[string]*modconfig.Control, map[string]*modconfig.Control) {
	//  build a list of long and short names for these controls
	var controlMap = make(map[string]*modconfig.Control)
	var localControlMap = make(map[string]*modconfig.Control)

	for _, c := range w.Mod.Controls {
		localControlMap[c.UnqualifiedName] = c
		controlMap[c.Name()] = c
	}

	// for mod dependencies, add resources keyed by long name only
	for _, mod := range modMap {
		for _, q := range mod.Controls {
			// if this mod is a direct dependency of the workspace mod, add it to the map _without_ a verison
			controlMap[q.Name()] = q

		}
	}
	return controlMap, localControlMap
}

func (w *Workspace) buildBenchmarkMap(modMap modconfig.ModMap) (map[string]*modconfig.Benchmark, map[string]*modconfig.Benchmark) {
	//  build a list of long and short names for these benchmarks
	var benchmarkMap = make(map[string]*modconfig.Benchmark)
	var localBenchmarkMap = make(map[string]*modconfig.Benchmark)

	for _, c := range w.Mod.Benchmarks {
		localBenchmarkMap[c.UnqualifiedName] = c
		localBenchmarkMap[c.Name()] = c
		benchmarkMap[c.Name()] = c
	}

	// for mod dependencies, add resources keyed by long name only
	for _, mod := range modMap {
		for _, q := range mod.Benchmarks {
			// if this mod is a direct dependency of the workspace mod, add it to the map _without_ a verison
			benchmarkMap[q.Name()] = q

		}
	}
	return benchmarkMap, localBenchmarkMap
}

func (w *Workspace) buildDashboardMap(modMap modconfig.ModMap) map[string]*modconfig.DashboardContainer {
	var res = make(map[string]*modconfig.DashboardContainer)

	for _, r := range w.Mod.Dashboards {
		res[r.Name()] = r
	}

	for _, mod := range modMap {
		for _, r := range mod.Dashboards {
			res[r.Name()] = r
		}
	}
	return res
}

func (w *Workspace) buildDashboardContainerMap(modMap modconfig.ModMap) map[string]*modconfig.DashboardContainer {
	var res = make(map[string]*modconfig.DashboardContainer)

	for _, c := range w.Mod.DashboardContainers {
		res[c.Name()] = c
	}

	for _, mod := range modMap {
		for _, c := range mod.DashboardContainers {
			res[c.Name()] = c
		}
	}
	return res
}

func (w *Workspace) buildDashboardCardMap(modMap modconfig.ModMap) map[string]*modconfig.DashboardCard {
	var res = make(map[string]*modconfig.DashboardCard)

	for _, p := range w.Mod.DashboardCards {
		res[p.Name()] = p
	}

	for _, mod := range modMap {
		for _, p := range mod.DashboardCards {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildDashboardChartMap(modMap modconfig.ModMap) map[string]*modconfig.DashboardChart {
	var res = make(map[string]*modconfig.DashboardChart)

	for _, c := range w.Mod.DashboardCharts {
		res[c.Name()] = c
	}

	for _, mod := range modMap {
		for _, c := range mod.DashboardCharts {
			res[c.Name()] = c
		}
	}
	return res
}

func (w *Workspace) buildDashboardHierarchyMap(modMap modconfig.ModMap) map[string]*modconfig.DashboardHierarchy {
	var res = make(map[string]*modconfig.DashboardHierarchy)

	for _, p := range w.Mod.DashboardHierarchies {
		res[p.Name()] = p
	}

	for _, mod := range modMap {
		for _, p := range mod.DashboardHierarchies {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildDashboardImageMap(modMap modconfig.ModMap) map[string]*modconfig.DashboardImage {
	var res = make(map[string]*modconfig.DashboardImage)

	for _, p := range w.Mod.DashboardImages {
		res[p.Name()] = p
	}

	for _, mod := range modMap {
		for _, p := range mod.DashboardImages {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildDashboardInputMap(modMap modconfig.ModMap) map[string]*modconfig.DashboardInput {
	var res = make(map[string]*modconfig.DashboardInput)

	for _, p := range w.Mod.DashboardInputs {
		res[p.Name()] = p
	}

	for _, mod := range modMap {
		for _, p := range mod.DashboardInputs {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildDashboardTableMap(modMap modconfig.ModMap) map[string]*modconfig.DashboardTable {
	var res = make(map[string]*modconfig.DashboardTable)

	for _, c := range w.Mod.DashboardTables {
		res[c.Name()] = c
	}

	for _, mod := range modMap {
		for _, c := range mod.DashboardTables {
			res[c.Name()] = c
		}
	}
	return res
}

func (w *Workspace) buildDashboardTextMap(modMap modconfig.ModMap) map[string]*modconfig.DashboardText {
	var res = make(map[string]*modconfig.DashboardText)

	for _, c := range w.Mod.DashboardTexts {
		res[c.Name()] = c
	}

	for _, mod := range modMap {
		for _, c := range mod.DashboardTexts {
			res[c.Name()] = c
		}
	}
	return res
}
