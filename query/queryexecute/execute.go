package queryexecute

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/db/db_common"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/interactive"
	"github.com/turbot/steampipe/utils"
)

func RunInteractiveSession(initChan *chan *db_common.QueryInitData) {
	utils.LogTime("execute.RunInteractiveSession start")
	defer utils.LogTime("execute.RunInteractiveSession end")

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := interactive.RunInteractivePrompt(initChan)
	utils.FailOnError(err)

	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		// signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.AllResultsRead()
	}
}

func RunBatchSession(ctx context.Context, initDataChan chan *db_common.QueryInitData) int {
	// wait for init
	initData := <-initDataChan
	if err := initData.Result.Error; err != nil {
		utils.FailOnError(err)
	}
	// ensure we close client
	defer func() {
		if initData.Client != nil {
			initData.Client.Close()
		}
	}()

	// display any initialisation messages/warnings
	initData.Result.DisplayMessages()

	failures := 0
	if len(initData.Queries) > 0 {
		// if we have resolved any queries, run them
		failures = executeQueries(ctx, initData.Queries, initData.Client)
	}
	// set global exit code
	return failures
}

func executeQueries(ctx context.Context, queries []string, client db_common.Client) int {
	utils.LogTime("queryexecute.executeQueries start")
	defer utils.LogTime("queryexecute.executeQueries end")

	// run all queries
	failures := 0
	for i, q := range queries {
		if err := executeQuery(ctx, q, client); err != nil {
			failures++
			utils.ShowWarning(fmt.Sprintf("executeQueries: query %d of %d failed: %v", i+1, len(queries), err))
		}
		// TODO move into display layer
		if showBlankLineBetweenResults() {
			fmt.Println()
		}
	}

	return failures
}

func executeQuery(ctx context.Context, queryString string, client db_common.Client) error {
	utils.LogTime("query.execute.executeQuery start")
	defer utils.LogTime("query.execute.executeQuery end")

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db_common.ExecuteQuery(ctx, queryString, client)
	if err != nil {
		return err
	}

	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		// signal to the resultStreamer that we are done with this result
		resultsStreamer.AllResultsRead()
	}
	return nil
}

// if we are displaying csv with no header, do not include lines between the query results
func showBlankLineBetweenResults() bool {
	return !(viper.GetString(constants.ArgOutput) == "csv" && !viper.GetBool(constants.ArgHeader))
}
