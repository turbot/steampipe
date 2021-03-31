package workspace

import (
	"fmt"
	"os"
	"strings"
	"sync"

	filehelpers "github.com/turbot/go-kit/files"

	"github.com/fsnotify/fsnotify"
	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/mod"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// .gitignore format exclusion string for workspace .steampipe directory
var workspaceDataDirExclusion = []string{fmt.Sprintf("**/%s*", constants.WorkspaceDataDir)}

type Workspace struct {
	Path          string
	Mod           *modconfig.Mod
	namedQueryMap map[string]*modconfig.Query
	watcher       *utils.FileWatcher
	loadLock      sync.Mutex
}

func Load() (*Workspace, error) {
	// workspace is always the working directory
	workspacePath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// create shell workspace
	workspace := &Workspace{Path: workspacePath}
	if err := workspace.loadMod(); err != nil {
		return nil, err
	}

	if err := workspace.setupWatcher(); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (w *Workspace) loadMod() error {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// parse all hcl files under the workspace and either parse or create a mod
	// it is valid for 0 or 1 mod to be defined (if no mod is defined, create a default one)
	// populate mod with all hcl resources we parse
	// pass flag to create pseudo resources and default mod
	opts := w.getWorkspaceLoadOptions()
	m, err := steampipeconfig.LoadMod(w.Path, opts)
	if err != nil {
		return err
	}
	w.Mod = m

	// now load all mods in the workspace
	modsPath := constants.WorkspaceModPath(w.Path)
	modMap, err := mod.LoadModDependencies(w.Mod, modsPath)
	if err != nil {
		return err
	}

	w.namedQueryMap = w.buildNamedQueryMap(modMap)

	// TODO validate unique aliases

	// TODO load workspace config

	return nil

}

// build options used to load workspace
// ignore .steampipe folder
// TODO load .gitignore
// set flags to create pseudo resources and a default mod if needed
func (w *Workspace) getWorkspaceLoadOptions() *steampipeconfig.LoadModOptions {
	return &steampipeconfig.LoadModOptions{
		Exclude: workspaceDataDirExclusion,
		Flags:   steampipeconfig.CreatePseudoResources | steampipeconfig.CreateDefaultMod,
	}
}

func (w *Workspace) GetNamedQueryMap() map[string]*modconfig.Query {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	return w.namedQueryMap
}

func (w *Workspace) GetNamedQuery(queryName string) (*modconfig.Query, bool) {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// if the name starts with 'local', remove the prefix and try to resolve the short name
	queryName = strings.TrimPrefix(queryName, "local.")

	if namedQuery, ok := w.namedQueryMap[queryName]; ok {
		return namedQuery, true
	}

	return nil, false
}

func (w *Workspace) buildNamedQueryMap(modMap mod.ModMap) map[string]*modconfig.Query {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.Query)

	// add local queries by short name: query.xxxx and long name: <workspace>.query.xxxx
	for _, q := range w.Mod.Queries {
		shortName := fmt.Sprintf("query.%s", q.Name)
		res[shortName] = q
	}
	// add queries from mod dependencies by FQN
	for _, mod := range modMap {
		for _, q := range mod.Queries {
			longName := fmt.Sprintf("%s.query.%s", mod.Name, q.Name)
			res[longName] = q
		}
	}
	return res
}

func (w *Workspace) setupWatcher() error {
	watcher, err := utils.NewWatcher(&utils.WatcherOptions{
		Path:             w.Path,
		FolderExclusions: workspaceDataDirExclusion,
		FileInclusions:   filehelpers.InclusionsFromExtensions(steampipeconfig.GetModFileExtensions()),
		OnChange:         func(fsnotify.Event) { w.loadMod() },
		//OnError:          nil,
	})
	if err != nil {
		return err
	}
	w.watcher = watcher

	return nil
}

func (w *Workspace) Close() {
	if w.watcher != nil {
		w.watcher.Close()
	}
}
