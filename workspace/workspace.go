package workspace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/turbot/steampipe/steampipeconfig/parse"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

type Workspace struct {
	Path string
	Mod  *modconfig.Mod

	// maps of mod resources from this mod and ALL DEPENDENCIES, keyed by long and short names
	QueryMap        map[string]*modconfig.Query
	ControlMap      map[string]*modconfig.Control
	ControlGroupMap map[string]*modconfig.ControlGroup

	watcher    *utils.FileWatcher
	loadLock   sync.Mutex
	exclusions []string
	// should we load/watch files recursively
	listFlag     filehelpers.ListFlag
	watcherError error
}

func Load(workspacePath string) (*Workspace, error) {
	// create shell workspace
	workspace := &Workspace{Path: workspacePath}

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

func (w *Workspace) Close() {
	if w.watcher != nil {
		w.watcher.Close()
	}
}

func (w *Workspace) GetNamedQueryMap() map[string]*modconfig.Query {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	return w.QueryMap
}

func (w *Workspace) GetNamedQuery(queryName string) (*modconfig.Query, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// if the name starts with 'local', remove the prefix and try to resolve the short name
	queryName = strings.TrimPrefix(queryName, "local.")

	if query, ok := w.QueryMap[queryName]; ok {
		return query, true
	}

	return nil, false
}

// GetControlsForArg :: resolve the arg into one or more controls
func (w *Workspace) GetControlsForArg(arg string) []*modconfig.Control {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// if arg is in fact a controlGroup,  get all controls underneath the control group
	name, err := modconfig.ParseResourceName(arg)
	if err != nil {
		return nil
	}
	if name.ItemType == modconfig.BlockTypeControlGroup {
		// look in the workspace control group map for this control group
		if controlGroup, ok := w.ControlGroupMap[arg]; ok {
			return controlGroup.GetChildControls()
		}
		return nil
	}

	// check whether the arg is a control name (removing a 'local' prefix if there is one)
	if control, ok := w.ControlMap[strings.TrimPrefix(arg, "local.")]; ok {
		return []*modconfig.Control{control}
	}

	// so arg is not a control group or a control name - check the following possible scopes:
	// 1) 'all' - all controls from all mods
	// 2) '<modName>' - all controls from mod <modName>
	var result []*modconfig.Control
	// the workspace resource maps have duplicate entries, keyed by long and short name.
	// keep track of which controls we have identified in order to avoid dupes
	controlsMatched := make(map[string]bool)
	for _, c := range w.ControlMap {
		if _, alreadyMatched := controlsMatched[c.Name()]; !alreadyMatched {
			if arg == "all" || arg == c.GetMetadata().ModShortName {
				controlsMatched[c.Name()] = true
				result = append(result, c)
			}
		}
	}
	return result
}

func (w *Workspace) loadMod() error {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

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
	w.ControlGroupMap = make(map[string]*modconfig.ControlGroup)

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
	w.ControlGroupMap = w.buildControlGroupMap(modMap)

	return nil
}

// load all dependencies of workspace mod
// used to load all mods in a workspace, using the workspace manifest mod
func (w *Workspace) loadModDependencies(modsFolder string) (modconfig.ModMap, error) {
	var res = modconfig.ModMap{}
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

func (w *Workspace) buildControlGroupMap(modMap modconfig.ModMap) map[string]*modconfig.ControlGroup {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.ControlGroup)

	// for LOCAL controls, add map entries keyed by both short name: control_group.<shortName> and  long name: <modName>.control_group.<shortName?
	for _, c := range w.Mod.ControlGroups {
		res[c.Name()] = c
		res[c.QualifiedName()] = c
	}

	// for mod dependencies, add queries keyed by long name only
	for _, mod := range modMap {
		for _, c := range mod.ControlGroups {
			res[c.QualifiedName()] = c
		}
	}
	return res
}

func (w *Workspace) SetupWatcher(client *db.Client) error {

	watcherOptions := &utils.WatcherOptions{
		Directories: []string{w.Path},
		Include:     filehelpers.InclusionsFromExtensions(steampipeconfig.GetModFileExtensions()),
		Exclude:     w.exclusions,
		OnChange: func(ev fsnotify.Event) {
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

func (w *Workspace) GetResourceMaps() *modconfig.WorkspaceResourceMaps {
	workspaceMap := &modconfig.WorkspaceResourceMaps{
		ModMap:          make(map[string]*modconfig.Mod),
		QueryMap:        w.QueryMap,
		ControlMap:      w.ControlMap,
		ControlGroupMap: w.ControlGroupMap,
	}
	// TODO add in all mod dependencies
	if !w.Mod.IsDefaultMod() {
		workspaceMap.ModMap[w.Mod.Name()] = w.Mod
	}

	return workspaceMap
}
