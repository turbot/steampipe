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

// GetResourceMaps implements ResourceMapsProvider
func (w *Workspace) GetResourceMaps() *modconfig.WorkspaceResourceMaps {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// this will only occur for unit tests
	if w.resourceMaps == nil {
		w.populateResourceMaps()
	}

	return w.resourceMaps
}

func (w *Workspace) populateResourceMaps() {
	queries, localQueries := w.buildQueryMap()
	controls, localControls := w.buildControlMap()
	benchmarks, localBenchmarks := w.buildBenchmarkMap()

	w.resourceMaps = &modconfig.WorkspaceResourceMaps{
		Mod:                   w.Mod,
		Mods:                  make(map[string]*modconfig.Mod),
		LocalQueries:          localQueries,
		Queries:               queries,
		Controls:              controls,
		LocalControls:         localControls,
		Benchmarks:            benchmarks,
		LocalBenchmarks:       localBenchmarks,
		Variables:             w.Variables,
		Dashboards:            w.buildDashboardMap(),
		DashboardContainers:   w.buildDashboardContainerMap(),
		DashboardCards:        w.buildDashboardCardMap(),
		DashboardCharts:       w.buildDashboardChartMap(),
		DashboardHierarchies:  w.buildDashboardHierarchyMap(),
		DashboardImages:       w.buildDashboardImageMap(),
		DashboardInputs:       w.buildDashboardInputMap(),
		GlobalDashboardInputs: w.buildGlobalDashboardInputMap(),
		DashboardTables:       w.buildDashboardTableMap(),
		DashboardTexts:        w.buildDashboardTextMap(),
	}
	w.resourceMaps.PopulateReferences()
	// if mod is not a default mod (i.e. if there is a mod.sp), add it into the resource maps
	if !w.Mod.IsDefaultMod() {
		w.resourceMaps.Mods[w.Mod.Name()] = w.Mod
	}

}

// resource map building
func (w *Workspace) buildQueryMap() (map[string]*modconfig.Query, map[string]*modconfig.Query) {
	//  build a list of long and short names for these queries
	var queryMap = make(map[string]*modconfig.Query)
	var localQueryMap = make(map[string]*modconfig.Query)

	for _, q := range w.Mod.Queries {
		localQueryMap[q.UnqualifiedName] = q
		queryMap[q.Name()] = q
	}

	// for mod dependencies, add resources keyed by long name only
	for _, mod := range w.Mods {
		for _, q := range mod.Queries {
			// if this mod is a direct dependency of the workspace mod, add it to the map _without_ a verison
			queryMap[q.Name()] = q

		}
	}
	return queryMap, localQueryMap
}

func (w *Workspace) buildControlMap() (map[string]*modconfig.Control, map[string]*modconfig.Control) {
	//  build a list of long and short names for these controls
	var controlMap = make(map[string]*modconfig.Control)
	var localControlMap = make(map[string]*modconfig.Control)

	for _, c := range w.Mod.Controls {
		localControlMap[c.UnqualifiedName] = c
		controlMap[c.Name()] = c
	}

	// for mod dependencies, add resources keyed by long name only
	for _, mod := range w.Mods {
		for _, q := range mod.Controls {
			// if this mod is a direct dependency of the workspace mod, add it to the map _without_ a verison
			controlMap[q.Name()] = q

		}
	}
	return controlMap, localControlMap
}

func (w *Workspace) buildBenchmarkMap() (map[string]*modconfig.Benchmark, map[string]*modconfig.Benchmark) {
	//  build a list of long and short names for these benchmarks
	var benchmarkMap = make(map[string]*modconfig.Benchmark)
	var localBenchmarkMap = make(map[string]*modconfig.Benchmark)

	for _, c := range w.Mod.Benchmarks {
		localBenchmarkMap[c.UnqualifiedName] = c
		localBenchmarkMap[c.Name()] = c
		benchmarkMap[c.Name()] = c
	}

	// for mod dependencies, add resources keyed by long name only
	for _, mod := range w.Mods {
		for _, q := range mod.Benchmarks {
			// if this mod is a direct dependency of the workspace mod, add it to the map _without_ a verison
			benchmarkMap[q.Name()] = q

		}
	}
	return benchmarkMap, localBenchmarkMap
}

func (w *Workspace) buildDashboardMap() map[string]*modconfig.Dashboard {
	var res = make(map[string]*modconfig.Dashboard)

	for _, d := range w.Mod.Dashboards {
		res[d.Name()] = d
	}

	for _, mod := range w.Mods {
		for _, d := range mod.Dashboards {
			res[d.Name()] = d
		}
	}
	return res
}

func (w *Workspace) buildDashboardContainerMap() map[string]*modconfig.DashboardContainer {
	var res = make(map[string]*modconfig.DashboardContainer)

	for _, c := range w.Mod.DashboardContainers {
		res[c.Name()] = c
	}

	for _, mod := range w.Mods {
		for _, c := range mod.DashboardContainers {
			res[c.Name()] = c
		}
	}
	return res
}

func (w *Workspace) buildDashboardCardMap() map[string]*modconfig.DashboardCard {
	var res = make(map[string]*modconfig.DashboardCard)

	for _, p := range w.Mod.DashboardCards {
		res[p.Name()] = p
	}

	for _, mod := range w.Mods {
		for _, p := range mod.DashboardCards {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildDashboardChartMap() map[string]*modconfig.DashboardChart {
	var res = make(map[string]*modconfig.DashboardChart)

	for _, c := range w.Mod.DashboardCharts {
		res[c.Name()] = c
	}

	for _, mod := range w.Mods {
		for _, c := range mod.DashboardCharts {
			res[c.Name()] = c
		}
	}
	return res
}

func (w *Workspace) buildDashboardHierarchyMap() map[string]*modconfig.DashboardHierarchy {
	var res = make(map[string]*modconfig.DashboardHierarchy)

	for _, p := range w.Mod.DashboardHierarchies {
		res[p.Name()] = p
	}

	for _, mod := range w.Mods {
		for _, p := range mod.DashboardHierarchies {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildDashboardImageMap() map[string]*modconfig.DashboardImage {
	var res = make(map[string]*modconfig.DashboardImage)

	for _, p := range w.Mod.DashboardImages {
		res[p.Name()] = p
	}

	for _, mod := range w.Mods {
		for _, p := range mod.DashboardImages {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildDashboardInputMap() map[string]map[string]*modconfig.DashboardInput {
	var res = make(map[string]map[string]*modconfig.DashboardInput)

	for dashboardName, dashboardInputs := range w.Mod.DashboardInputs {
		res[dashboardName] = make(map[string]*modconfig.DashboardInput)

		for _, i := range dashboardInputs {
			res[dashboardName][i.Name()] = i
		}
	}
	for _, mod := range w.Mods {
		for dashboardName, dashboardInputs := range mod.DashboardInputs {
			res[dashboardName] = make(map[string]*modconfig.DashboardInput)

			for _, i := range dashboardInputs {
				res[dashboardName][i.Name()] = i
			}
		}
	}
	return res
}

func (w *Workspace) buildGlobalDashboardInputMap() map[string]*modconfig.DashboardInput {
	var res = make(map[string]*modconfig.DashboardInput)

	for _, i := range w.Mod.GlobalDashboardInputs {
		res[i.Name()] = i
	}

	for _, mod := range w.Mods {
		for _, i := range mod.GlobalDashboardInputs {
			res[i.Name()] = i
		}
	}
	return res
}

func (w *Workspace) buildDashboardTableMap() map[string]*modconfig.DashboardTable {
	var res = make(map[string]*modconfig.DashboardTable)

	for _, c := range w.Mod.DashboardTables {
		res[c.Name()] = c
	}

	for _, mod := range w.Mods {
		for _, c := range mod.DashboardTables {
			res[c.Name()] = c
		}
	}
	return res
}

func (w *Workspace) buildDashboardTextMap() map[string]*modconfig.DashboardText {
	var res = make(map[string]*modconfig.DashboardText)

	for _, c := range w.Mod.DashboardTexts {
		res[c.Name()] = c
	}

	for _, mod := range w.Mods {
		for _, c := range mod.DashboardTexts {
			res[c.Name()] = c
		}
	}
	return res
}
