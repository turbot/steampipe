package db_local

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

const logRetentionDays = 7

func TrimLogs() {
	fileLocation := filepaths.EnsureLogDir()
	files, err := os.ReadDir(fileLocation)
	if err != nil {
		log.Println("[TRACE] error listing db log directory", err)
	}
	for _, file := range files {
		fi, err := file.Info()
		if err != nil {
			log.Printf("[TRACE] error reading file info of %s. continuing\n", file.Name())
			continue
		}

		fileName := fi.Name()
		if filepath.Ext(fileName) != ".log" {
			continue
		}

		age := time.Since(fi.ModTime()).Hours()
		if age > logRetentionDays*24 {
			logPath := filepath.Join(fileLocation, fileName)
			err := os.Remove(logPath)
			if err != nil {
				log.Printf("[TRACE] failed to delete log file %s\n", logPath)
			}
		}
	}
}
