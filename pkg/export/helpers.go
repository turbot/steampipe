package export

import (
	"fmt"
	"io"
	"os"
	"time"
)

func GenerateDefaultExportFileName(executionName, fileExtension string) string {
	now := time.Now()
	timeFormatted := fmt.Sprintf("%d%02d%02dT%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	return fmt.Sprintf("%s.%s%s", executionName, timeFormatted, fileExtension)
}

func Write(filePath string, exportData io.Reader) error {
	// Write to temp file first for atomic operation
	tempPath := filePath + ".tmp"
	destination, err := os.Create(tempPath)
	if err != nil {
		return err
	}

	// Copy data to temp file
	_, err = io.Copy(destination, exportData)
	destination.Close()

	if err != nil {
		// Clean up temp file on error
		os.Remove(tempPath)
		return err
	}

	// Atomic rename - either succeeds completely or not at all
	return os.Rename(tempPath, filePath)
}
