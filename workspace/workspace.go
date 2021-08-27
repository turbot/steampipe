package workspace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/types"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/report/reportevents"
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
	listFlag                filehelpers.ListFlag
	fileWatcherErrorHandler func(error)
	watcherError            error
	// event handlers
	reportEventHandlers []reportevents.ReportEventHandler
}

// Load creates a Workspace and loads the workspace mod
func Load(workspacePath string) (*Workspace, error) {
	utils.LogTime("workspace.Load start")
	defer utils.LogTime("workspace.Load end")

	// create shell workspace
	workspace := &Workspace{
		Path: workspacePath,
	}

	// determine whether to load files recursively or just from the top level folder
	workspace.setListFlag()

	// load the .steampipe ignore file
	if err := workspace.loadExclusions(); err != nil {
		return nil, err
	}

	if err := workspace.loadWorkspaceMod(); err != nil {
		return nil, err
	}

	// return context error so calling code can handle cancellations
	return workspace, nil
}

func (w *Workspace) SetupWatcher(client db_common.Client, errorHandler func(error)) error {
	watcherOptions := &utils.WatcherOptions{
		Directories: []string{w.Path},
		Include:     filehelpers.InclusionsFromExtensions(steampipeconfig.GetModFileExtensions()),
		Exclude:     w.exclusions,
		ListFlag:    w.listFlag,
		// we should look into passing the callback function into the underlying watcher
		// we need to analyze the kind of errors that come out from the watcher and
		// decide how to handle them
		// OnError: errCallback,
		OnChange: func(events []fsnotify.Event) {
			w.handleFileWatcherEvent(client, events)
		},
	}
	watcher, err := utils.NewWatcher(watcherOptions)
	if err != nil {
		return err
	}
	w.watcher = watcher

	// set the file watcher error handler, which will get called when there are parsing errors
	// after a file watcher event
	w.fileWatcherErrorHandler = errorHandler
	if w.fileWatcherErrorHandler == nil {
		w.fileWatcherErrorHandler = func(err error) {
			fmt.Println()
			utils.ShowErrorWithMessage(err, "Failed to reload mod from file watcher")
		}
	}

	return nil
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

// clear all resource maps
func (w *Workspace) reset() {
	w.QueryMap = make(map[string]*modconfig.Query)
	w.ControlMap = make(map[string]*modconfig.Control)
	w.BenchmarkMap = make(map[string]*modconfig.Benchmark)
	w.ModMap = make(map[string]*modconfig.Mod)
	w.ReportMap = make(map[string]*modconfig.Report)
	w.PanelMap = make(map[string]*modconfig.Panel)
}

// determine whether to load files recursively or just from the top level folder
// if there is a mod file in the workspace folder, load recursively
func (w *Workspace) setListFlag() {
	modFilePath := filepath.Join(w.Path, constants.WorkspaceModFileName)
	_, err := os.Stat(modFilePath)
	modFileExists := err == nil
	if modFileExists {
		// only load/watch recursively if a mod sp file exists in the workspace folder
		w.listFlag = filehelpers.FilesRecursive
	} else {
		w.listFlag = filehelpers.Files
	}
}

func (w *Workspace) loadWorkspaceMod() error {
	inputVariables, err := w.getAllVariables()
	if err != nil {
		return err
	}

	// clear all resource maps
	w.reset()

	// build options used to load workspace
	// set flags to create pseudo resources and a default mod if needed
	opts := &parse.ParseModOptions{
		Flags: parse.CreatePseudoResources | parse.CreateDefaultMod,
		ListOptions: &filehelpers.ListOptions{
			// listFlag specifies whether to load files recursively
			Flags:   w.listFlag,
			Exclude: w.exclusions,
		},
		Variables: inputVariables.JustValues(),
	}

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

	// for mod dependencies, add queries keyed by long name only
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

func (w *Workspace) loadExclusions() error {
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

// return a map of all unique panels, keyed by name
// not we cannot just use PanelMap as this contains duplicates (qualified and unqualified version)
func (w *Workspace) getPanelMap() map[string]*modconfig.Panel {
	panels := make(map[string]*modconfig.Panel, len(w.PanelMap))
	for _, p := range w.PanelMap {
		// refetch the name property to avoid duplicates
		// (as we save resources with qualified and unqualified name)
		panels[p.Name()] = p
	}
	return panels
}

// return a map of all unique reports, keyed by name
// not we cannot just use ReportMap as this contains duplicates (qualified and unqualified version)
func (w *Workspace) getReportMap() map[string]*modconfig.Report {
	reports := make(map[string]*modconfig.Report, len(w.ReportMap))
	for _, p := range w.ReportMap {
		// refetch the name property to avoid duplicates
		// (as we save resources with qualified and unqualified name)
		reports[p.Name()] = p
	}
	return reports
}

// GetQueriesFromArgs retrieves queries from args
//
// For each arg check if it is a named query or a file, before falling back to treating it as sql
func (w *Workspace) GetQueriesFromArgs(args []string) ([]string, error) {
	utils.LogTime("execute.GetQueriesFromArgs start")
	defer utils.LogTime("execute.GetQueriesFromArgs end")

	var queries []string
	for _, arg := range args {
		// in case of a named query call with params, parse the where clause
		queryName, params, err := parse.ParsePreparedStatementInvocation(arg)
		if err != nil {
			return nil, err
		}
		query, err := w.GetQueryFromArg(queryName, params)
		if err != nil {
			return nil, err
		}
		if len(query) > 0 {
			queries = append(queries, query)
		}
	}
	return queries, nil
}

// GetQueryFromArg attempts to resolve 'arg' to a query
// the second return value indicates whether the arg was resolved as a named query/SQL file
func (w *Workspace) GetQueryFromArg(arg string, params *modconfig.QueryParams) (string, error) {
	// 1) is this a named query
	if namedQuery, ok := w.GetQuery(arg); ok {
		sql, err := namedQuery.GetExecuteSQL(params)
		if err != nil {
			return "", fmt.Errorf("GetQueryFromArg failed for value %s: %v", arg, err)
		}
		return sql, nil
	}
	// check if this is a control
	if control, ok := w.GetControl(arg); ok {
		return typeHelpers.SafeString(control.SQL), nil
	}

	// 	2) is this a file
	fileQuery, fileExists, err := w.getQueryFromFile(arg)
	if fileExists {
		if err != nil {
			return "", fmt.Errorf("GetQueryFromArg failed: error opening file '%s': %v", arg, err)
		}
		if len(fileQuery) == 0 {
			utils.ShowWarning(fmt.Sprintf("file '%s' does not contain any data", arg))
			// (just return the empty string - it will be filtered above)
		}
		return fileQuery, nil
	}

	// 3) just use the arg string as is and assume it is valid SQL
	return arg, nil
}

func (w *Workspace) getQueryFromFile(filename string) (string, bool, error) {
	// get absolute filename
	path, err := filepath.Abs(filename)
	if err != nil {
		return "", false, nil
	}
	// does it exist?
	if _, err := os.Stat(path); err != nil {
		// if this gives any error, return not exist. we may get a not found or a path too long for example
		return "", false, nil
	}

	// read file
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", true, err
	}

	return string(fileBytes), true, nil
}
