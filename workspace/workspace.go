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
	Path        string
	ModManifest *modconfig.Mod
	Mods        mod.ModMap
}

func LoadWorkspace(workspacePath string) (*Workspace, error) {
	// verify workspacePath exists
	if _, err := os.Stat(workspacePath); err != nil {
		return nil, err
	}

	// now load the manifest mod
	manifest, err := steampipeconfig.LoadMod(workspacePath)
	if err != nil {
		return nil, err
	}

	workspace := &Workspace{Path: workspacePath, ModManifest: manifest}

	// now load all mods in the workspace
	modPath := workspace.ModPath()
	modMap, err := mod.LoadModDependencies(manifest, modPath)
	if err != nil {
		return nil, err
	}
	workspace.Mods = modMap

	// TODO LOAD CONFIG
	return workspace, nil

}

func (w *Workspace) ModPath() string {
	return path.Join(w.Path, constants.ModDir)
}
