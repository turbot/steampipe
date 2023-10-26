package steampipe_db_client

import (
	"context"
	"github.com/turbot/pipe-fittings/queryresult"
	"github.com/turbot/pipe-fittings/utils"
)

// TODO KAI MAKE METHOD

// ExecuteQuery executes a single query and returns a result streamer
func ExecuteQuery(ctx context.Context, client *SteampipeDbClient, queryString string, args ...any) (*queryresult.ResultStreamer, error) {
	utils.LogTime("db.ExecuteQuery start")
	defer utils.LogTime("db.ExecuteQuery end")

	resultsStreamer := queryresult.NewResultStreamer()
	result, err := client.Execute(ctx, queryString, args...)
	if err != nil {
		return nil, err
	}
	go func() {
		resultsStreamer.StreamResult(result)
		resultsStreamer.Close()
	}()
	return resultsStreamer, nil
}
