package ociinstaller

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

type tempDir struct {
	Path string
}

// NewTempDir :: returns the temp directory, creating it if it does not exist
func NewTempDir(path string) *tempDir {
	return &tempDir{
		Path: getOrCreateTempDir(path),
	}
}

func getOrCreateTempDir(ref string) string {
	pluginCacheDir := filepath.Join(constants.TempDir(), safeDirName(ref))

	if _, err := os.Stat(pluginCacheDir); os.IsNotExist(err) {
		err = os.MkdirAll(pluginCacheDir, 0755)
		utils.FailOnErrorWithMessage(err, "could not create cache directory")
	}

	return pluginCacheDir
}

func (d *tempDir) Delete() error {
	return os.RemoveAll(d.Path)
}

func safeDirName(dirName string) string {
	newName := strings.ReplaceAll(dirName, "/", "_")
	newName = strings.ReplaceAll(newName, ":", "@")

	return newName
}
