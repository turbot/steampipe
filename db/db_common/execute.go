package db_common

import (
	"context"

	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/utils"
)

// ExecuteQuery :: execute a single query. If shutdownAfterCompletion is true, shutdown the client after completion
func ExecuteQuery(ctx context.Context, queryString string, client Client) (*queryresult.ResultStreamer, error) {
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
