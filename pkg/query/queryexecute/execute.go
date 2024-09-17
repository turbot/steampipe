package queryexecute

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/contexthelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/querydisplay"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/connection_sync"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/query"
)

func RunInteractiveSession(ctx context.Context, initData *query.InitData) error {
	utils.LogTime("execute.RunInteractiveSession start")
	defer utils.LogTime("execute.RunInteractiveSession end")

	// the db executor sends result data over resultsStreamer
	result := interactive.RunInteractivePrompt(ctx, initData)

	// print the data as it comes
	for r := range result.Streamer.Results {
		rowCount, _ := querydisplay.ShowOutput(ctx, r)
		// show timing
		display.DisplayTiming(r, rowCount)
		// signal to the resultStreamer that we are done with this chunk of the stream
		result.Streamer.AllResultsRead()
	}
	return result.PromptErr
}

func RunBatchSession(ctx context.Context, initData *query.InitData) (int, error) {
	// start cancel handler to intercept interrupts and cancel the context
	// NOTE: use the initData Cancel function to ensure any initialisation is cancelled if needed
	contexthelpers.StartCancelHandler(initData.Cancel)

	// wait for init
	<-initData.Loaded

	if err := initData.Result.Error; err != nil {
		return 0, err
	}

	// display any initialisation messages/warnings
	initData.Result.DisplayMessages()

	// if there is a custom search path, wait until the first connection of each plugin has loaded
	if customSearchPath := initData.Client.GetCustomSearchPath(); customSearchPath != nil {
		if err := connection_sync.WaitForSearchPathSchemas(ctx, initData.Client, customSearchPath); err != nil {
			return 0, err
		}
	}

	failures := 0
	if len(initData.Queries) > 0 {
		// if we have resolved any queries, run them
		failures = executeQueries(ctx, initData)
	}
	// return the number of query failures and the number of rows that returned errors
	return failures, nil
}

func executeQueries(ctx context.Context, initData *query.InitData) int {
	utils.LogTime("queryexecute.executeQueries start")
	defer utils.LogTime("queryexecute.executeQueries end")

	// failures return the number of queries that failed and also the number of rows that
	// returned errors
	failures := 0
	t := time.Now()

	var err error

	for i, q := range initData.Queries {
		// if executeQuery fails it returns err, else it returns the number of rows that returned errors while execution
		if err, failures = executeQuery(ctx, initData.Client, q); err != nil {
			failures++
			error_helpers.ShowWarning(fmt.Sprintf("query %d of %d failed: %v", i+1, len(initData.Queries), error_helpers.DecodePgError(err)))
			// if timing flag is enabled, show the time taken for the query to fail
			if cmdconfig.Viper().GetString(pconstants.ArgTiming) != pconstants.ArgOff {
				querydisplay.DisplayErrorTiming(t)
			}
		}
		// TODO move into display layer
		// Only show the blank line between queries, not after the last one
		if (i < len(initData.Queries)-1) && showBlankLineBetweenResults() {
			fmt.Println()
		}
	}

	return failures
}

func executeQuery(ctx context.Context, client db_common.Client, resolvedQuery *modconfig.ResolvedQuery) (error, int) {
	utils.LogTime("query.execute.executeQuery start")
	defer utils.LogTime("query.execute.executeQuery end")

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db_common.ExecuteQuery(ctx, client, resolvedQuery.ExecuteSQL, resolvedQuery.Args...)
	if err != nil {
		return err, 0
	}

	rowErrors := 0 // get the number of rows that returned an error
	// print the data as it comes
	for r := range resultsStreamer.Results {
		rowCount, _ := querydisplay.ShowOutput(ctx, r)
		// show timing
		display.DisplayTiming(r, rowCount)

		// signal to the resultStreamer that we are done with this result
		resultsStreamer.AllResultsRead()
	}
	return nil, rowErrors
}

// if we are displaying csv with no header, do not include lines between the query results
func showBlankLineBetweenResults() bool {
	return !(viper.GetString(pconstants.ArgOutput) == "csv" && !viper.GetBool(pconstants.ArgHeader))
}
