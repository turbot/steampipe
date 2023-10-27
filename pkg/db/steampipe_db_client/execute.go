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

	// TODO KAI REMOVED
	//sessionResult := c.AcquireSession(ctx)
	//if sessionResult.Error != nil {
	//	return nil, sessionResult.Error
	//}

	// TODO steampipe only
	//// disable statushooks when timing is enabled, because setShouldShowTiming internally calls the readRows funcs which
	//// calls the statushooks.Done, which hides the `Executing query…` spinner, when timing is enabled.
	//timingCtx := statushooks.DisableStatusHooks(ctx)
	//// re-read ArgTiming from viper (in case the .timing command has been run)
	//// (this will refetch ScanMetadataMaxId if timing has just been enabled)
	//c.setShouldShowTiming(timingCtx, sessionResult.Session)

	// define callback to close session when the async execution is complete
	// TODO KAI session close  waited for pg shutdown

	//closeSessionCallback := func() { databaseConnection.Close() }
	//return c.executeOnConnection(ctx, databaseConnection, closeSessionCallback, query, args...)

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
