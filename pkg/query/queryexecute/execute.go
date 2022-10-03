package queryexecute

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/query"
	"github.com/turbot/steampipe/pkg/utils"
)

func RunInteractiveSession(ctx context.Context, initData *query.InitData) {
	utils.LogTime("execute.RunInteractiveSession start")
	defer utils.LogTime("execute.RunInteractiveSession end")

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := interactive.RunInteractivePrompt(ctx, initData)
	utils.FailOnError(err)

	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(ctx, r)
		// signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.AllResultsRead()
	}
}

func RunBatchSession(ctx context.Context, initData *query.InitData) int {
	// ensure we close client
	defer initData.Cleanup(ctx)

	// start cancel handler to intercept interrupts and cancel the context
	// NOTE: use the initData Cancel function to ensure any initialisation is cancelled if needed
	contexthelpers.StartCancelHandler(initData.Cancel)

	// wait for init
	<-initData.Loaded
	if err := initData.Result.Error; err != nil {
		utils.FailOnError(err)
	}

	// display any initialisation messages/warnings
	initData.Result.DisplayMessages()

	failures := 0
	if len(initData.Queries) > 0 {
		// if we have resolved any queries, run them
		failures = executeQueries(ctx, initData)
	}
	// set global exit code
	return failures
}

func executeQueries(ctx context.Context, initData *query.InitData) int {
	utils.LogTime("queryexecute.executeQueries start")
	defer utils.LogTime("queryexecute.executeQueries end")

	// run all queries
	failures := 0
	t := time.Now()
	// build ordered list of queries
	// (ordered for testing repeatability)
	var queryNames []string = utils.SortedMapKeys(initData.Queries)

	for i, name := range queryNames {
		q := initData.Queries[name]
		if err := executeQuery(ctx, q, initData.Client); err != nil {
			failures++
			utils.ShowWarning(fmt.Sprintf("executeQueries: query %d of %d failed: %v", i+1, len(queryNames), err))
			// if timing flag is enabled, show the time taken for the query to fail
			if cmdconfig.Viper().GetBool(constants.ArgTiming) {
				display.DisplayErrorTiming(t)
			}
		}
		// TODO move into display layer
		// Only show the blank line between queries, not after the last one
		if (i < len(queryNames)-1) && showBlankLineBetweenResults() {
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
		display.ShowOutput(ctx, r)
		// signal to the resultStreamer that we are done with this result
		resultsStreamer.AllResultsRead()
	}
	return nil
}

// if we are displaying csv with no header, do not include lines between the query results
func showBlankLineBetweenResults() bool {
	return !(viper.GetString(constants.ArgOutput) == "csv" && !viper.GetBool(constants.ArgHeader))
}
