package constants

import (
	"path"
)

// mod related constants
const (
	WorkspaceDataDir        = ".steampipe"
	WorkspaceModDir         = "mods"
	WorkspaceConfigFileName = "workspace.spc"
	WorkspaceIgnoreFile     = ".steampipeignore"
	WorkspaceDefaultModName = "local"
	WorkspaceModFileName    = "mod.sp"
	DefaultVarsFileName     = "steampipe.spvars"
	MaxControlRunAttempts   = 2
)

func WorkspaceModPath(workspacePath string) string {
	return path.Join(workspacePath, WorkspaceDataDir, WorkspaceModDir)
}
func DefaultVarsFilePath(workspacePath string) string {
	return path.Join(workspacePath, DefaultVarsFileName)
}
