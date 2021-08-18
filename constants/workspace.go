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
	WorkspaceDefaultModName = "local"
	MaxControlRunAttempts   = 3
)

func WorkspaceModPath(workspacePath string) string {
	return path.Join(workspacePath, WorkspaceDataDir, WorkspaceModDir)
}
