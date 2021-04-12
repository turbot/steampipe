package workspace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	filehelpers "github.com/turbot/go-kit/files"

	"github.com/fsnotify/fsnotify"
	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type Workspace struct {
	Path          string
	Mod           *modconfig.Mod
	namedQueryMap map[string]*modconfig.Query
	watcher       *utils.FileWatcher
	loadLock      sync.Mutex
	exclusions    []string
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

	if err := workspace.setupWatcher(); err != nil {
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

func (w *Workspace) loadMod() error {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// parse all hcl files under the workspace and either parse or create a mod
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

	w.namedQueryMap = w.buildNamedQueryMap(modMap)

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

func (w *Workspace) buildNamedQueryMap(modMap modconfig.ModMap) map[string]*modconfig.Query {
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
		Path:           w.Path,
		DirExclusions:  []string{},
		FileInclusions: filehelpers.InclusionsFromExtensions(steampipeconfig.GetModFileExtensions()),
		FileExclusions: w.exclusions,
		OnFileChange: func(ev fsnotify.Event) {
			// ignore rename and chmod
			//if ev.Op == fsnotify.Create || ev.Op == fsnotify.Remove || ev.Op == fsnotify.Write {
			w.loadMod()
			//}
		},
		//OnError:          nil,
	})
	if err != nil {
		return err
	}
	w.watcher = watcher

	return nil
}

func (w *Workspace) LoadExclusions() error {
	// add in a hard coded exclusion to the data directory (.steampipe)
	w.exclusions = []string{fmt.Sprintf("**/%s/*", constants.WorkspaceDataDir)}

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
			// add exclusion to the workspace path (to ensure relative pattenrs work)
			absoluteExclusion := filepath.Join(w.Path, line)
			w.exclusions = append(w.exclusions, absoluteExclusion)

		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}
