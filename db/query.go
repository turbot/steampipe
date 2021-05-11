package db

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/definitions/results"
	"github.com/turbot/steampipe/utils"
)

// EnsureDbAndStartService :: ensure db is installed and start service if necessary
func EnsureDbAndStartService(invoker Invoker) error {
	logging.LogTime("db.EnsureDbAndStartService start")
	log.Println("[TRACE] db.EnsureDbAndStartService start")

	EnsureDBInstalled()
	status, err := GetStatus()
	if err != nil {
		return errors.New("could not retrieve service status")
	}

	if status != nil && status.Invoker == invoker {
		return fmt.Errorf("You already have a %s session open. To run multiple sessions, first run %s.\nTo kill existing sessions run %s", constants.Bold("steampipe query"), constants.Bold("steampipe service start"), constants.Bold("steampipe service stop --force"))
	}

	if status == nil {
		// the db service is not started - start it
		StartService(invoker)
	}
	return nil
}

// RunInteractivePrompt :: start the interactive query prompt
func RunInteractivePrompt(workspace NamedQueryProvider, client *Client) (*results.ResultStreamer, error) {
	resultsStreamer := results.NewResultStreamer()

	interactiveClient, err := newInteractiveClient(workspace, client, resultsStreamer)
	if err != nil {
		utils.ShowErrorWithMessage(err, "interactive client failed to initialize")
		Shutdown(client, InvokerQuery)
		return nil, err
	}

	// start the interactive prompt in a go routine
	go interactiveClient.InteractiveQuery()

	return resultsStreamer, nil
}

// ExecuteQuery :: execute a single query. If shutdownAfterCompletion is true, shutdown the client after completion
func ExecuteQuery(ctx context.Context, queryString string, client *Client) (*results.ResultStreamer, error) {
	logging.LogTime("db.ExecuteQuery start")
	defer logging.LogTime("db.ExecuteQuery end")

	resultsStreamer := results.NewResultStreamer()
	result, err := client.ExecuteQuery(ctx, queryString, false)
	if err != nil {
		return nil, err
	}
	go resultsStreamer.StreamSingleResult(result)
	return resultsStreamer, nil
}
