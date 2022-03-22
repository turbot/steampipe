package db_local

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

const logRetentionDays = 7

func TrimLogs() {
	fileLocation := getDatabaseLogDirectory()
	files, err := os.ReadDir(fileLocation)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fi, err := file.Info()
		if err != nil {
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
