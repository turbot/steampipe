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
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/filewatcher"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/modinstaller"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
	"github.com/turbot/steampipe/pkg/utils"
)

type Workspace struct {
	Path                string
	ModInstallationPath string
	Mod                 *modconfig.Mod

	Mods map[string]*modconfig.Mod
	// the input variables used in the parse
	VariableValues map[string]string
	CloudMetadata  *steampipeconfig.CloudMetadata

	// source snapshot paths
	// if this is set, no other mod resources are loaded and
	// the ResourceMaps returned by GetModResources will contain only the snapshots
	SourceSnapshots []string

	watcher     *filewatcher.FileWatcher
	loadLock    sync.Mutex
	exclusions  []string
	modFilePath string
	// should we load/watch files recursively
	listFlag                filehelpers.ListFlag
	fileWatcherErrorHandler func(context.Context, error)
	watcherError            error
	// event handlers
	dashboardEventHandlers []dashboardevents.DashboardEventHandler
	// callback function called when there is a file watcher event
	onFileWatcherEventMessages func()
	loadPseudoResources        bool
	// channel used to send dashboard events to the handleDashbooardEvent goroutine
	dashboardEventChan chan dashboardevents.DashboardEvent
}

// Load creates a Workspace and loads the workspace mod
func Load(ctx context.Context, workspacePath string) (*Workspace, error) {
	utils.LogTime("workspace.Load start")
	defer utils.LogTime("workspace.Load end")

	workspace, err := createShellWorkspace(workspacePath)
	if err != nil {
		return nil, err
	}

	// load the workspace mod
	if err := workspace.loadWorkspaceMod(ctx); err != nil {
		return nil, err
	}

	// return context error so calling code can handle cancellations
	return workspace, nil
}

// LoadVariables creates a Workspace and uses it to load all variables, ignoring any value resolution errors
// this is use for the variable list command
func LoadVariables(ctx context.Context, workspacePath string) ([]*modconfig.Variable, error) {
	utils.LogTime("workspace.LoadVariables start")
	defer utils.LogTime("workspace.LoadVariables end")

	// create shell workspace
	workspace, err := createShellWorkspace(workspacePath)
	if err != nil {
		return nil, err
	}

	// resolve variables values, WITHOUT validating missing vars
	validateMissing := false
	variableMap, err := workspace.getInputVariables(ctx, validateMissing)
	if err != nil {
		return nil, err
	}

	// convert into a sorted array
	return variableMap.ToArray(), nil
}

func createShellWorkspace(workspacePath string) (*Workspace, error) {
	// create shell workspace
	workspace := &Workspace{
		Path:           workspacePath,
		VariableValues: make(map[string]string),
	}

	// check whether the workspace contains a modfile
	// this will determine whether we load files recursively, and create pseudo resources for sql files
	workspace.setModfileExists()

	// load the .steampipe ignore file
	if err := workspace.loadExclusions(); err != nil {
		return nil, err
	}

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
	watcherOptions := &filewatcher.WatcherOptions{
		Directories: []string{w.Path},
		Include:     filehelpers.InclusionsFromExtensions(steampipeconfig.GetModFileExtensions()),
		Exclude:     w.exclusions,
		ListFlag:    w.listFlag,
		EventMask:   fsnotify.Create | fsnotify.Remove | fsnotify.Rename | fsnotify.Write,
		// we should look into passing the callback function into the underlying watcher
		// we need to analyze the kind of errors that come out from the watcher and
		// decide how to handle them
		// OnError: errCallback,
		OnChange: func(events []fsnotify.Event) {
			w.handleFileWatcherEvent(ctx, client, events)
		},
	}
	watcher, err := filewatcher.NewWatcher(watcherOptions)
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
			error_helpers.ShowErrorWithMessage(ctx, err, "failed to reload mod from file watcher")
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
	if w.dashboardEventChan != nil {
		close(w.dashboardEventChan)
	}
}

func (w *Workspace) ModfileExists() bool {
	return len(w.modFilePath) > 0
}

// check  whether the workspace contains a modfile
// this will determine whether we load files recursively, and create pseudo resources for sql files
func (w *Workspace) setModfileExists() {
	modFile, err := w.findModFilePath(w.Path)
	modFileExists := err != ErrorNoModDefinition

	if modFileExists {
		log.Printf("[TRACE] modfile exists in workspace folder - creating pseudo-resources and loading files recursively ")
		// only load/watch recursively if a mod sp file exists in the workspace folder
		w.listFlag = filehelpers.FilesRecursive
		w.loadPseudoResources = true
		w.modFilePath = modFile

		// also set it in the viper config, so that it is available to whoever is using it
		viper.Set(constants.ArgModLocation, filepath.Dir(modFile))
		w.Path = filepath.Dir(modFile)
	} else {
		log.Printf("[TRACE] no modfile exists in workspace folder - NOT creating pseudoresources and only loading resource files from top level folder")
		w.listFlag = filehelpers.Files
		w.loadPseudoResources = false
	}
}

func (w *Workspace) findModFilePath(folder string) (string, error) {
	folder, err := filepath.Abs(folder)
	if err != nil {
		return "", err
	}
	modFilePath := filepaths.ModFilePath(folder)
	_, err = os.Stat(modFilePath)
	if err == nil {
		// found the modfile
		return modFilePath, nil
	}

	if os.IsNotExist(err) {
		// if the file wasn't found, search in the parent directory
		parent := filepath.Dir(folder)
		if folder == parent {
			// this typically means that we are already in the root directory
			return "", ErrorNoModDefinition
		}
		return w.findModFilePath(filepath.Dir(folder))
	}
	return modFilePath, nil
}

func (w *Workspace) loadWorkspaceMod(ctx context.Context) error {
	// resolve values of all input variables
	// we WILL validate missing variables when loading
	validateMissing := true
	inputVariables, err := w.getInputVariables(ctx, validateMissing)
	if err != nil {
		return err
	}
	// populate the parsed variable values
	w.VariableValues = inputVariables.VariableValues

	// build run context which we use to load the workspace
	parseCtx, err := w.getParseContext()
	if err != nil {
		return err
	}
	// add variables to runContext
	parseCtx.AddInputVariables(inputVariables)
	// do not reload variables as we already have them
	parseCtx.BlockTypeExclusions = []string{modconfig.BlockTypeVariable}

	// load the workspace mod
	m, err := steampipeconfig.LoadMod(w.Path, parseCtx)
	if err != nil {
		return err
	}

	// now set workspace properties
	// populate the mod references map references
	m.ResourceMaps.PopulateReferences()
	// set the mod
	w.Mod = m
	w.Mods = parseCtx.LoadedDependencyMods
	// NOTE: add in the workspace mod to the dependency mods
	w.Mods[w.Mod.Name()] = w.Mod

	// verify all runtime dependencies can be resolved
	return w.verifyResourceRuntimeDependencies()
}

func (w *Workspace) getInputVariables(ctx context.Context, validateMissing bool) (*modconfig.ModVariableMap, error) {
	// build a run context just to use to load variable definitions
	variablesRunCtx, err := w.getParseContext()
	if err != nil {
		return nil, err
	}

	// load variable definitions
	variableMap, err := steampipeconfig.LoadVariableDefinitions(w.Path, variablesRunCtx)
	if err != nil {
		return nil, err
	}

	return steampipeconfig.GetVariableValues(ctx, variablesRunCtx, variableMap, validateMissing)
}

// build options used to load workspace
// set flags to create pseudo resources and a default mod if needed
func (w *Workspace) getParseContext() (*parse.ModParseContext, error) {
	parseFlag := parse.CreateDefaultMod
	if w.loadPseudoResources {
		parseFlag |= parse.CreatePseudoResources
	}
	workspaceLock, err := w.loadWorkspaceLock()
	if err != nil {
		return nil, err
	}
	parseCtx := parse.NewModParseContext(
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

	return parseCtx, nil
}

// load the workspace lock, migrating it if necessary
func (w *Workspace) loadWorkspaceLock() (*versionmap.WorkspaceLock, error) {
	workspaceLock, err := versionmap.LoadWorkspaceLock(w.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load installation cache from %s: %s", w.Path, err)
	}

	// if this is the old format, migrate by reinstalling dependencies
	if workspaceLock.StructVersion() != versionmap.WorkspaceLockStructVersion {
		opts := &modinstaller.InstallOpts{WorkspacePath: viper.GetString(constants.ArgModLocation)}
		installData, err := modinstaller.InstallWorkspaceDependencies(opts)
		if err != nil {
			return nil, err
		}
		workspaceLock = installData.NewLock
	}
	return workspaceLock, nil
}

func (w *Workspace) loadExclusions() error {
	// default to ignoring hidden files and folders
	w.exclusions = []string{
		// ignore any hidden folder
		fmt.Sprintf("%s/.*", w.Path),
		// and sub files/folders of hidden folders
		fmt.Sprintf("%s/.*/**", w.Path),
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
	parseCtx, err := w.getParseContext()
	if err != nil {
		return nil, err
	}

	workspaceResourceNames, err := steampipeconfig.LoadModResourceNames(w.Path, parseCtx)
	if err != nil {
		return nil, err
	}

	return workspaceResourceNames, nil
}

func (w *Workspace) verifyResourceRuntimeDependencies() error {
	for _, d := range w.Mod.ResourceMaps.Dashboards {
		if err := d.BuildRuntimeDependencyTree(w); err != nil {
			return err
		}
	}
	return nil
}
