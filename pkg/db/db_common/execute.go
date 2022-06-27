package db_common

import (
	"context"

	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/utils"
)

// ExecuteQuery executes a single query. If shutdownAfterCompletion is true, shutdown the client after completion
func ExecuteQuery(ctx context.Context, queryString string, client Client) (*queryresult.ResultStreamer, error) {
	utils.LogTime("db.ExecuteQuery start")
	defer utils.LogTime("db.ExecuteQuery end")

	resultsStreamer := queryresult.NewResultStreamer()
	result, err := client.Execute(ctx, queryString)
	if err != nil {
		return nil, err
	}
	go func() {
		resultsStreamer.StreamResult(result)
		resultsStreamer.Close()
	}()
	return resultsStreamer, nil
}
