package db

import (
	"errors"
	"fmt"
	"log"

	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/definitions/results"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

// StartServiceForQuery :: ensure db is installed and start service if necessary
func StartServiceForQuery() error {
	logging.LogTime("db.ExecuteQuery start")
	log.Println("[TRACE] db.ExecuteQuery start")

	EnsureDBInstalled()
	status, err := GetStatus()
	if err != nil {
		return errors.New("could not retrieve service status")
	}

	if status != nil && status.Invoker == InvokerQuery {
		return fmt.Errorf("You already have a %s session open. To run multiple sessions, first run %s.\nTo kill existing sessions run %s", constants.Bold("steampipe query"), constants.Bold("steampipe service start"), constants.Bold("steampipe service stop --force"))
	}

	if status == nil {
		// the db service is not started - start it
		StartService(InvokerQuery)
	}
	return nil
}

// RunInteractivePrompt :: start the interactive query prompt
func RunInteractivePrompt(workspace *workspace.Workspace) (*results.ResultStreamer, error) {
	client, err := NewClient(true)
	if err != nil {
		return nil, err
	}
	resultsStreamer := results.NewResultStreamer()

	interactiveClient, err := newInteractiveClient(client, workspace)
	if err != nil {
		utils.ShowErrorWithMessage(err, "interactive client failed to initialize")
		Shutdown(client, InvokerQuery)
		return nil, err
	}

	// start the interactive prompt in a go routine
	go interactiveClient.InteractiveQuery(resultsStreamer)

	logging.LogTime("db.ExecuteQuery end")
	return resultsStreamer, nil
}

// ExecuteQuery :: execute a single query. If shutdownAfterCompletion is true, shutdown the client after completion
func ExecuteQuery(queryString string, client *Client) (*results.ResultStreamer, error) {
	resultsStreamer := results.NewResultStreamer()

	result, err := client.executeQuery(queryString, false)
	if err != nil {
		return nil, err
	}
	go resultsStreamer.StreamSingleResult(result)

	logging.LogTime("db.ExecuteQuery end")
	return resultsStreamer, nil
}
