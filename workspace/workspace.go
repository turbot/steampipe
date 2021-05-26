package workspace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/turbot/steampipe/report/reporteventpublisher"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/utils"
)

type Workspace struct {
	Path string
	Mod  *modconfig.Mod

	// maps of mod resources from this mod and ALL DEPENDENCIES, keyed by long and short names
	QueryMap     map[string]*modconfig.Query
	ControlMap   map[string]*modconfig.Control
	BenchmarkMap map[string]*modconfig.Benchmark
	ModMap       map[string]*modconfig.Mod
	ReportMap    map[string]*modconfig.Report
	PanelMap     map[string]*modconfig.Panel

	watcher    *utils.FileWatcher
	loadLock   sync.Mutex
	exclusions []string
	// should we load/watch files recursively
	listFlag     filehelpers.ListFlag
	watcherError error
	// event publisher
	reportEventPublisher *reporteventpublisher.ReportEventPublisher
}

// Load creates a Workspace and loads the workspace mod
func Load(workspacePath string) (*Workspace, error) {
	utils.LogTime("workspace.Load start")
	defer utils.LogTime("workspace.Load end")

	// create shell workspace
	workspace := &Workspace{
		Path:                 workspacePath,
		reportEventPublisher: &reporteventpublisher.ReportEventPublisher{},
	}

	// determine whether to load files recursively or just from the top level folder
	workspace.setListFlag()

	// load the .steampipe ignore file
	if err := workspace.LoadExclusions(); err != nil {
		return nil, err
	}

	if err := workspace.loadMod(); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (w *Workspace) SetupWatcher(client *db.Client) error {

	watcherOptions := &utils.WatcherOptions{
		Directories: []string{w.Path},
		Include:     filehelpers.InclusionsFromExtensions(steampipeconfig.GetModFileExtensions()),
		Exclude:     w.exclusions,
		OnChange: func(events []fsnotify.Event) {
			w.handleFileWatcherEvent(client, events)
		},
		ListFlag: w.listFlag,
		//onError:          nil,
	}
	watcher, err := utils.NewWatcher(watcherOptions)
	if err != nil {
		return err
	}
	w.watcher = watcher

	return nil
}

func (w *Workspace) LoadExclusions() error {
	w.exclusions = []string{}

	ignorePath := filepath.Join(w.Path, constants.WorkspaceIgnoreFile)
	file, err := os.Open(ignorePath)
	if err != nil {
		// if file does not exist, just return
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) != 0 && !strings.HasPrefix(line, "#") {
			// add exclusion to the workspace path (to ensure relative patterns work)
			absoluteExclusion := filepath.Join(w.Path, line)
			w.exclusions = append(w.exclusions, absoluteExclusion)
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (w *Workspace) GetSortedBenchmarksAndControlNames() []string {
	benchmarkList := []string{}
	controlList := []string{}

	for key := range w.BenchmarkMap {
		benchmarkList = append(benchmarkList, key)
	}

	for key := range w.ControlMap {
		controlList = append(controlList, key)
	}

	sort.Strings(benchmarkList)
	sort.Strings(controlList)

	return append(benchmarkList, controlList...)
}

func (w *Workspace) GetSortedNamedQueryNames() []string {
	namedQueries := []string{}
	for key := range w.GetQueryMap() {
		namedQueries = append(namedQueries, key)
	}
	sort.Strings(namedQueries)
	return namedQueries
}

func (w *Workspace) Close() {
	if w.watcher != nil {
		w.watcher.Close()
	}
}

func (w *Workspace) GetQueryMap() map[string]*modconfig.Query {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	return w.QueryMap
}

func (w *Workspace) GetQuery(queryName string) (*modconfig.Query, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	if query, ok := w.QueryMap[queryName]; ok {
		return query, true
	}
	return nil, false
}

func (w *Workspace) GetControlMap() map[string]*modconfig.Control {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	return w.ControlMap
}

func (w *Workspace) GetControl(controlName string) (*modconfig.Control, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	if control, ok := w.ControlMap[controlName]; ok {
		return control, true
	}
	return nil, false
}

// GetChildControls builds a flat list of all controls in the worlspace, including dependencies
func (w *Workspace) GetChildControls() []*modconfig.Control {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()
	var result []*modconfig.Control
	// the workspace resource maps have duplicate entries, keyed by long and short name.
	// keep track of which controls we have identified in order to avoid dupes
	controlsMatched := make(map[string]bool)
	for _, c := range w.ControlMap {
		if _, alreadyMatched := controlsMatched[c.Name()]; !alreadyMatched {
			controlsMatched[c.Name()] = true
			result = append(result, c)
		}
	}
	return result
}

func (w *Workspace) GetResourceMaps() *modconfig.WorkspaceResourceMaps {
	workspaceMap := &modconfig.WorkspaceResourceMaps{
		ModMap:       make(map[string]*modconfig.Mod),
		QueryMap:     w.QueryMap,
		ControlMap:   w.ControlMap,
		BenchmarkMap: w.BenchmarkMap,
	}
	// TODO add in all mod dependencies
	if !w.Mod.IsDefaultMod() {
		workspaceMap.ModMap[w.Mod.Name()] = w.Mod
	}

	return workspaceMap
}

// GetMod attempts to return the mod with a name matching 'modName'
// It first checks the workspace mod, then checks all mod dependencies
func (w *Workspace) GetMod(modName string) *modconfig.Mod {
	// is it the workspace mod?
	if modName == w.Mod.Name() {
		return w.Mod
	}
	// try workspace mod dependencies
	return w.ModMap[modName]
}

// Mods returns a flat list of all mods - the workspace mod and depdnency mods
func (w *Workspace) Mods() []*modconfig.Mod {
	var res = []*modconfig.Mod{w.Mod}
	for _, m := range w.ModMap {
		res = append(res, m)
	}
	return res
}

// determine whether to load files recursively or just from the top level folder
// if there is a mod file in the workspace folder, load recursively
func (w *Workspace) setListFlag() {

	modFilePath := filepath.Join(w.Path, "mod.sp")
	_, err := os.Stat(modFilePath)
	modFileExists := err == nil
	if modFileExists {
		// only load/watch recursively if a mod sp file exists in the workspace folder
		w.listFlag = filehelpers.FilesRecursive
	} else {
		w.listFlag = filehelpers.Files
	}
}

func (w *Workspace) loadMod() error {
	// parse all hcl files in the workspace folder (non recursively) and either parse or create a mod
	// it is valid for 0 or 1 mod to be defined (if no mod is defined, create a default one)
	// populate mod with all hcl resources we parse

	// build options used to load workspace
	// set flags to create pseudo resources and a default mod if needed
	opts := &parse.ParseModOptions{
		Flags: parse.CreatePseudoResources | parse.CreateDefaultMod,
		ListOptions: &filehelpers.ListOptions{
			// listFlag specifies whether to load files recursively
			Flags:   w.listFlag,
			Exclude: w.exclusions,
		},
	}

	// clear all our maps
	w.QueryMap = make(map[string]*modconfig.Query)
	w.ControlMap = make(map[string]*modconfig.Control)
	w.BenchmarkMap = make(map[string]*modconfig.Benchmark)

	m, err := steampipeconfig.LoadMod(w.Path, opts)
	if err != nil {
		return err
	}
	w.Mod = m

	// now load all mods in the workspace
	modsPath := constants.WorkspaceModPath(w.Path)
	modMap, err := w.loadModDependencies(modsPath)
	if err != nil {
		return err
	}

	w.QueryMap = w.buildQueryMap(modMap)
	w.ControlMap = w.buildControlMap(modMap)
	w.BenchmarkMap = w.buildBenchmarkMap(modMap)
	w.ReportMap = w.buildReportMap(modMap)
	w.PanelMap = w.buildPanelMap(modMap)
	w.ModMap = modMap

	return nil
}

// load all dependencies of workspace mod
// used to load all mods in a workspace, using the workspace manifest mod
func (w *Workspace) loadModDependencies(modsFolder string) (modconfig.ModMap, error) {
	var res = modconfig.ModMap{
		w.Mod.Name(): w.Mod,
	}
	if err := steampipeconfig.LoadModDependencies(w.Mod, modsFolder, res, false); err != nil {
		return nil, err
	}
	return res, nil
}

func (w *Workspace) buildQueryMap(modMap modconfig.ModMap) map[string]*modconfig.Query {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.Query)

	// for LOCAL queries, add map entries keyed by both short name: query.<shortName> and  long name: <modName>.query.<shortName?
	for _, q := range w.Mod.Queries {
		res[q.Name()] = q
		longName := fmt.Sprintf("%s.query.%s", types.SafeString(w.Mod.ShortName), q.ShortName)
		res[longName] = q
	}

	// for mode dependencies, add queries keyed by long name only
	for _, mod := range modMap {
		for _, q := range mod.Queries {
			longName := fmt.Sprintf("%s.query.%s", types.SafeString(mod.ShortName), q.ShortName)
			res[longName] = q
		}
	}
	return res
}

func (w *Workspace) buildControlMap(modMap modconfig.ModMap) map[string]*modconfig.Control {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.Control)

	// for LOCAL controls, add map entries keyed by both short name: control.<shortName> and  long name: <modName>.control.<shortName?
	for _, c := range w.Mod.Controls {
		res[c.Name()] = c
		res[c.QualifiedName()] = c
	}

	// for mode dependencies, add queries keyed by long name only
	for _, mod := range modMap {
		for _, c := range mod.Controls {
			res[c.QualifiedName()] = c
		}
	}
	return res
}

func (w *Workspace) buildBenchmarkMap(modMap modconfig.ModMap) map[string]*modconfig.Benchmark {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.Benchmark)

	// for LOCAL controls, add map entries keyed by both short name: benchmark.<shortName> and  long name: <modName>.benchmark.<shortName?
	for _, c := range w.Mod.Benchmarks {
		res[c.Name()] = c
		res[c.QualifiedName()] = c
	}

	// for mod dependencies, add queries keyed by long name only
	for _, mod := range modMap {
		for _, c := range mod.Benchmarks {
			res[c.QualifiedName()] = c
		}
	}
	return res
}

func (w *Workspace) buildReportMap(modMap modconfig.ModMap) map[string]*modconfig.Report {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.Report)

	// for LOCAL reports, add map entries keyed by both short name: benchmark.<shortName> and  long name: <modName>.benchmark.<shortName?
	for _, r := range w.Mod.Reports {
		res[r.Name()] = r
		res[r.QualifiedName()] = r
	}

	// for mod dependencies, add queries keyed by long name only
	for _, mod := range modMap {
		for _, r := range mod.Reports {
			res[r.QualifiedName()] = r
		}
	}
	return res
}

func (w *Workspace) buildPanelMap(modMap modconfig.ModMap) map[string]*modconfig.Panel {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.Panel)

	// for LOCAL panels, add map entries keyed by both short name: benchmark.<shortName> and  long name: <modName>.benchmark.<shortName?
	for _, r := range w.Mod.Panels {
		res[r.Name()] = r
		res[r.QualifiedName()] = r
	}

	// for mod dependencies, add queries keyed by long name only
	for _, mod := range modMap {
		for _, p := range mod.Panels {
			res[p.QualifiedName()] = p
		}
	}
	return res
}

func (w *Workspace) handleFileWatcherEvent(client *db.Client, events []fsnotify.Event) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// we build a list of diffs for panels and workspaces so store the old ones
	// TODO - same for all resources??
	prevPanels, prevReports := w.saveCurrentResources()

	err := w.loadMod()
	if err != nil {
		// if we are already in an error state, do not show error
		if w.watcherError == nil {
			fmt.Println()
			utils.ShowErrorWithMessage(err, "Failed to reload mod from file watcher")
		}
	}
	// now store/clear watcher error so we only show message once
	w.watcherError = err
	// todo detect differences and only refresh if necessary
	db.UpdateMetadataTables(w.GetResourceMaps(), client)

	w.raiseChangeEvents(prevPanels, prevReports)
}

func (w *Workspace) saveCurrentResources() (map[string]*modconfig.Panel, map[string]*modconfig.Report) {
	prevPanels := make(map[string]*modconfig.Panel, len(w.PanelMap))
	prevReports := make(map[string]*modconfig.Report, len(w.ReportMap))
	for name, p := range w.PanelMap {
		prevPanels[name] = p
	}
	for name, p := range w.ReportMap {
		prevReports[name] = p
	}
	return prevPanels, prevReports
}

func (w *Workspace) raiseChangeEvents(prevPanels map[string]*modconfig.Panel, prevReports map[string]*modconfig.Report) {
	event := &reportevents.ReportChanged{}

	// first detect detect changes to existing panels/reports and removed panels and reports
	for name, prevPanel := range prevPanels {
		if currentPanel, ok := w.PanelMap[name]; ok {
			diff := prevPanel.Diff(currentPanel)
			if diff.HasChanges() {
				event.ChangedPanels = append(event.ChangedPanels, diff)
			}
		} else {
			event.DeletedPanels = append(event.DeletedPanels, prevPanel)
		}
	}
	for name, prevReport := range prevReports {
		if currentReport, ok := w.ReportMap[name]; ok {
			diff := prevReport.Diff(currentReport)
			if diff.HasChanges() {
				event.ChangedReports = append(event.ChangedReports, diff)
			}
		} else {
			event.DeletedReports = append(event.DeletedReports, prevReport)
		}
	}
	// now detect new panels/reports
	for name, p := range w.PanelMap {
		if _, ok := prevPanels[name]; !ok {
			event.NewPanels = append(event.NewPanels, p)
		}
	}
	for name, p := range w.ReportMap {
		if _, ok := prevReports[name]; !ok {
			event.NewReports = append(event.NewReports, p)
		}
	}
	if event.HasChanges() {
		w.PublishReportEvent(event)
	}

}
