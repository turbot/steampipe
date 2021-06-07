package execute

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

func RunInteractiveSession(workspace *workspace.Workspace, client *db.Client) {
	// start the workspace file watcher
	if viper.GetBool(constants.ArgWatch) {
		err := workspace.SetupWatcher(client)
		utils.FailOnError(err)
	}

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db.RunInteractivePrompt(workspace, client)
	utils.FailOnError(err)

	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		// signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.Done()
	}
}

func ExecuteQueries(ctx context.Context, queries []string, client *db.Client) int {
	utils.LogTime("query.execute.ExecuteQueries start")
	defer utils.LogTime("query.execute.ExecuteQueries end")
	// run all queries
	failures := 0
	for i, q := range queries {
		if err := executeQuery(ctx, q, client); err != nil {
			failures++
			utils.ShowWarning(fmt.Sprintf("executeQueries: query %d of %d failed: %v", i+1, len(queries), err))
		}
		if showBlankLineBetweenResults() {
			fmt.Println()
		}
	}

	return failures
}

func executeQuery(ctx context.Context, queryString string, client *db.Client) error {
	utils.LogTime("query.execute.executeQuery start")
	defer utils.LogTime("query.execute.executeQuery end")
	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db.ExecuteQuery(ctx, queryString, client)
	if err != nil {
		return err
	}

	// TODO encapsulate this in display object
	// print the data as it comes
	for r := range resultsStreamer.Results {
		display.ShowOutput(r)
		// signal to the resultStreamer that we are done with this chunk of the stream
		resultsStreamer.Done()
	}
	return nil
}

// if we are displaying csv with no header, do not include lines between the query results
func showBlankLineBetweenResults() bool {
	return !(viper.GetString(constants.ArgOutput) == "csv" && !viper.GetBool(constants.ArgHeader))
}
