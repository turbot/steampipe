package db

import (
	"errors"
	"fmt"
	"log"

	"github.com/turbot/steampipe-plugin-sdk/logging"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// ExecuteQuery :: entry point for executing ad-hoc queries from outside the package
func ExecuteQuery(queryString string) (*ResultStreamer, error) {
	var err error

	logging.LogTime("db.ExecuteQuery start")
	log.Println("[TRACE] db.ExecuteQuery start")

	EnsureDBInstalled()
	status, err := GetStatus()
	if err != nil {
		return nil, errors.New("could not retrieve service status")
	}

	if status != nil && status.Invoker == InvokerQuery {
		return nil, fmt.Errorf("You already have a %s session open. To run multiple sessions, first run %s. To kill existing sessions run %s", constants.Bold("steampipe query"), constants.Bold("steampipe service start"), constants.Bold("steampipe service stop --force`"))
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

	// this is a callback to close the db et-al. when things get done - no matter the mode
	onComplete := func() { shutdown(client) }

	if queryString == "" {
		interactiveClient, err := newInteractiveClient(client)
		utils.FailOnErrorWithMessage(err, "interactive client failed to initialize")

		// start the interactive prompt in a go routine
		go interactiveClient.InteractiveQuery(resultsStreamer, onComplete)
	} else {
		result, err := client.executeQuery(queryString)
		if err != nil {
			return nil, err
		}
		// send a single result to the streamer - this will close the channel afterwards
		// pass the onComplete callback function to shutdown the db
		onComplete := func() { shutdown(client) }
		go resultsStreamer.streamSingleResult(result, onComplete)
	}

	logging.LogTime("db.ExecuteQuery end")
	return resultsStreamer, nil
}

func shutdown(client *Client) {
	log.Println("[TRACE] shutdown")
	if client != nil {
		client.close()
	}

	status, _ := GetStatus()

	// force stop if invoked by `query` and we are the last one
	if status.Invoker == InvokerQuery {
		_, err := StopDB(true)
		if err != nil {
			utils.ShowError(err)
		}
	}
}
