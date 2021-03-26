package workspace

import (
	"fmt"
	"os"
	"path"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/mod"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type Workspace struct {
	Path          string
	Mod           *modconfig.Mod
	NamedQueryMap map[string]*modconfig.Query
}

func Load(workspacePath string) (*Workspace, error) {
	// verify workspacePath exists
	if _, err := os.Stat(workspacePath); err != nil {
		return nil, err
	}

	workspace := &Workspace{Path: workspacePath}
	// now load the manifest mod
	manifest, err := steampipeconfig.LoadMod(workspacePath)
	if err != nil {
		return nil, err
	}

	if manifest == nil {
		// this is not a workspace folder
		workspace.NamedQueryMap = make(map[string]*modconfig.Query)
		return workspace, nil
	}
	workspace.Mod = manifest

	// now load all mods in the workspace
	modPath := workspace.ModPath()
	modMap, err := mod.LoadModDependencies(manifest, modPath)
	if err != nil {
		return nil, err
	}

	workspace.NamedQueryMap = workspace.buildNamedQueryMap(modMap)

	// TODO validate unique aliases

	// TODO LOAD CONFIG

	return workspace, nil

}

func (w *Workspace) GetNamedQuery(input string) (*modconfig.Query, bool) {
	if namedQuery, ok := w.NamedQueryMap[input]; ok {
		return namedQuery, true
	}
	return nil, false
}

func (w *Workspace) ModPath() string {
	return path.Join(w.Path, constants.ModDir)
}

func (w *Workspace) buildNamedQueryMap(modMap mod.ModMap) map[string]*modconfig.Query {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.Query)

	// add local queries by short name: query.xxxx and long name: <workspace>.query.xxxx
	for _, q := range w.Mod.Queries {
		shortName := fmt.Sprintf("query.%s", q.Name)
		longName := fmt.Sprintf("%s.query.%s", w.Mod.Name, q.Name)

		res[shortName] = q
		res[longName] = q
	}
	// ad queries from mode dependencies by FQN
	for _, mod := range modMap {
		for _, q := range mod.Queries {
			longName := fmt.Sprintf("%s.query.%s", mod.Name, q.Name)
			res[longName] = q
		}
	}
	return res
}
