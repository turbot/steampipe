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
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

type Workspace struct {
	Path string
	Mod  *modconfig.Mod

	// maps of mod resources from this mod and ALL DEPENDENCIES, keyed by long and short names
	queryMap        map[string]*modconfig.Query
	controlMap      map[string]*modconfig.Control
	controlGroupMap map[string]*modconfig.ControlGroup

	watcher    *utils.FileWatcher
	loadLock   sync.Mutex
	exclusions []string
}

func Load(workspacePath string) (*Workspace, error) {
	// create shell workspace
	workspace := &Workspace{Path: workspacePath}

	// load the .steampipe ignore file
	if err := workspace.LoadExclusions(); err != nil {
		return nil, err
	}

	if err := workspace.loadMod(); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (w *Workspace) Close() {
	if w.watcher != nil {
		w.watcher.Close()
	}
}

func (w *Workspace) GetNamedQueryMap() map[string]*modconfig.Query {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	return w.queryMap
}

func (w *Workspace) GetNamedQuery(queryName string) (*modconfig.Query, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// if the name starts with 'local', remove the prefix and try to resolve the short name
	queryName = strings.TrimPrefix(queryName, "local.")

	if query, ok := w.queryMap[queryName]; ok {
		return query, true
	}

	return nil, false
}

func (w *Workspace) GetControls(controlName string) ([]*modconfig.Control, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// if the name starts with 'local', remove the prefix and try to resolve the short name
	controlName = strings.TrimPrefix(controlName, "local.")

	// if controlName is in fact a controlgroup,  get all controls underneath thje control group
	name, err := modconfig.ParseModResourceName(controlName)
	if err != nil {
		return nil, false
	}

	switch name.ItemType {
	case modconfig.BlockTypeControl:
		// look in the workspace control map for this control
		if control, ok := w.controlMap[controlName]; ok {
			return []*modconfig.Control{control}, true
		}
	case modconfig.BlockTypeControlGroup:
		// look in the workspace control group map for this control group
		if controlGroup, ok := w.controlGroupMap[controlName]; ok {
			return controlGroup.GetChildControls(), true
		}
	}
	return nil, false
}

func (w *Workspace) loadMod() error {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// parse all hcl files in the workspace folder (non recursively) and either parse or create a mod
	// it is valid for 0 or 1 mod to be defined (if no mod is defined, create a default one)
	// populate mod with all hcl resources we parse

	// build options used to load workspace
	// set flags to create pseudo resources and a default mod if needed
	opts := &steampipeconfig.LoadModOptions{
		Exclude: w.exclusions,
		Flags:   steampipeconfig.CreatePseudoResources | steampipeconfig.CreateDefaultMod,
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

	w.queryMap = w.buildQueryMap(modMap)
	w.controlMap = w.buildControlMap(modMap)
	w.controlGroupMap = w.buildControlGroupMap(modMap)

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

	// for LOCAL queries, add map entries keyed by both short name: query.xxxx and  long name: <workspace>.query.xxxx
	for _, q := range w.Mod.Queries {
		shortName := fmt.Sprintf("query.%s", types.SafeString(q.ShortName))
		res[shortName] = q
	}

	// for mode dependencies, add queries keyed by long name only
	for _, mod := range modMap {
		for _, q := range mod.Queries {
			longName := fmt.Sprintf("%s.query.%s", types.SafeString(mod.Name), types.SafeString(q.ShortName))
			res[longName] = q
		}
	}
	return res
}

func (w *Workspace) buildControlMap(modMap modconfig.ModMap) map[string]*modconfig.Control {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.Control)

	// for LOCAL controls, add map entries keyed by both short name: query.xxxx and  long name: <workspace>.query.xxxx
	for _, c := range w.Mod.Controls {
		shortName := fmt.Sprintf("control.%s", types.SafeString(c.ShortName))
		res[shortName] = c
	}

	// for mode dependencies, add queries keyed by long name only
	for _, mod := range modMap {
		for _, c := range mod.Controls {
			longName := fmt.Sprintf("%s.control.%s", types.SafeString(mod.Name), types.SafeString(c.ShortName))
			res[longName] = c
		}
	}
	return res
}

func (w *Workspace) buildControlGroupMap(modMap modconfig.ModMap) map[string]*modconfig.ControlGroup {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.ControlGroup)

	// for LOCAL controls, add map entries keyed by both short name: query.xxxx and  long name: <workspace>.query.xxxx
	for _, c := range w.Mod.ControlGroups {
		shortName := fmt.Sprintf("control_group.%s", types.SafeString(c.Name))
		res[shortName] = c
	}

	// for mode dependencies, add queries keyed by long name only
	for _, mod := range modMap {
		for _, c := range mod.ControlGroups {
			longName := fmt.Sprintf("%s.control_group.%s", types.SafeString(mod.Name), types.SafeString(c.Name))
			res[longName] = c
		}
	}
	return res
}

func (w *Workspace) SetupWatcher() error {
	watcherOptions := &utils.WatcherOptions{
		Directories: []string{w.Path},
		Include:     filehelpers.InclusionsFromExtensions(steampipeconfig.GetModFileExtensions()),
		Exclude:     w.exclusions,
		OnChange: func(ev fsnotify.Event) {
			w.loadMod()
		},
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
