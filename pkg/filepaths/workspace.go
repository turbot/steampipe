package filepaths

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/turbot/steampipe/pkg/constants/runtime"
)

var ModFileNames = []string{"mod.sp", "mod.pp"}

// mod related constants
const (
	WorkspaceDataDir            = ".steampipe"
	WorkspaceModDir             = "mods"
	WorkspaceModShadowDirPrefix = ".mods."
	WorkspaceConfigFileName     = "workspace.spc"
	WorkspaceIgnoreFile         = ".steampipeignore"
	DefaultVarsFileName         = "steampipe.spvars"
	WorkspaceLockFileName       = ".mod.cache.json"
)

func WorkspaceModPath(workspacePath string) string {
	return path.Join(workspacePath, WorkspaceDataDir, WorkspaceModDir)
}

func WorkspaceModShadowPath(workspacePath string) string {
	return path.Join(workspacePath, WorkspaceDataDir, fmt.Sprintf("%s%s", WorkspaceModShadowDirPrefix, runtime.ExecutionID))
}

func IsModInstallShadowPath(dirName string) bool {
	return strings.HasPrefix(dirName, WorkspaceModShadowDirPrefix)
}

func WorkspaceLockPath(workspacePath string) string {
	return path.Join(workspacePath, WorkspaceLockFileName)
}

func DefaultVarsFilePath(workspacePath string) string {
	return path.Join(workspacePath, DefaultVarsFileName)
}

func DefaultModFilePath(modFolder string) string {
	return filepath.Join(modFolder, ModFileNames[0])
}

func ModFilePaths(modFolder string) []string {
	var modFilePaths []string

	for _, modFileName := range ModFileNames {
		modFilePaths = append(modFilePaths, filepath.Join(modFolder, modFileName))

	}
	return modFilePaths
}
