package constants

import (
	"path"
)

// mod related constants
const (
	WorkspaceModDir         = "mods"
	WorkspaceDataDir        = ".steampipe"
	WorkspaceConfigFileName = "workspace.spc"
	WorkspaceIgnoreFile     = ".steampipeignore"
)

func WorkspaceModPath(workspacePath string) string {
	return path.Join(workspacePath, WorkspaceDataDir, WorkspaceModDir)
}
