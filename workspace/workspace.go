package workspace

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/types"
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

// LoadResourceNames builds lists of all workspace respurce names
func LoadResourceNames(workspacePath string) (*modconfig.WorkspaceResources, error) {
	utils.LogTime("workspace.LoadResourceNames start")
	defer utils.LogTime("workspace.LoadResourceNames end")

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

	return workspace.loadWorkspaceResourceName()
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

// access functions
// NOTE: all access functions lock 'loadLock' - this is to avoid conflicts with th efile watcher

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

// GetResourceMaps returns all resource maps
// NOTE: this function DOES NOT LOCK the load lock so should only be called in a context where the file watcher is not running
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
func (w *Workspace) loadWorkspaceResourceName() (*modconfig.WorkspaceResources, error) {
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

	workspaceResourceNames, err := steampipeconfig.LoadModResourceNames(w.Path, opts)
	if err != nil {
		return nil, err
	}

	// TODO load resource names for dependency mods
	//modsPath := constants.WorkspaceModPath(w.Path)
	//dependencyResourceNames, err := w.loadModDependencyResourceNames(modsPath)
	//if err != nil {
	//	return nil, err
	//}

	return workspaceResourceNames, nil
}

// load resource names for all dependencies of workspace mod
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
	// default to ignoring hidden files and folders
	w.exclusions = []string{"**/.*"}

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
func (w *Workspace) GetQueriesFromArgs(args []string) ([]string, *modconfig.WorkspaceResourceMaps, error) {
	utils.LogTime("execute.GetQueriesFromArgs start")
	defer utils.LogTime("execute.GetQueriesFromArgs end")

	var queries []string
	// build map of prepared statement providers
	var resourceMap = modconfig.NewWorkspaceResourceMaps()
	for _, arg := range args {
		query, preparedStatementProvider, err := w.ResolveQueryAndArgs(arg)
		if err != nil {
			return nil, nil, err
		}
		if len(query) > 0 {
			queries = append(queries, query)
			resourceMap.AddPreparedStatementProvider(preparedStatementProvider)
		}
	}
	return queries, resourceMap, nil
}

// ResolveQueryAndArgs attempts to resolve 'arg' to a query and query args
func (w *Workspace) ResolveQueryAndArgs(sqlString string) (string, modconfig.PreparedStatementProvider, error) {
	var args *modconfig.QueryArgs
	var err error

	// if this looks like a named query or named control invocation, parse the sql string for arguments
	if isNamedQueryOrControl(sqlString) {
		sqlString, args, err = parse.ParsePreparedStatementInvocation(sqlString)
		if err != nil {
			return "", nil, err
		}
	}

	return w.ResolveQuery(sqlString, args)
}

func (w *Workspace) ResolveQuery(sqlString string, args *modconfig.QueryArgs) (string, modconfig.PreparedStatementProvider, error) {
	// query or control providing the named query
	var preparedStatementProvider modconfig.PreparedStatementProvider

	log.Printf("[TRACE] ResolveQuery %s args %s", sqlString, args)
	// 1) check if this is a control
	if control, ok := w.GetControl(sqlString); ok {
		preparedStatementProvider = control
		log.Printf("[TRACE] query string is a control: %s", control.FullName)

		if args == nil || args.Empty() {
			// set args to control args (which may also be nil!)
			args = control.Args
			log.Printf("[TRACE] using control args: %s", args)
		} else {
			// so command line args were provided
			// check if the control supports them (it will NOT is it specifies a 'query' property)
			if control.Query != nil {
				return "", nil, fmt.Errorf("%s defines a query property and so does not support command line arguments", control.FullName)
			}
			log.Printf("[TRACE] using command line args: %s", args)
		}

		// copy control SQL into query and continue resolution
		var err error
		sqlString, err = w.ResolveControlQuery(control)
		if err != nil {
			return "", nil, err
		}
		log.Printf("[TRACE] resolved control query: %s", sqlString)
	}

	// 2) is this a named query
	if namedQuery, ok := w.GetQuery(sqlString); ok {
		preparedStatementProvider = namedQuery
		sql, err := modconfig.GetPreparedStatementExecuteSQL(namedQuery, args)
		if err != nil {
			return "", nil, err
		}
		return sql, preparedStatementProvider, nil
	}

	// 	3) is this a file
	fileQuery, fileExists, err := w.getQueryFromFile(sqlString)
	if fileExists {
		if err != nil {
			return "", nil, fmt.Errorf("ResolveQueryAndArgs failed: error opening file '%s': %v", sqlString, err)
		}
		if len(fileQuery) == 0 {
			utils.ShowWarning(fmt.Sprintf("file '%s' does not contain any data", sqlString))
			// (just return the empty string - it will be filtered above)
		}
		return fileQuery, preparedStatementProvider, nil
	}

	// 4) so we have not managed to resolve this - if it looks like a named query or control, return an error
	if isNamedQueryOrControl(sqlString) {
		return "", nil, fmt.Errorf("'%s' not found in workspace", sqlString)
	}

	// 5) just use the query string as is and assume it is valid SQL
	return sqlString, preparedStatementProvider, nil
}

// ResolveControlQuery resolves the query for the given Control
func (w *Workspace) ResolveControlQuery(control *modconfig.Control) (string, error) {
	log.Printf("[TRACE] ResolveControlQuery for %s", control.FullName)

	// verify we have either SQL or a Query defined
	if control.SQL == nil && control.Query == nil {
		// this should never happen as we should catch it in the parsing stage
		return "", fmt.Errorf("%s must define either a 'sql' property or a 'query' property", control.FullName)
	}

	// set the source for the query - this will either be the control itself or any named query the control refers to
	// either via its SQL property (passing a query name) or Query property (using a reference to a query object)
	// default to using the 'Query' property
	var source modconfig.PreparedStatementProvider = control.Query

	// if the control has SQL set, use that
	if control.SQL != nil {
		log.Printf("[TRACE] control defines inline SQL")
		// if the control SQL refers to a named query, this is the same as if the control 'Query' property is set
		if namedQuery, ok := w.GetQuery(*control.SQL); ok {
			// in this case, it is NOT valid for the control to define its own Param definitions
			if control.Params != nil {
				return "", fmt.Errorf("%s has an 'SQL' property which refers to %s, so it cannot define 'param' blocks", control.FullName, namedQuery.FullName)
			}
			source = namedQuery
		} else {
			// so the control sql is NOT a named query - set the source to be the control
			source = control
		}
	}

	return modconfig.GetPreparedStatementExecuteSQL(source, control.Args)
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

// does this resource name look like a control or query
func isNamedQueryOrControl(name string) bool {
	parsedResourceName, err := modconfig.ParseResourceName(name)
	return err == nil && parsedResourceName.ItemType == "query" || parsedResourceName.ItemType == "control"
}
