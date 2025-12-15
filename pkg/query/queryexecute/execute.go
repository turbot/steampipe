package queryexecute

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/constants"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/contexthelpers"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/pipes"
	"github.com/turbot/pipe-fittings/v2/querydisplay"
	pqueryresult "github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/pipe-fittings/v2/steampipeconfig"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/connection_sync"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/display"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/interactive"
	"github.com/turbot/steampipe/v2/pkg/query"
	"github.com/turbot/steampipe/v2/pkg/query/queryresult"
	"github.com/turbot/steampipe/v2/pkg/snapshot"
)

func RunInteractiveSession(ctx context.Context, initData *query.InitData) error {
	utils.LogTime("execute.RunInteractiveSession start")
	defer utils.LogTime("execute.RunInteractiveSession end")

	// the db executor sends result data over resultsStreamer
	result := interactive.RunInteractivePrompt(ctx, initData)

	// print the data as it comes
	for r := range result.Streamer.Results {
		// wrap the result from pipe-fittings with our wrapper that has idempotent Close
		wrapped := queryresult.WrapResult(r)
		rowCount, _ := querydisplay.ShowOutput(ctx, r)
		// show timing
		display.DisplayTiming(wrapped, rowCount)
		// signal to the resultStreamer that we are done with this chunk of the stream
		result.Streamer.AllResultsRead()
	}
	return result.PromptErr
}

func RunBatchSession(ctx context.Context, initData *query.InitData) (int, error) {
	if initData == nil {
		return 0, fmt.Errorf("initData cannot be nil")
	}

	// start cancel handler to intercept interrupts and cancel the context
	// NOTE: use the initData Cancel function to ensure any initialisation is cancelled if needed
	contexthelpers.StartCancelHandler(initData.Cancel)

	// wait for init, respecting context cancellation
	select {
	case <-initData.Loaded:
		// initialization complete, continue
	case <-ctx.Done():
		// context cancelled before initialization completed
		return 0, ctx.Err()
	}

	if err := initData.Result.Error; err != nil {
		return 0, err
	}

	// display any initialisation messages/warnings
	initData.Result.DisplayMessages()

	// validate that Client is not nil
	if initData.Client == nil {
		return 0, fmt.Errorf("client is required but not initialized")
	}

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

	// Check if Client is nil - this can happen if initialization failed
	if initData.Client == nil {
		error_helpers.ShowWarning("cannot execute queries: database client is not initialized")
		return len(initData.Queries)
	}

	// failures return the number of queries that failed and also the number of rows that
	// returned errors
	failures := 0
	t := time.Now()

	var err error

	for i, q := range initData.Queries {
		// if executeQuery fails it returns err, else it returns the number of rows that returned errors while execution
		if err, failures = executeQuery(ctx, initData, q); err != nil {
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

func executeQuery(ctx context.Context, initData *query.InitData, resolvedQuery *modconfig.ResolvedQuery) (error, int) {
	utils.LogTime("query.execute.executeQuery start")
	defer utils.LogTime("query.execute.executeQuery end")

	var snap *steampipeconfig.SteampipeSnapshot

	// the db executor sends result data over resultsStreamer
	resultsStreamer, err := db_common.ExecuteQuery(ctx, initData.Client, resolvedQuery.ExecuteSQL, resolvedQuery.Args...)
	if err != nil {
		return err, 0
	}

	rowErrors := 0 // get the number of rows that returned an error
	// print the data as it comes
	for r := range resultsStreamer.Results {
		// wrap the result from pipe-fittings with our wrapper that has idempotent Close
		wrapped := queryresult.WrapResult(r)

		// if the output format is snapshot or export is set or share/snapshot args are set, we need to generate a snapshot
		if needSnapshot() {
			snap, err = snapshot.QueryResultToSnapshot(ctx, r, resolvedQuery, initData.Client.GetRequiredSessionSearchPath(), initData.StartTime)
			if err != nil {
				return err, 0
			}

			// re-generate the query result from the snapshot. since the row stream in the actual queryresult has been exhausted(while generating the snapshot),
			// we need to re-generate it for other output formats
			newQueryResult, err := snapshot.SnapshotToQueryResult[pqueryresult.TimingContainer](snap, initData.StartTime)
			if err != nil {
				return err, 0
			}

			// if the output format is snapshot we don't call the querydisplay code in pipe-fittings, instead we
			// generate the snapshot and display it to stdout
			outputFormat := viper.GetString(pconstants.ArgOutput)
			if outputFormat == pconstants.OutputFormatSnapshot || outputFormat == pconstants.OutputFormatSteampipeSnapshotShort {

				// display the snapshot as JSON
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.SetEscapeHTML(false)
				if err := encoder.Encode(snap); err != nil {
					//nolint:forbidigo // acceptable
					fmt.Print("Error displaying result as snapshot", err)
					return err, 0
				}
			}

			// if we need to export the snapshot, we export it directly from here
			if viper.IsSet(pconstants.ArgExport) {
				exportArgs := viper.GetStringSlice(pconstants.ArgExport)
				exportMsg, err := initData.ExportManager.DoExport(ctx, "query", snap, exportArgs)
				if err != nil {
					return err, 0
				}
				// print the location where the file is exported
				if len(exportMsg) > 0 && viper.GetBool(pconstants.ArgProgress) {
					fmt.Printf("\n")                           //nolint:forbidigo // intentional use of fmt
					fmt.Println(strings.Join(exportMsg, "\n")) //nolint:forbidigo // intentional use of fmt
					fmt.Printf("\n")                           //nolint:forbidigo // intentional use of fmt
				}
			}

			// if we need to publish the snapshot, we publish it directly from here
			if err := publishSnapshotIfNeeded(ctx, snap); err != nil {
				return err, 0
			}

			// if other output formats are also needed, we call the querydisplay using the re-generated query result
			rowCount, _ := querydisplay.ShowOutput(ctx, newQueryResult)
			// show timing
			display.DisplayTiming(wrapped, rowCount)

			// signal to the resultStreamer that we are done with this result
			resultsStreamer.AllResultsRead()
			return nil, rowErrors
		}

		// for other output formats, we call the querydisplay code in pipe-fittings
		rowCount, rowErrs := querydisplay.ShowOutput(ctx, r)
		// show timing
		display.DisplayTiming(wrapped, rowCount)

		// signal to the resultStreamer that we are done with this result
		resultsStreamer.AllResultsRead()
		rowErrors = rowErrs
	}
	return nil, rowErrors
}

func needSnapshot() bool {
	// Get the output format from the configuration
	outputFormat := viper.GetString(pconstants.ArgOutput)
	shouldShare := viper.GetBool(pconstants.ArgShare)
	shouldUpload := viper.GetBool(pconstants.ArgSnapshot)

	// Check if the output format is a snapshot format or if ArgExport is set
	if outputFormat == pconstants.OutputFormatSnapshot || outputFormat == pconstants.OutputFormatSteampipeSnapshotShort || viper.IsSet(pconstants.ArgExport) || shouldShare || shouldUpload {
		return true
	}

	// If none of the conditions are met, return false
	return false
}

func publishSnapshotIfNeeded(ctx context.Context, snapshot *steampipeconfig.SteampipeSnapshot) error {
	shouldShare := viper.GetBool(pconstants.ArgShare)
	shouldUpload := viper.GetBool(pconstants.ArgSnapshot)

	if !(shouldShare || shouldUpload) {
		return nil
	}

	message, err := pipes.PublishSnapshot(ctx, snapshot, shouldShare)
	if err != nil {
		// reword "402 Payment Required" error
		return handlePublishSnapshotError(err)
	}
	if viper.GetBool(constants.ArgProgress) {
		fmt.Println(message)
	}
	return nil
}

func handlePublishSnapshotError(err error) error {
	if err.Error() == "402 Payment Required" {
		return fmt.Errorf("maximum number of snapshots reached")
	}
	return err
}

// if we are displaying csv with no header, do not include lines between the query results
func showBlankLineBetweenResults() bool {
	return !(viper.GetString(pconstants.ArgOutput) == "csv" && !viper.GetBool(pconstants.ArgHeader))
}
