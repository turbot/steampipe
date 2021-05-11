package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/utils"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ExecuteSync :: execute a query against this client and wait for the result
func (c *Client) ExecuteSync(ctx context.Context, query string) (*queryresult.SyncQueryResult, error) {
	result, err := c.ExecuteQuery(ctx, query, false)
	if err != nil {
		return nil, err
	}
	syncResult := &queryresult.SyncQueryResult{ColTypes: result.ColTypes}
	for row := range *result.RowChan {
		select {
		case <-ctx.Done():
		default:
			syncResult.Rows = append(syncResult.Rows, row)
		}
	}
	syncResult.Duration = <-result.Duration
	return syncResult, nil
}

func (c *Client) ExecuteQuery(ctx context.Context, query string, countStream bool) (*queryresult.Result, error) {
	if query == "" {
		return &queryresult.Result{}, nil
	}

	startTime := time.Now()

	// channel to flag to spinner that the query has run
	queryDone := make(chan bool, 1)

	// start spinner after a short delay
	var spinner *spinner.Spinner

	if cmdconfig.Viper().GetBool(constants.ConfigKeyShowInteractiveOutput) {
		// if showspinner is false, the spinner gets created, but is never shown
		// so the s.Active() will always come back false . . .
		spinner = utils.StartSpinnerAfterDelay("Loading results...", constants.SpinnerShowTimeout, queryDone)
	}

	rows, err := c.dbClient.QueryContext(ctx, query)
	queryDone <- true

	if err != nil {
		// in case the query takes a long time to fail
		utils.StopSpinner(spinner)
		return nil, err
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("error reading columns from query: %v", err)
	}

	result := queryresult.NewQueryResult(colTypes)

	// read the rows in a go routine
	go readRows(ctx, startTime, rows, result, spinner)

	return result, nil
}

func readRows(ctx context.Context, start time.Time, rows *sql.Rows, result *queryresult.Result, activeSpinner *spinner.Spinner) {
	// defer this, so that these get cleaned up even if there is an unforeseen error
	defer func() {
		// close the sql rows object
		rows.Close()
		if err := rows.Err(); err != nil {
			result.StreamError(err)
		}
		// close the channels in the result object
		result.Close()
	}()

	rowCount := 0
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		result.StreamError(err)
		return
	}
	cols, err := rows.Columns()
	if err != nil {
		result.StreamError(err)
		return
	}

	for rows.Next() {
		continueToNext := true
		select {
		case <-ctx.Done():
			utils.UpdateSpinnerMessage(activeSpinner, "Cancelling query")
			continueToNext = false
		default:
			if rowResult, err := readRow(rows, cols, colTypes); err != nil {
				result.StreamError(err)
				continueToNext = false
			} else {
				result.StreamRow(rowResult)
			}
			// update the spinner message with the count of rows that have already been fetched
			// this will not show if the spinner is not active
			utils.UpdateSpinnerMessage(activeSpinner, fmt.Sprintf("Loading results: %3s", humanizeRowCount(rowCount)))
			rowCount++
		}
		if !continueToNext {
			break
		}
	}
	// we are done fetching results. time for display. remove the spinner
	utils.StopSpinner(activeSpinner)

	// set the time that it took for this one to execute
	result.Duration <- time.Since(start)
}

func readRow(rows *sql.Rows, cols []string, colTypes []*sql.ColumnType) ([]interface{}, error) {
	// slice of interfaces to receive the row data
	columnValues := make([]interface{}, len(cols))
	// make a slice of pointers to the result to pass to scan
	resultPtrs := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i := range columnValues {
		resultPtrs[i] = &columnValues[i]
	}
	err := rows.Scan(resultPtrs...)
	if err != nil {
		if err == context.Canceled {
			err = fmt.Errorf("Cancelled")
		}
		return nil, err
	}
	return populateRow(columnValues, colTypes), nil
}

func humanizeRowCount(count int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", count)
}
