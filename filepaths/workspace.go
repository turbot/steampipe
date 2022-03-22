package filepaths

import (
	"path"
	"path/filepath"
)

// mod related constants
const (
	WorkspaceDataDir        = ".steampipe"
	WorkspaceModDir         = "mods"
	WorkspaceConfigFileName = "workspace.spc"
	WorkspaceIgnoreFile     = ".steampipeignore"
	ModFileName             = "mod.sp"
	DefaultVarsFileName     = "steampipe.spvars"
	WorkspaceLockFileName   = ".mod.cache.json"
)

func WorkspaceModPath(workspacePath string) string {
	return path.Join(workspacePath, WorkspaceDataDir, WorkspaceModDir)
}

func WorkspaceLockPath(workspacePath string) string {
	return path.Join(workspacePath, WorkspaceLockFileName)
}

func DefaultVarsFilePath(workspacePath string) string {
	return path.Join(workspacePath, DefaultVarsFileName)
}

func ModFilePath(modFolder string) string {
	modFilePath := filepath.Join(modFolder, ModFileName)
	return modFilePath
}
