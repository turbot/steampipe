package db

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/query/queryresult"
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

	alreadyStartedByCheckOrQueryOrPlugin := status != nil && (status.Invoker == InvokerQuery || status.Invoker == InvokerCheck || status.Invoker == InvokerPlugin)
	tryingStartByCheckOrQueryOrPlugin := (invoker == InvokerQuery || invoker == InvokerCheck || invoker == InvokerPlugin)

	if alreadyStartedByCheckOrQueryOrPlugin && tryingStartByCheckOrQueryOrPlugin {
		return fmt.Errorf("You already have a %s session open. To run multiple sessions, first run %s.\nTo kill existing sessions run %s", constants.Bold(fmt.Sprintf("steampipe %s", status.Invoker)), constants.Bold("steampipe service start"), constants.Bold("steampipe service stop --force"))
	}

	if status == nil {
		// the db service is not started - start it
		StartService(invoker)
	}
	return nil
}

// RunInteractivePrompt :: start the interactive query prompt
func RunInteractivePrompt(workspace NamedQueryProvider, client *Client) (*queryresult.ResultStreamer, error) {
	resultsStreamer := queryresult.NewResultStreamer()

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
func ExecuteQuery(ctx context.Context, queryString string, client *Client) (*queryresult.ResultStreamer, error) {
	logging.LogTime("db.ExecuteQuery start")
	defer logging.LogTime("db.ExecuteQuery end")

	resultsStreamer := queryresult.NewResultStreamer()
	result, err := client.ExecuteQuery(ctx, queryString, false)
	if err != nil {
		return nil, err
	}
	go resultsStreamer.StreamSingleResult(result)
	return resultsStreamer, nil
}
