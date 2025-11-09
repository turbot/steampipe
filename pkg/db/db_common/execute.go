package db_common

import (
	"context"

	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/query/queryresult"
)

// ExecuteQuery executes a single query. If shutdownAfterCompletion is true, shutdown the client after completion
func ExecuteQuery(ctx context.Context, client Client, queryString string, args ...any) (*queryresult.ResultStreamer, error) {
	utils.LogTime("db.ExecuteQuery start")
	defer utils.LogTime("db.ExecuteQuery end")

	resultsStreamer := queryresult.NewResultStreamer()
	result, err := client.Execute(ctx, queryString, args...)
	if err != nil {
		return nil, err
	}
	go func() {
		// Pass the embedded pipe-fittings Result to the streamer
		resultsStreamer.StreamResult(result.Result)
		resultsStreamer.Close()
	}()
	return resultsStreamer, nil
}
