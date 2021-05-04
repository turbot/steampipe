package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/query/queryresult"
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

func (c *Client) ExecuteQuery(ctx context.Context, query string, countStream bool) (res *queryresult.Result, err error) {
	if query == "" {
		return &queryresult.Result{}, nil
	}
	startTime := time.Now()
	// channel to flag to spinner that the query has run
	queryDone := make(chan bool, 1)
	var spinner *spinner.Spinner

	c.QueryLock.Lock()
	defer func() {
		// if there is no error, readRows() will unlock the QueryLock
		// if there IS an error we need to unlock it here
		if err != nil {
			c.QueryLock.Unlock()
			// stop spinner in case of error
			display.StopSpinner(spinner)
		}
		close(queryDone)
	}()

	if cmdconfig.Viper().GetBool(constants.ConfigKeyShowInteractiveOutput) {
		// if `show-interactive-output` is false, the spinner gets created, but is never shown
		// so the s.Active() will always come back false . . .
		spinner = display.StartSpinnerAfterDelay("Loading results...", constants.SpinnerShowTimeout, queryDone)
	}

	// begin a transaction
	var tx *sql.Tx
	tx, err = c.dbClient.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("error creating transaction: %v", err)
		return
	}

	var rows *sql.Rows
	rows, err = c.dbClient.QueryContext(ctx, query)
	if err != nil {
		// error - rollback transaction
		tx.Rollback()
		return
	}
	// commit transaction
	tx.Commit()

	var colTypes []*sql.ColumnType
	colTypes, err = rows.ColumnTypes()
	if err != nil {
		err = fmt.Errorf("error reading columns from query: %v", err)
		return
	}

	result := queryresult.NewQueryResult(colTypes)

	// read the rows in a go routine
	// NOTE: readRows will unlock QueryLock
	go c.readRows(ctx, startTime, rows, result, spinner)

	return result, nil
}

func (c *Client) readRows(ctx context.Context, start time.Time, rows *sql.Rows, result *queryresult.Result, activeSpinner *spinner.Spinner) {
	// defer this, so that these get cleaned up even if there is an unforeseen error
	defer func() {
		// close the sql rows object
		rows.Close()
		if err := rows.Err(); err != nil {
			result.StreamError(err)
		}
		// close the channels in the result object
		result.Close()
		// the Unlock will have been locked by the calling function, ExecuteQuery
		c.QueryLock.Unlock()
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
			display.UpdateSpinnerMessage(activeSpinner, "Cancelling query")
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
			display.UpdateSpinnerMessage(activeSpinner, fmt.Sprintf("Loading results: %3s", humanizeRowCount(rowCount)))
			rowCount++
		}
		if !continueToNext {
			break
		}
	}
	// we are done fetching results. time for display. remove the spinner
	display.StopSpinner(activeSpinner)

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
