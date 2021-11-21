package db_client

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/utils"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ExecuteSync implements Client
// execute a query against this client and wait for the result
func (c *DbClient) ExecuteSync(ctx context.Context, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error) {
	// acquire a session
	sessionResult := c.AcquireSession(ctx)
	if sessionResult.Error != nil {
		return nil, sessionResult.Error
	}
	defer sessionResult.Session.Close()
	return c.ExecuteSyncInSession(ctx, sessionResult.Session, query, disableSpinner)
}

// ExecuteSyncInSession implements Client
// execute a query against this client and wait for the result
func (c *DbClient) ExecuteSyncInSession(ctx context.Context, session *db_common.DatabaseSession, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error) {
	if query == "" {
		return &queryresult.SyncQueryResult{}, nil
	}

	result, err := c.ExecuteInSession(ctx, session, query, nil, disableSpinner)
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

// Execute  implements Client
// execute the provided query against the Database in the given context.Context
// Bear in mind that whenever ExecuteQuery is called, the returned `queryresult.Result` MUST be fully read -
// otherwise the transaction is left open, which will block the connection and will prevent subsequent communications
// with the service
func (c *DbClient) Execute(ctx context.Context, query string, disableSpinner bool) (*queryresult.Result, error) {
	// acquire a session
	sessionResult := c.AcquireSession(ctx)
	if sessionResult.Error != nil {
		return nil, sessionResult.Error
	}

	// define callback to close session when the async execution is complete
	closeSessionCallback := func() { sessionResult.Session.Close() }
	return c.ExecuteInSession(ctx, sessionResult.Session, query, closeSessionCallback, disableSpinner)
}

func (c *DbClient) ExecuteInSession(ctx context.Context, session *db_common.DatabaseSession, query string, onComplete func(), disableSpinner bool) (res *queryresult.Result, err error) {
	if query == "" {
		return queryresult.NewQueryResult(nil), nil
	}

	startTime := time.Now()
	// channel to flag to spinner that the query has run
	var spinner *spinner.Spinner
	var tx *sql.Tx

	defer func() {
		if err != nil {
			// stop spinner in case of error
			display.StopSpinner(spinner)
			// error - rollback transaction if we have one
			if tx != nil {
				tx.Rollback()
			}
			// call the completion callback - if one was provided
			if onComplete != nil {
				onComplete()
			}
		}
	}()

	if !disableSpinner && cmdconfig.Viper().GetBool(constants.ConfigKeyShowInteractiveOutput) {
		// if `show-interactive-output` is false, the spinner gets created, but is never shown
		// so the s.Active() will always come back false . . .
		spinner = display.ShowSpinner("Loading results...")
	}

	// begin a transaction
	tx, err = c.createTransaction(ctx, session.Connection, true)
	if err != nil {
		return
	}
	// start query
	var rows *sql.Rows
	rows, err = c.startQuery(ctx, query, tx)
	if err != nil {
		return
	}

	var colTypes []*sql.ColumnType
	colTypes, err = rows.ColumnTypes()
	if err != nil {
		err = fmt.Errorf("error reading columns from query: %v", err)
		return
	}

	result := queryresult.NewQueryResult(colTypes)

	// read the rows in a go routine
	go func() {
		// read in the rows and stream to the query result object
		c.readRows(ctx, startTime, rows, result, spinner)
		// commit transaction
		if ctx.Err() == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
		if onComplete != nil {
			onComplete()
		}
	}()

	return result, nil
}

// run query in a goroutine, so we can check for cancellation
// in case the client becomes unresponsive and does not respect context cancellation
func (c *DbClient) startQuery(ctx context.Context, query string, tx *sql.Tx) (rows *sql.Rows, err error) {
	doneChan := make(chan bool)
	defer func() {
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				// if the context deadline has been exceeded, call refreshDbClient to create a new SQL client
				// this will refresh the session data which will have been cleared by the SQL client error handling
				c.refreshDbClient(context.Background())
			}
		}
	}()
	go func() {
		// start asynchronous query
		rows, err = tx.QueryContext(ctx, query)
		close(doneChan)
	}()

	select {
	case <-doneChan:
	case <-ctx.Done():
		err = ctx.Err()
	}
	return
}

func (c *DbClient) readRows(ctx context.Context, start time.Time, rows *sql.Rows, result *queryresult.Result, activeSpinner *spinner.Spinner) {
	// defer this, so that these get cleaned up even if there is an unforeseen error
	defer func() {
		// we are done fetching results. time for display. remove the spinner
		display.StopSpinner(activeSpinner)
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
		// we do not need to stream because
		// defer takes care of it!
		return
	}
	cols, err := rows.Columns()
	if err != nil {
		// we do not need to stream because
		// defer takes care of it!
		return
	}

	for rows.Next() {
		continueToNext := true
		select {
		case <-ctx.Done():
			display.UpdateSpinnerMessage(activeSpinner, "Cancelling query")
			continueToNext = false
		default:
			if rowResult, err := readRowContext(ctx, rows, cols, colTypes); err != nil {
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

	// set the time that it took for this one to execute
	result.Duration <- time.Since(start)
}

func readRowContext(ctx context.Context, rows *sql.Rows, cols []string, colTypes []*sql.ColumnType) ([]interface{}, error) {
	c := make(chan bool, 1)
	var readRowResult []interface{}
	var readRowError error
	go func() {
		readRowResult, readRowError = readRow(rows, cols, colTypes)
		close(c)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c:
		return readRowResult, readRowError
	}

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
		// return error, handling cancellation error explicitly
		return nil, utils.HandleCancelError(err)
	}
	return populateRow(columnValues, colTypes), nil
}

func populateRow(columnValues []interface{}, colTypes []*sql.ColumnType) []interface{} {
	result := make([]interface{}, len(columnValues))
	for i, columnValue := range columnValues {
		if columnValue != nil {
			colType := colTypes[i]
			dbType := colType.DatabaseTypeName()
			switch dbType {
			case "JSON", "JSONB":
				var val interface{}
				if err := json.Unmarshal(columnValue.([]byte), &val); err != nil {
					// what???
					// TODO how to handle error
				}
				result[i] = val
			default:
				result[i] = columnValue
			}
		}
	}
	return result
}

func humanizeRowCount(count int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", count)
}
