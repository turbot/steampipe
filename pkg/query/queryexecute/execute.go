package queryexecute

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
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
		display.ShowOutput(ctx, r, nil)
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
	queries := initData.Queries
	utils.LogTime("queryexecute.executeQueries start")
	defer utils.LogTime("queryexecute.executeQueries end")

	// run all queries
	failures := 0
	t := time.Now()
	idx := 0
	for name, q := range queries {
		// try to resolve the source query provider - this is used for snapshot creation
		queryProvider := resolveQueryProvider(name, initData, q)

		if err := executeQuery(ctx, q, queryProvider, initData.Client); err != nil {
			failures++
			utils.ShowWarning(fmt.Sprintf("executeQueries: query %d of %d failed: %v", idx+1, len(queries), err))
			// if timing flag is enabled, show the time taken for the query to fail
			if cmdconfig.Viper().GetBool(constants.ArgTiming) {
				display.DisplayErrorTiming(t)
			}
		}
		// TODO move into display layer
		// Only show the blank line between queries, not after the last one
		if (idx < len(queries)-1) && showBlankLineBetweenResults() {
			fmt.Println()
		}
	}

	return failures
}

func resolveQueryProvider(name string, initData *query.InitData, q string) modconfig.HclResource {
	var queryProvider modconfig.HclResource
	if parsedName, err := modconfig.ParseResourceName(name); err == nil {
		queryProvider, _ = modconfig.GetResource(initData.Workspace, parsedName)
	}
	if queryProvider == nil {
		queryProvider = &modconfig.Query{
			ShortName: "local_query",
			Title:     utils.ToStringPointer("Local Query"),
			SQL:       utils.ToStringPointer(q),
		}
	}
	return queryProvider
}

func executeQuery(ctx context.Context, queryString string, queryProvider modconfig.HclResource, client db_common.Client) error {
	utils.LogTime("query.execute.executeQuery start")
	defer utils.LogTime("query.execute.executeQuery end")

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db_common.ExecuteQuery(ctx, queryString, client)
	if err != nil {
		return err
	}

	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(ctx, r, queryProvider)

		// signal to the resultStreamer that we are done with this result
		resultsStreamer.AllResultsRead()
	}
	return nil
}

// if we are displaying csv with no header, do not include lines between the query results
func showBlankLineBetweenResults() bool {
	return !(viper.GetString(constants.ArgOutput) == "csv" && !viper.GetBool(constants.ArgHeader))
}
