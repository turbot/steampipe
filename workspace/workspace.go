package workspace

import (
	"os"
	"path"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/mod"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type Workspace struct {
	Path          string
	ModManifest   *modconfig.Mod
	Mods          mod.ModMap
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
	workspace.ModManifest = manifest

	// now load all mods in the workspace
	modPath := workspace.ModPath()
	modMap, err := mod.LoadModDependencies(manifest, modPath)
	if err != nil {
		return nil, err
	}
	workspace.Mods = modMap

	workspace.NamedQueryMap = workspace.Mods.BuildNamedQueryMap()
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
