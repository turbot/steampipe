package workspace

import (
	"fmt"
	"os"

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
	_, err := os.Stat(workspacePath)
	if err != nil {
		return nil, err
	}

	// create shell workspace
	workspace := &Workspace{Path: workspacePath}

	// parse all hcl files under the workspace and either parse or create a mod
	// it is valid for 0 or 1 mod to be defined (if no mod is defined, create a default one)
	// populate mod with all hcl resources we parse
	workspace.Mod, err = steampipeconfig.LoadMod(workspacePath)
	if err != nil {
		return nil, err
	}

	// now

	// now load all mods in the workspace
	modsPath := constants.WorkspaceModPath(workspacePath)
	modMap, err := mod.LoadModDependencies(workspace.Mod, modsPath)
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
