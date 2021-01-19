package db

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/turbot/steampipe-plugin-sdk/logging"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// ExecuteQuery :: entry point for executing ad-hoc queries from outside the package
func ExecuteQuery(queryString string) (*ResultStreamer, error) {
	var err error

	logging.LogTime("db.ExecuteQuery start")
	log.Println("[TRACE] db.ExecuteQuery start")

	if createAnInstanceFile() != nil {
		return nil, errors.New("could not create lock")
	}

	EnsureDBInstalled()
	status, err := GetStatus()
	if err != nil {
		return nil, errors.New("could not retrieve service status")
	}

	if status == nil {
		// the db service is not started - start it
		StartService(InvokerQuery)
	}

	client, err := GetClient(false)
	utils.FailOnErrorWithMessage(err, "client failed to initialize")

	// refresh connections
	if err = refreshConnections(client); err != nil {
		// shutdown the service if something went wrong!!!
		shutdown(client)
		return nil, fmt.Errorf("failed to refresh connections: %v", err)
	}

	resultsStreamer := newQueryResults()
	if queryString == "" {
		interactiveClient, err := newInteractiveClient(client)
		utils.FailOnErrorWithMessage(err, "interactive client failed to initialize")

		// start the interactive prompt in a go routine
		go interactiveClient.InteractiveQuery(resultsStreamer)
	} else {
		result, err := client.executeQuery(queryString)
		if err != nil {
			return nil, err
		}
		// send a single result to the streamer - this will close the channel afterwards
		// pass an onComplete callback function to shutdown the db
		onComplete := func() { shutdown(client) }
		go resultsStreamer.streamSingleResult(result, onComplete)
	}

	logging.LogTime("db.ExecuteQuery end")
	return resultsStreamer, nil
}

func shutdown(client *Client) {
	log.Println("[TRACE] shutdown")
	defer removeAnInstanceFile()
	if client != nil {
		client.close()
	}

	status, _ := GetStatus()

	// force stop if invoked by `query` and we are the last one
	if status.Invoker == InvokerQuery && amITheLastQueryInstance() {
		_, err := StopDB(true)
		if err != nil {
			utils.ShowError(err)
		}
	}
}

func createAnInstanceFile() error {
	// create a file called `query~uuidv4.lck` in internal
	lockFile := filepath.Join(constants.InternalDir(), fmt.Sprintf("query~%s.lck", uuid.New().String()))
	return ioutil.WriteFile(lockFile, []byte(""), 0644)
}
func amITheLastQueryInstance() bool {
	// look for a file in `internal` with the name `query~uuidv4.lck` and return true if count is 1
	lockCount := 0

	files, err := ioutil.ReadDir(constants.InternalDir())
	if err != nil {
		return false
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "query~") && strings.HasSuffix(file.Name(), "lck") {
			// this is a lock file
			// increment and continue
			lockCount++
			if lockCount > 1 {
				// if we encounter a second lockfile,
				// then obviously we are not the last
				// no point continuing
				return false
			}
		}
	}

	return true
}
func removeAnInstanceFile() error {
	// look for a file in `internal` with the name `query~uuidv4.lck` and remove it
	// doesn't need to be the one we created
	files, err := ioutil.ReadDir(constants.InternalDir())
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "query~") && strings.HasSuffix(file.Name(), "lck") {
			// this is a lock file
			// remove it and get out
			return os.Remove(filepath.Join(constants.InternalDir(), file.Name()))
		}
	}

	return nil
}
