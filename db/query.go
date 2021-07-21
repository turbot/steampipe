package db

import (
	"context"
	"errors"
	"log"

	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/utils"
)

// EnsureDbAndStartService :: ensure db is installed and start service if necessary
func EnsureDbAndStartService(invoker Invoker, refreshConnections bool) error {
	utils.LogTime("db.EnsureDbAndStartService start")
	defer utils.LogTime("db.EnsureDbAndStartService end")

	log.Println("[TRACE] db.EnsureDbAndStartService start")

	EnsureDBInstalled()
	status, err := GetStatus()
	if err != nil {
		return errors.New("could not retrieve service status")
	}

	if status == nil {
		// the db service is not started - start it
		return StartImplicitService(invoker, refreshConnections)
	}
	return nil
}

// RunInteractivePrompt :: start the interactive query prompt
func RunInteractivePrompt(initChan *chan *QueryInitData) (*queryresult.ResultStreamer, error) {
	resultsStreamer := queryresult.NewResultStreamer()

	interactiveClient, err := newInteractiveClient(initChan, resultsStreamer)
	if err != nil {
		utils.ShowErrorWithMessage(err, "interactive client failed to initialize")
		Shutdown(nil, InvokerQuery)
		return nil, err
	}

	// start the interactive prompt in a go routine
	go interactiveClient.InteractiveQuery()

	return resultsStreamer, nil
}

// ExecuteQuery :: execute a single query. If shutdownAfterCompletion is true, shutdown the client after completion
func ExecuteQuery(ctx context.Context, queryString string, client *Client) (*queryresult.ResultStreamer, error) {
	utils.LogTime("db.ExecuteQuery start")
	defer utils.LogTime("db.ExecuteQuery end")

	resultsStreamer := queryresult.NewResultStreamer()
	result, err := client.Execute(ctx, queryString, false)
	if err != nil {
		return nil, err
	}
	go resultsStreamer.StreamSingleResult(result)
	return resultsStreamer, nil
}
