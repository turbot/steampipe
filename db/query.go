package db

import (
	"errors"
	"fmt"
	"log"

	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/utils"
)

// ExecuteQuery :: entry point for executing ad-hoc queries from outside the package
func ExecuteQuery(queryString string) (*ResultStreamer, error) {
	var didWeStartService bool
	var err error

	logging.LogTime("db.ExecuteQuery start")
	log.Println("[TRACE] db.ExecuteQuery start")

	EnsureDBInstalled()
	status, err := GetStatus()
	if err != nil {
		return nil, errors.New("could not retrieve service status")
	}

	if status == nil {
		// the db service is not started - start it
		StartService(QueryInvoker)
		didWeStartService = true
	}

	client, err := GetClient(false)
	utils.FailOnErrorWithMessage(err, "client failed to initialize")

	// refresh connections
	if err = refreshConnections(client); err != nil {
		// shutdown the service if something went wrong!!!
		shutdown(client, didWeStartService)
		return nil, fmt.Errorf("failed to refresh connections: %v", err)
	}

	resultsStreamer := newQueryResults()
	if queryString == "" {
		interactiveClient, err := newInteractiveClient(client)
		utils.FailOnErrorWithMessage(err, "interactive client failed to initialize")

		// start the interactive prompt in a go routine
		go interactiveClient.InteractiveQuery(resultsStreamer, didWeStartService)
	} else {
		result, err := client.executeQuery(queryString)
		if err != nil {
			return nil, err
		}
		// send a single result to the streamer - this will close the channel afterwards
		// pass an onComplete callback function to shutdown the db
		onComplete := func() { shutdown(client, didWeStartService) }
		go resultsStreamer.streamSingleResult(result, onComplete)
	}

	logging.LogTime("db.ExecuteQuery end")
	return resultsStreamer, nil
}

func shutdown(client *Client, stopService bool) {
	log.Println("[TRACE] shutdown", stopService)
	if client != nil {
		client.close()
	}

	status, _ := GetStatus()

	// force stop
	if stopService || status.Invoker == QueryInvoker {
		_, err := StopDB(true)
		if err != nil {
			utils.ShowError(err)
		}
	}
}
