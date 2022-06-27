package utils

import (
	"os"
)

func EnsureDirectoryPermission(directoryPath string) error {
	// verify that we can read and write to the directory
	tmpFile, err := os.CreateTemp(directoryPath, "tmp")
	if err != nil {
		return err
	}
	tmpFile.Close()
	os.Remove(tmpFile.Name())
	return nil
}
