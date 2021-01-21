package db

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/turbot/steampipe/utils"
)

const days = 14

func TrimLogs() {
	fileLocation := getDatabaseLogDirectory()
	files, err := ioutil.ReadDir(fileLocation)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fileName := file.Name()
		diff := time.Now().Sub(file.ModTime()).Hours()
		if diff > days*24 {
			err := os.Remove(fileLocation + "/" + fileName)
			if err != nil {
				utils.FailOnErrorWithMessage(err, "could not delete the log file")
			}
		}
	}
}
