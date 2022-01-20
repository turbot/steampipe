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

func (w *Workspace) GetControlMap() map[string]*modconfig.Control {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	return w.Controls
}

func (w *Workspace) GetLocalControlMap() map[string]*modconfig.Control {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	return w.LocalControls
}

func (w *Workspace) GetQueryMap() map[string]*modconfig.Query {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	return w.Queries
}

func (w *Workspace) GetLocalQueryMap() map[string]*modconfig.Query {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	return w.LocalQueries
}

// GetChildControls builds a flat list of all controls in the worlspace, including dependencies
func (w *Workspace) GetChildControls() []*modconfig.Control {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()
	var result []*modconfig.Control
	// the workspace resource maps have duplicate entries, keyed by long and short name.
	// keep track of which controls we have identified in order to avoid dupes
	controlsMatched := make(map[string]bool)
	for _, c := range w.Controls {
		if _, alreadyMatched := controlsMatched[c.Name()]; !alreadyMatched {
			controlsMatched[c.Name()] = true
			result = append(result, c)
		}
	}
	return result
}

// GetResourceMaps returns all resource maps
// TODO KAI CHECK LACK OF LOCKING IS OK HERE
// NOTE: this function DOES NOT LOCK the load lock so should only be called in a context where the file watcher is not running
func (w *Workspace) GetResourceMaps() *modconfig.WorkspaceResourceMaps {
	workspaceMap := &modconfig.WorkspaceResourceMaps{
		Mods:             make(map[string]*modconfig.Mod),
		Queries:          w.Queries,
		Controls:         w.Controls,
		Benchmarks:       w.Benchmarks,
		Variables:        w.Variables,
		Reports:          w.Reports,
		ReportContainers: w.ReportContainers,
		ReportCharts:     w.ReportCharts,
		ReportCounters:   w.ReportCounters,
		ReportImages:     w.ReportImages,
		ReportTables:     w.ReportTables,
		ReportTexts:      w.ReportTexts,
	}
	workspaceMap.PopulateReferences()

	// TODO add in all mod dependencies

	if !w.Mod.IsDefaultMod() {
		workspaceMap.Mods[w.Mod.Name()] = w.Mod
	}

	return workspaceMap
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
