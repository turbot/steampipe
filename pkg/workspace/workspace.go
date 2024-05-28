package workspace

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/exp/maps"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/filewatcher"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/modinstaller"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
	"github.com/turbot/steampipe/pkg/task"
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
	// channel used to send dashboard events to the handleDashboardEvent goroutine
	dashboardEventChan chan dashboardevents.DashboardEvent
}

// LoadWorkspaceVars creates a Workspace and loads the variables
func LoadWorkspaceVars(ctx context.Context) (*Workspace, *modconfig.ModVariableMap, error_helpers.ErrorAndWarnings) {
	log.Printf("[INFO] LoadWorkspaceVars: creating workspace, loading variable and resolving variable values")
	workspacePath := viper.GetString(constants.ArgModLocation)

	utils.LogTime("workspace.Load start")
	defer utils.LogTime("workspace.Load end")

	workspace, err := createShellWorkspace(workspacePath)
	if err != nil {
		log.Printf("[INFO] createShellWorkspace failed %s", err.Error())
		return nil, nil, error_helpers.NewErrorsAndWarning(err)
	}

	// check if your workspace path is home dir and if modfile exists - if yes then warn and ask user to continue or not
	if err := HomeDirectoryModfileCheck(ctx, workspacePath); err != nil {
		log.Printf("[INFO] HomeDirectoryModfileCheck failed %s", err.Error())
		return nil, nil, error_helpers.NewErrorsAndWarning(err)
	}
	inputVariables, errorsAndWarnings := workspace.PopulateVariables(ctx)
	if errorsAndWarnings.Error != nil {
		log.Printf("[WARN] PopulateVariables failed %s", errorsAndWarnings.Error.Error())
		return nil, nil, errorsAndWarnings
	}

	log.Printf("[INFO] LoadWorkspaceVars succededed - got values for vars: %s", strings.Join(maps.Keys(workspace.VariableValues), ", "))

	return workspace, inputVariables, errorsAndWarnings
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
func LoadResourceNames(ctx context.Context, workspacePath string) (*modconfig.WorkspaceResources, error) {
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

	return workspace.loadWorkspaceResourceName(ctx)
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
	if ch := w.dashboardEventChan; ch != nil {
		// NOTE: set nil first
		w.dashboardEventChan = nil
		log.Printf("[TRACE] closing dashboardEventChan")
		close(ch)
	}
}

func (w *Workspace) ModfileExists() bool {
	return len(w.modFilePath) > 0
}

// check  whether the workspace contains a modfile
// this will determine whether we load files recursively, and create pseudo resources for sql files
func (w *Workspace) setModfileExists() {
	modFile, err := FindModFilePath(w.Path)
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

// FindModFilePath looks in the current folder for mod.sp
// if not found it looks in the parent folder - right up to the root
func FindModFilePath(folder string) (string, error) {
	folder, err := filepath.Abs(folder)
	if err != nil {
		return "", err
	}
	for _, modFilePath := range filepaths.ModFilePaths(folder) {
		_, err = os.Stat(modFilePath)
		if err == nil {
			// found the modfile
			return modFilePath, nil
		}
	}

	// if the file wasn't found, search in the parent directory
	parent := filepath.Dir(folder)
	if folder == parent {
		// this typically means that we are already in the root directory
		return "", ErrorNoModDefinition
	}
	return FindModFilePath(filepath.Dir(folder))
}

func HomeDirectoryModfileCheck(ctx context.Context, workspacePath string) error {
	// bypass all the checks if ConfigKeyBypassHomeDirModfileWarning is set - it means home dir modfile check
	// has already happened before
	if viper.GetBool(constants.ConfigKeyBypassHomeDirModfileWarning) {
		return nil
	}
	// get the cmd and home dir
	cmd := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	home, _ := os.UserHomeDir()

	var modFileExists bool
	for _, modFilePath := range filepaths.ModFilePaths(workspacePath) {
		if _, err := os.Stat(modFilePath); err == nil {
			modFileExists = true
		}
	}

	// check if your workspace path is home dir and if modfile exists
	if workspacePath == home && modFileExists {
		// for interactive query - ask for confirmation to continue
		if cmd.Name() == "query" && viper.GetBool(constants.ConfigKeyInteractive) {
			confirm, err := utils.UserConfirmation(ctx, fmt.Sprintf("%s: You have a mod.sp file in your home directory. This is not recommended.\nAs a result, steampipe will try to load all the files in home and its sub-directories, which can cause performance issues.\nBest practice is to put mod.sp files in their own directories.\nDo you still want to continue? (y/n)", color.YellowString("Warning")))
			if err != nil {
				return err
			}
			if !confirm {
				return sperr.New("failed to load workspace: execution cancelled")
			}
			return nil
		}

		// for batch query mode - if output is table, just warn
		if task.IsBatchQueryCmd(cmd, viper.GetStringSlice(constants.ConfigKeyActiveCommandArgs)) && cmdconfig.Viper().GetString(constants.ArgOutput) == constants.OutputFormatTable {
			error_helpers.ShowWarning("You have a mod.sp file in your home directory. This is not recommended.\nAs a result, steampipe will try to load all the files in home and its sub-directories, which can cause performance issues.\nBest practice is to put mod.sp files in their own directories.\nHit Ctrl+C to stop.\n")
			return nil
		}

		// for other cmds - if home dir has modfile, just warn
		error_helpers.ShowWarning("You have a mod.sp file in your home directory. This is not recommended.\nAs a result, steampipe will try to load all the files in home and its sub-directories, which can cause performance issues.\nBest practice is to put mod.sp files in their own directories.\nHit Ctrl+C to stop.\n")
	}

	return nil
}

func (w *Workspace) LoadWorkspaceMod(ctx context.Context, inputVariables *modconfig.ModVariableMap) error_helpers.ErrorAndWarnings {
	var errorsAndWarnings = error_helpers.ErrorAndWarnings{}

	// build run context which we use to load the workspace
	parseCtx, err := w.getParseContext(ctx, inputVariables)
	if err != nil {
		errorsAndWarnings.Error = err
		return errorsAndWarnings
	}

	// do not reload variables as we already have them
	parseCtx.BlockTypeExclusions = []string{modconfig.BlockTypeVariable}

	// load the workspace mod
	m, otherErrorAndWarning := steampipeconfig.LoadMod(ctx, w.Path, parseCtx)
	errorsAndWarnings.Merge(otherErrorAndWarning)
	if errorsAndWarnings.Error != nil {
		return errorsAndWarnings
	}

	// now set workspace properties
	// populate the mod references map references
	m.ResourceMaps.PopulateReferences()
	// set the mod
	w.Mod = m
	// set the child mods
	w.Mods = parseCtx.GetTopLevelDependencyMods()
	// NOTE: add in the workspace mod to the dependency mods
	w.Mods[w.Mod.Name()] = w.Mod

	// verify all runtime dependencies can be resolved
	errorsAndWarnings.Error = w.verifyResourceRuntimeDependencies()
	return errorsAndWarnings
}

func (w *Workspace) PopulateVariables(ctx context.Context) (*modconfig.ModVariableMap, error_helpers.ErrorAndWarnings) {
	log.Printf("[TRACE] Workspace.PopulateVariables")
	// resolve values of all input variables
	// we WILL validate missing variables when loading
	validateMissing := true
	inputVariables, errorsAndWarnings := w.getInputVariables(ctx, validateMissing)
	if errorsAndWarnings.Error != nil {
		// so there was an error - was it missing variables error
		var missingVariablesError steampipeconfig.MissingVariableError
		ok := errors.As(errorsAndWarnings.GetError(), &missingVariablesError)
		// if there was an error which is NOT a MissingVariableError, return it
		if !ok {
			return nil, errorsAndWarnings
		}
		// if there are missing transitive dependency variables, fail as we do not prompt for these
		if len(missingVariablesError.MissingTransitiveVariables) > 0 {
			return nil, errorsAndWarnings
		}
		// if interactive input is disabled, return the missing variables error
		if !viper.GetBool(constants.ArgInput) {
			return nil, error_helpers.NewErrorsAndWarning(missingVariablesError)
		}
		// so we have missing variables - prompt for them
		// first hide spinner if it is there
		statushooks.Done(ctx)
		if err := promptForMissingVariables(ctx, missingVariablesError.MissingVariables, w.Path); err != nil {
			log.Printf("[TRACE] Interactive variables prompting returned error %v", err)
			return nil, error_helpers.NewErrorsAndWarning(err)
		}

		// now try to load vars again
		inputVariables, errorsAndWarnings = w.getInputVariables(ctx, validateMissing)
		if errorsAndWarnings.Error != nil {
			return nil, errorsAndWarnings
		}

	}
	// populate the parsed variable values
	w.VariableValues, errorsAndWarnings.Error = inputVariables.GetPublicVariableValues()

	return inputVariables, errorsAndWarnings
}

func (w *Workspace) getInputVariables(ctx context.Context, validateMissing bool) (*modconfig.ModVariableMap, error_helpers.ErrorAndWarnings) {
	log.Printf("[TRACE] Workspace.getInputVariables")
	// build a run context just to use to load variable definitions
	variablesParseCtx, err := w.getParseContext(ctx, nil)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// load variable definitions
	variableMap, err := steampipeconfig.LoadVariableDefinitions(ctx, w.Path, variablesParseCtx)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	log.Printf("[INFO] loaded variable definitions: %s", variableMap)

	// get the values
	return steampipeconfig.GetVariableValues(variablesParseCtx, variableMap, validateMissing)
}

// build options used to load workspace
// set flags to create pseudo resources and a default mod if needed
func (w *Workspace) getParseContext(ctx context.Context, variables *modconfig.ModVariableMap) (*parse.ModParseContext, error) {
	parseFlag := parse.CreateDefaultMod
	if w.loadPseudoResources {
		parseFlag |= parse.CreatePseudoResources
	}
	workspaceLock, err := w.loadWorkspaceLock(ctx)
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
			Include: filehelpers.InclusionsFromExtensions(constants.ModDataExtensions),
		})

	// add any evaluated variables to the context
	if variables != nil {
		parseCtx.AddInputVariableValues(variables)
	}

	return parseCtx, nil
}

// load the workspace lock, migrating it if necessary
func (w *Workspace) loadWorkspaceLock(ctx context.Context) (*versionmap.WorkspaceLock, error) {
	workspaceLock, err := versionmap.LoadWorkspaceLock(ctx, w.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load installation cache from %s: %s", w.Path, err)
	}

	// if this is the old format, migrate by reinstalling dependencies
	if workspaceLock.StructVersion() != versionmap.WorkspaceLockStructVersion {
		opts := &modinstaller.InstallOpts{WorkspaceMod: w.Mod}
		installData, err := modinstaller.InstallWorkspaceDependencies(ctx, opts)
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

func (w *Workspace) loadWorkspaceResourceName(ctx context.Context) (*modconfig.WorkspaceResources, error) {
	// build options used to load workspace
	parseCtx, err := w.getParseContext(ctx, nil)
	if err != nil {
		return nil, err
	}

	workspaceResourceNames, err := steampipeconfig.LoadModResourceNames(ctx, w.Mod, parseCtx)
	if err != nil {
		return nil, err
	}

	return workspaceResourceNames, nil
}

func (w *Workspace) verifyResourceRuntimeDependencies() error {
	for _, d := range w.Mod.ResourceMaps.Dashboards {
		if err := d.ValidateRuntimeDependencies(w); err != nil {
			return err
		}
	}
	return nil
}
