package filepaths

import (
	"fmt"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/error_helpers"
	"os"
	"path/filepath"
)

// PipesInstallDir is the location of config files commen between pipelings
// this must be set by the application at startup
var PipesInstallDir = ""

func ensurePipesInstallSubDir(dirName string) string {
	subDir := installPipesSubDir(dirName)

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		err = os.MkdirAll(subDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("could not create %s directory", dirName))
	}

	return subDir
}

func installPipesSubDir(dirName string) string {
	if PipesInstallDir == "" {
		panic(fmt.Errorf("cannot call any pipes directory functions before PipesInstallDir is set", app_specific.AppName))
	}
	return filepath.Join(app_specific.InstallDir, dirName)
}

// EnsurePipesInternalDir returns the path to the pipes internal directory (creates if missing)
func EnsurePipesInternalDir() string {
	return ensurePipesInstallSubDir("internal")
}
