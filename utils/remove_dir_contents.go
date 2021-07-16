package utils

import (
	"os"
	"path/filepath"
)

func RemoveDirectoryContents(removePath string) error {
	files, err := filepath.Glob(filepath.Join(removePath, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}
