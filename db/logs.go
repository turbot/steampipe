package db

import (
	"io/ioutil"
	"log"
	"os"
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
		diff := time.Now().Sub(file.ModTime()).Hours()
		if diff > logRetentionDays*24 {
			logPath := fileLocation + "/" + fileName
			err := os.Remove(logPath)
			if err != nil {
				log.Printf("[INFO] failed to delete log file %s\n", logPath)
			}
		}
	}
}
