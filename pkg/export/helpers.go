package export

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func GenerateDefaultExportFileName(executionName, fileExtension string) string {
	now := time.Now()
	timeFormatted := fmt.Sprintf("%d%02d%02dT%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	return fmt.Sprintf("%s.%s%s", executionName, timeFormatted, fileExtension)
}

func Write(filePath string, exportData io.Reader) error {
	// Create a temporary file in the same directory as the target file
	// This ensures the temp file is on the same filesystem for atomic rename
	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, ".steampipe-export-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup of temp file on failure
	defer func() {
		tmpFile.Close()
		// If we still have a temp file at this point, remove it
		// (successful path will have already renamed it)
		os.Remove(tmpPath)
	}()

	// Write data to temp file
	_, err = io.Copy(tmpFile, exportData)
	if err != nil {
		return err
	}

	// Ensure all data is written to disk
	if err := tmpFile.Sync(); err != nil {
		return err
	}

	// Close the temp file before renaming
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Atomically move temp file to final destination
	// This is atomic on POSIX systems and will not leave partial files
	return os.Rename(tmpPath, filePath)
}
