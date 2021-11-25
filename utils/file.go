package utils

import (
	"os"
	"time"
)

func FileModTime(filePath string) (time.Time, error) {
	file, err := os.Stat(filePath)

	if err != nil {
		return time.Time{}, err
	}

	return file.ModTime(), nil
}
