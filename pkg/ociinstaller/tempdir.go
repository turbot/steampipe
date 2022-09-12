package ociinstaller

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/turbot/steampipe/pkg/error_helpers"
)

type tempDir struct {
	Path string
}

// NewTempDir creates a directory under the given parent directory.
func NewTempDir(parent string) *tempDir {
	return &tempDir{
		Path: getOrCreateTempDir(parent),
	}
}

func getOrCreateTempDir(parent string) string {
	cacheDir := filepath.Join(parent, safeDirName(fmt.Sprintf("tmp-%s", generateTempDirName())))

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create cache directory")
	}
	return cacheDir
}

func (d *tempDir) Delete() error {
	return os.RemoveAll(d.Path)
}

func safeDirName(dirName string) string {
	newName := strings.ReplaceAll(dirName, "/", "_")
	newName = strings.ReplaceAll(newName, ":", "@")

	return newName
}

func generateTempDirName() string {
	u, err := uuid.NewRandom()
	if err != nil {
		// Should never happen?
		panic(err)
	}
	s := u.String()
	return s[9:23]
}
