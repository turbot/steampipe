package db

import (
	"errors"
	"fmt"
	"log"

	"github.com/turbot/steampipe-plugin-sdk/logging"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/definitions/results"
	"github.com/turbot/steampipe/utils"
)

// ExecuteQuery :: entry point for executing ad-hoc queries from outside the package
func ExecuteQuery(queryString string) (*results.ResultStreamer, error) {
	var err error

	logging.LogTime("db.ExecuteQuery start")
	log.Println("[TRACE] db.ExecuteQuery start")

	EnsureDBInstalled()
	status, err := GetStatus()
	if err != nil {
		return nil, errors.New("could not retrieve service status")
	}

	if status != nil && status.Invoker == InvokerQuery {
		return nil, fmt.Errorf("You already have a %s session open. To run multiple sessions, first run %s.\nTo kill existing sessions run %s", constants.Bold("steampipe query"), constants.Bold("steampipe service start"), constants.Bold("steampipe service stop --force"))
	}

	if status == nil {
		// the db service is not started - start it
		StartService(InvokerQuery)
	}

	client, err := GetClient(false)
	utils.FailOnErrorWithMessage(err, "client failed to initialize")

	// refresh connections
	connectionsUpdated, err := client.RefreshConnections()
	if err != nil {
		// shutdown the service if something went wrong!!!
		Shutdown(client, InvokerQuery)
		return nil, fmt.Errorf("failed to refresh connections: %v", err.Error())
	}
	if err = refreshFunctions(); err != nil {
		// shutdown the service if something went wrong!!!
		Shutdown(client, InvokerQuery)
		return nil, fmt.Errorf("failed to add functions: %v", err)
	}

	// workaround: if connections were updated, create a new client so that any change in the search patch is reflected
	if connectionsUpdated {
		clientSingleton = nil
		client, err = GetClient(false)
		utils.FailOnErrorWithMessage(err, "client failed to reinitialize")
	}
	resultsStreamer := results.NewResultStreamer()

	// this is a callback to close the db et-al. when things get done - no matter the mode
	onComplete := func() { Shutdown(client, InvokerQuery) }

	if queryString == "" {
		interactiveClient, err := newInteractiveClient(client)
		if err != nil {
			utils.ShowErrorWithMessage(err, "interactive client failed to initialize")
			onComplete()
			return nil, err
		}

		// start the interactive prompt in a go routine
		go interactiveClient.InteractiveQuery(resultsStreamer, onComplete)
	} else {
		result, err := client.executeQuery(queryString, false)
		if err != nil {
			onComplete()
			return nil, err
		}
		go resultsStreamer.StreamSingleResult(result, onComplete)
	}

	logging.LogTime("db.ExecuteQuery end")
	return resultsStreamer, nil
}

// Shutdown :: closes the client connection and stops the
// database instance if the given `invoker` matches
func Shutdown(client *Client, invoker Invoker) {
	log.Println("[TRACE] shutdown")
	if client != nil {
		client.close()
	}

	status, _ := GetStatus()

	// force stop if the service was invoked by the same invoker and we are the last one
	if status != nil && status.Invoker == invoker {
		status, err := StopDB(false, invoker)
		if err != nil {
			utils.ShowError(err)
		}
		if status != ServiceStopped {
			StopDB(true, invoker)
		}
	}
}
