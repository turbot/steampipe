package workspace

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/steampipeconfig/versionmap"
	"github.com/turbot/steampipe/utils"
)

type Workspace struct {
	Path                string
	ModInstallationPath string
	Mod                 *modconfig.Mod

	// maps of mod resources from this mod and ALL DEPENDENCIES, keyed by long and short names

	Queries           map[string]*modconfig.Query
	Controls          map[string]*modconfig.Control
	Benchmarks        map[string]*modconfig.Benchmark
	Mods              map[string]*modconfig.Mod
	Reports           map[string]*modconfig.ReportContainer
	ReportContainers  map[string]*modconfig.ReportContainer
	ReportCharts      map[string]*modconfig.ReportChart
	ReportControls    map[string]*modconfig.ReportControl
	ReportCounters    map[string]*modconfig.ReportCounter
	ReportHierarchies map[string]*modconfig.ReportHierarchy
	ReportImages      map[string]*modconfig.ReportImage
	ReportTables      map[string]*modconfig.ReportTable
	ReportTexts       map[string]*modconfig.ReportText
	Variables         map[string]*modconfig.Variable

	//local  resources keyed by unqualifed name
	LocalQueries    map[string]*modconfig.Query
	LocalControls   map[string]*modconfig.Control
	LocalBenchmarks map[string]*modconfig.Benchmark

	watcher    *utils.FileWatcher
	loadLock   sync.Mutex
	exclusions []string
	// should we load/watch files recursively
	listFlag                filehelpers.ListFlag
	fileWatcherErrorHandler func(context.Context, error)
	watcherError            error
	// event handlers
	reportEventHandlers []reportevents.ReportEventHandler
	// callback function to reset display after the file watche displays messages
	onFileWatcherEventMessages func()
	modFileExists              bool
	loadPseudoResources        bool
}

// Load creates a Workspace and loads the workspace mod
func Load(ctx context.Context, workspacePath string) (*Workspace, error) {
	utils.LogTime("workspace.Load start")
	defer utils.LogTime("workspace.Load end")

	// create shell workspace
	workspace := &Workspace{
		Path: workspacePath,
	}

	// check whether the workspace contains a modfile
	// this will determine whether we load files recursively, and create pseudo resources for sql files
	workspace.setModfileExists()

	// load the .steampipe ignore file
	if err := workspace.loadExclusions(); err != nil {
		return nil, err
	}

	// load the workspace mod
	if err := workspace.loadWorkspaceMod(ctx); err != nil {
		return nil, err
	}

	// return context error so calling code can handle cancellations
	return workspace, nil
}

// LoadResourceNames builds lists of all workspace resource names
func LoadResourceNames(workspacePath string) (*modconfig.WorkspaceResources, error) {
	utils.LogTime("workspace.LoadResourceNames start")
	defer utils.LogTime("workspace.LoadResourceNames end")

	// create shell workspace
	workspace := &Workspace{
		Path: workspacePath,
	}

	// determine whether to load files recursively or just from the top level folder
	workspace.setModfileExists()

	// load the .steampipe ignore file
	if err := workspace.loadExclusions(); err != nil {
		return nil, err
	}

	return workspace.loadWorkspaceResourceName()
}

func (w *Workspace) SetupWatcher(ctx context.Context, client db_common.Client, errorHandler func(context.Context, error)) error {
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
			w.handleFileWatcherEvent(ctx, client, events)
		},
	}
	watcher, err := utils.NewWatcher(watcherOptions)
	if err != nil {
		return err
	}
	w.watcher = watcher
	// start the watcher
	watcher.Start()

	// set the file watcher error handler, which will get called when there are parsing errors
	// after a file watcher event
	w.fileWatcherErrorHandler = errorHandler
	if w.fileWatcherErrorHandler == nil {
		w.fileWatcherErrorHandler = func(ctx context.Context, err error) {
			fmt.Println()
			utils.ShowErrorWithMessage(ctx, err, "Failed to reload mod from file watcher")
		}
	}

	return nil
}

func (w *Workspace) SetOnFileWatcherEventMessages(f func()) {
	w.onFileWatcherEventMessages = f
}

// access functions
// NOTE: all access functions lock 'loadLock' - this is to avoid conflicts with the file watcher

func (w *Workspace) Close() {
	if w.watcher != nil {
		w.watcher.Close()
	}
}

// clear all resource maps
func (w *Workspace) reset() {
	w.Queries = make(map[string]*modconfig.Query)
	w.Controls = make(map[string]*modconfig.Control)
	w.Benchmarks = make(map[string]*modconfig.Benchmark)
	w.Mods = make(map[string]*modconfig.Mod)
	w.Reports = make(map[string]*modconfig.ReportContainer)
	w.ReportContainers = make(map[string]*modconfig.ReportContainer)
	w.ReportCharts = make(map[string]*modconfig.ReportChart)
	w.ReportControls = make(map[string]*modconfig.ReportControl)
	w.ReportCounters = make(map[string]*modconfig.ReportCounter)
	w.ReportHierarchies = make(map[string]*modconfig.ReportHierarchy)
	w.ReportImages = make(map[string]*modconfig.ReportImage)
	w.ReportTables = make(map[string]*modconfig.ReportTable)
	w.ReportTexts = make(map[string]*modconfig.ReportText)
	w.LocalQueries = make(map[string]*modconfig.Query)
	w.LocalControls = make(map[string]*modconfig.Control)
	w.LocalBenchmarks = make(map[string]*modconfig.Benchmark)
}

// check  whether the workspace contains a modfile
// this will determine whether we load files recursively, and create pseudo resources for sql files
func (w *Workspace) setModfileExists() {
	modFilePath := filepaths.ModFilePath(w.Path)
	_, err := os.Stat(modFilePath)
	modFileExists := err == nil

	if modFileExists {
		log.Printf("[TRACE] modfile exists in workspace folder - creating pseudo-resources and loading files recursively ")
		// only load/watch recursively if a mod sp file exists in the workspace folder
		w.listFlag = filehelpers.FilesRecursive
		w.loadPseudoResources = true
	} else {
		log.Printf("[TRACE] no modfile exists in workspace folder - NOT creating pseudoresources and onnly loading resource files from top level folder")
		w.listFlag = filehelpers.Files
		w.loadPseudoResources = false
	}
}

func (w *Workspace) loadWorkspaceMod(ctx context.Context) error {
	// clear all resource maps
	w.reset()
	// load and evaluate all variables
	inputVariables, err := w.getAllVariables(ctx)
	if err != nil {
		return err
	}

	// build run context which we use to load the workspace
	runCtx, err := w.getRunContext()
	if err != nil {
		return err
	}
	// add variables to runContext
	runCtx.AddVariables(inputVariables)

	// now load the mod
	m, err := steampipeconfig.LoadMod(w.Path, runCtx)
	if err != nil {
		return err
	}

	// now set workspace properties
	w.Mod = m
	w.Queries, w.LocalQueries = w.buildQueryMap(runCtx.LoadedDependencyMods)
	w.Controls, w.LocalControls = w.buildControlMap(runCtx.LoadedDependencyMods)
	w.Benchmarks, w.LocalBenchmarks = w.buildBenchmarkMap(runCtx.LoadedDependencyMods)
	w.Reports = w.buildReportMap(runCtx.LoadedDependencyMods)
	w.ReportContainers = w.buildReportContainerMap(runCtx.LoadedDependencyMods)
	w.ReportCharts = w.buildReportChartMap(runCtx.LoadedDependencyMods)
	w.ReportControls = w.buildReportControlMap(runCtx.LoadedDependencyMods)
	w.ReportCounters = w.buildReportCounterMap(runCtx.LoadedDependencyMods)
	w.ReportHierarchies = w.buildReportHierarchyMap(runCtx.LoadedDependencyMods)
	w.ReportImages = w.buildReportImageMap(runCtx.LoadedDependencyMods)
	w.ReportTables = w.buildReportTableMap(runCtx.LoadedDependencyMods)
	w.ReportTexts = w.buildReportTextMap(runCtx.LoadedDependencyMods)

	// set variables on workspace
	w.Variables = m.Variables
	w.Mods = runCtx.LoadedDependencyMods
	// NOTE: add in the workspace mod to the dependency mods
	w.Mods[w.Mod.Name()] = w.Mod

	return nil
}

// build options used to load workspace
// set flags to create pseudo resources and a default mod if needed
func (w *Workspace) getRunContext() (*parse.RunContext, error) {
	parseFlag := parse.CreateDefaultMod
	if w.loadPseudoResources {
		parseFlag |= parse.CreatePseudoResources
	}
	// load the workspace lock
	workspaceLock, err := versionmap.LoadWorkspaceLock(w.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load installation cache from %s: %s", w.Path, err)
	}

	runCtx := parse.NewRunContext(
		workspaceLock,
		w.Path,
		parseFlag,
		&filehelpers.ListOptions{
			// listFlag specifies whether to load files recursively
			Flags:   w.listFlag,
			Exclude: w.exclusions,
			// only load .sp files
			Include: filehelpers.InclusionsFromExtensions([]string{constants.ModDataExtension}),
		})

	return runCtx, nil
}

func (w *Workspace) loadExclusions() error {
	// default to ignoring hidden files and folders
	w.exclusions = []string{
		fmt.Sprintf("%s/**/.*", w.Path),
		fmt.Sprintf("%s/.*", w.Path),
	}

	ignorePath := filepath.Join(w.Path, filepaths.WorkspaceIgnoreFile)
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

func (w *Workspace) loadWorkspaceResourceName() (*modconfig.WorkspaceResources, error) {
	// build options used to load workspace
	runCtx, err := w.getRunContext()
	if err != nil {
		return nil, err
	}

	workspaceResourceNames, err := steampipeconfig.LoadModResourceNames(w.Path, runCtx)
	if err != nil {
		return nil, err
	}

	// TODO load resource names for dependency mods
	//modsPath := file_paths.WorkspaceModPath(w.Path)
	//dependencyResourceNames, err := w.loadModDependencyResourceNames(modsPath)
	//if err != nil {
	//	return nil, err
	//}

	return workspaceResourceNames, nil
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
		query, queryProvider, err := w.ResolveQueryAndArgs(arg)
		if err != nil {
			return nil, nil, err
		}
		if len(query) > 0 {
			queries = append(queries, query)
			resourceMap.AddQueryProvider(queryProvider)
		}
	}
	return queries, resourceMap, nil
}

// ResolveQueryAndArgs attempts to resolve 'arg' to a query and query args
func (w *Workspace) ResolveQueryAndArgs(sqlString string) (string, modconfig.QueryProvider, error) {
	var args = &modconfig.QueryArgs{}

	var err error

	// if this looks like a named query or named control invocation, parse the sql string for arguments
	if isNamedQueryOrControl(sqlString) {
		sqlString, args, err = parse.ParsePreparedStatementInvocation(sqlString)
		if err != nil {
			return "", nil, err
		}
	}

	return w.resolveQuery(sqlString, args)
}

func (w *Workspace) resolveQuery(sqlString string, args *modconfig.QueryArgs) (string, modconfig.QueryProvider, error) {
	// query or control providing the named query

	var queryProvider modconfig.QueryProvider

	log.Printf("[TRACE] resolveQuery %s args %s", sqlString, args)
	// 1) check if this is a control
	if control, ok := w.GetControl(sqlString); ok {
		queryProvider = control
		log.Printf("[TRACE] query string is a control: %s", control.FullName)

		// copy control SQL into query and continue resolution
		var err error
		sqlString, err = w.ResolveControlQuery(control, args)
		if err != nil {
			return "", nil, err
		}
		log.Printf("[TRACE] resolved control query: %s", sqlString)
		return sqlString, queryProvider, nil
	}

	// 2) is this a named query
	if namedQuery, ok := w.GetQuery(sqlString); ok {
		queryProvider = namedQuery
		sql, err := w.resolveNamedQuery(namedQuery, args)
		if err != nil {
			return "", nil, err
		}
		return sql, queryProvider, nil
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
		return fileQuery, queryProvider, nil
	}

	// 4) so we have not managed to resolve this - if it looks like a named query or control, return an error
	if isNamedQueryOrControl(sqlString) {
		return "", nil, fmt.Errorf("'%s' not found in workspace", sqlString)
	}

	// 5) just use the query string as is and assume it is valid SQL
	return sqlString, queryProvider, nil
}

func (w *Workspace) resolveNamedQuery(namedQuery *modconfig.Query, args *modconfig.QueryArgs) (string, error) {
	/// if there are no params, just return the sql
	if len(namedQuery.Params) == 0 {
		return typehelpers.SafeString(namedQuery.SQL), nil
	}

	// so there are params - this will be a prepared statement
	sql, err := modconfig.GetPreparedStatementExecuteSQL(namedQuery, args)
	if err != nil {
		return "", err
	}
	return sql, nil
}

// ResolveControlQuery resolves the query for the given Control
func (w *Workspace) ResolveControlQuery(control *modconfig.Control, args *modconfig.QueryArgs) (string, error) {
	args, err := w.resolveControlArgs(control, args)
	if err != nil {
		return "", err
	}

	log.Printf("[TRACE] ResolveControlQuery for %s", control.FullName)

	// verify we have either SQL or a Query defined
	if control.SQL == nil && control.Query == nil {
		// this should never happen as we should catch it in the parsing stage
		return "", fmt.Errorf("%s must define either a 'sql' property or a 'query' property", control.FullName)
	}

	// determine the source for the query
	// - this will either be the control itself or any named query the control refers to
	// either via its SQL property (passing a query name) or Query property (using a reference to a query object)

	// if a query is provided, us that to resolve the sql
	if control.Query != nil {
		return w.resolveNamedQuery(control.Query, args)
	}

	// if the control has SQL set, use that
	if control.SQL != nil {
		controlSQL := typehelpers.SafeString(control.SQL)
		log.Printf("[TRACE] control defines inline SQL")

		// if the control SQL refers to a named query, this is the same as if the control 'Query' property is set
		if namedQuery, ok := w.GetQuery(controlSQL); ok {
			// in this case, it is NOT valid for the control to define its own Param definitions
			if control.Params != nil {
				return "", fmt.Errorf("%s has an 'SQL' property which refers to %s, so it cannot define 'param' blocks", control.FullName, namedQuery.FullName)
			}
			return w.resolveNamedQuery(namedQuery, args)
		}
		// so the control sql is NOT a named query
		// if there are NO params, use the control SQL as is
		if len(control.Params) == 0 {
			return controlSQL, nil
		}
		// so the control sql is NOT a named query
		// if there are NO params, use the control SQL as is
		if len(control.Params) == 0 {
			return controlSQL, nil
		}
	}

	// so the control defines SQL and has params - it is a prepared statement
	return modconfig.GetPreparedStatementExecuteSQL(control, args)
}

func (w *Workspace) resolveControlArgs(control *modconfig.Control, args *modconfig.QueryArgs) (*modconfig.QueryArgs, error) {
	// if no args were provided,  set args to control args (which may also be nil!)
	if args == nil || args.Empty() {
		return control.Args, nil
		log.Printf("[TRACE] using control args: %s", args)
	}
	// so command line args were provided
	// check if the control supports them (it will NOT is it specifies a 'query' property)
	if control.Query != nil {
		return nil, fmt.Errorf("%s defines a query property and so does not support command line arguments", control.FullName)
	}
	log.Printf("[TRACE] using command line args: %s", args)

	// so the control defines SQL and has params - it is a prepared statement
	return args, nil
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
