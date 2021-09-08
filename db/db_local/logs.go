package db_local

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const logRetentionDays = 7

func TrimLogs() {
	fileLocation := getDatabaseLogDirectory()
	files, err := ioutil.ReadDir(fileLocation)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fileName := file.Name()
		if filepath.Ext(fileName) != ".log" {
			continue
		}
		age := time.Now().Sub(file.ModTime()).Hours()
		if age > logRetentionDays*24 {
			logPath := filepath.Join(fileLocation, fileName)
			err := os.Remove(logPath)
			if err != nil {
				log.Printf("[TRACE] failed to delete log file %s\n", logPath)
			}
		}
	}
}
