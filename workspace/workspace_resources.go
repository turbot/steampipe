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
		Mod:               w.Mod,
		Mods:              make(map[string]*modconfig.Mod),
		LocalQueries:      w.LocalQueries,
		Queries:           w.Queries,
		Controls:          w.Controls,
		LocalControls:     w.LocalControls,
		Benchmarks:        w.Benchmarks,
		LocalBenchmarks:   w.LocalBenchmarks,
		Variables:         w.Variables,
		Reports:           w.Reports,
		ReportContainers:  w.ReportContainers,
		ReportCharts:      w.ReportCharts,
		ReportCounters:    w.ReportCounters,
		ReportHierarchies: w.ReportHierarchies,
		ReportImages:      w.ReportImages,
		ReportInputs:      w.ReportInputs,
		ReportTables:      w.ReportTables,
		ReportTexts:       w.ReportTexts,
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

func (w *Workspace) buildReportMap(modMap modconfig.ModMap) map[string]*modconfig.ReportContainer {
	var res = make(map[string]*modconfig.ReportContainer)

	for _, r := range w.Mod.Reports {
		res[r.Name()] = r
	}

	for _, mod := range modMap {
		for _, r := range mod.Reports {
			res[r.Name()] = r
		}
	}
	return res
}

func (w *Workspace) buildReportContainerMap(modMap modconfig.ModMap) map[string]*modconfig.ReportContainer {
	var res = make(map[string]*modconfig.ReportContainer)

	for _, c := range w.Mod.ReportContainers {
		res[c.Name()] = c
	}

	for _, mod := range modMap {
		for _, c := range mod.ReportContainers {
			res[c.Name()] = c
		}
	}
	return res
}

func (w *Workspace) buildReportChartMap(modMap modconfig.ModMap) map[string]*modconfig.ReportChart {
	var res = make(map[string]*modconfig.ReportChart)

	for _, c := range w.Mod.ReportCharts {
		res[c.Name()] = c
	}

	for _, mod := range modMap {
		for _, c := range mod.ReportCharts {
			res[c.Name()] = c
		}
	}
	return res
}

func (w *Workspace) buildReportCounterMap(modMap modconfig.ModMap) map[string]*modconfig.ReportCounter {
	var res = make(map[string]*modconfig.ReportCounter)

	for _, p := range w.Mod.ReportCounters {
		res[p.Name()] = p
	}

	for _, mod := range modMap {
		for _, p := range mod.ReportCounters {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildReportHierarchyMap(modMap modconfig.ModMap) map[string]*modconfig.ReportHierarchy {
	var res = make(map[string]*modconfig.ReportHierarchy)

	for _, p := range w.Mod.ReportHierarchies {
		res[p.Name()] = p
	}

	for _, mod := range modMap {
		for _, p := range mod.ReportHierarchies {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildReportImageMap(modMap modconfig.ModMap) map[string]*modconfig.ReportImage {
	var res = make(map[string]*modconfig.ReportImage)

	for _, p := range w.Mod.ReportImages {
		res[p.Name()] = p
	}

	for _, mod := range modMap {
		for _, p := range mod.ReportImages {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildReportInputMap(modMap modconfig.ModMap) map[string]*modconfig.ReportInput {
	var res = make(map[string]*modconfig.ReportInput)

	for _, p := range w.Mod.ReportInputs {
		res[p.Name()] = p
	}

	for _, mod := range modMap {
		for _, p := range mod.ReportInputs {
			res[p.Name()] = p
		}
	}
	return res
}

func (w *Workspace) buildReportTableMap(modMap modconfig.ModMap) map[string]*modconfig.ReportTable {
	var res = make(map[string]*modconfig.ReportTable)

	for _, c := range w.Mod.ReportTables {
		res[c.Name()] = c
	}

	for _, mod := range modMap {
		for _, c := range mod.ReportTables {
			res[c.Name()] = c
		}
	}
	return res
}

func (w *Workspace) buildReportTextMap(modMap modconfig.ModMap) map[string]*modconfig.ReportText {
	var res = make(map[string]*modconfig.ReportText)

	for _, c := range w.Mod.ReportTexts {
		res[c.Name()] = c
	}

	for _, mod := range modMap {
		for _, c := range mod.ReportTexts {
			res[c.Name()] = c
		}
	}
	return res
}
