package db_client

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/sethvargo/go-retry"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
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
	session, err := c.AcquireSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return c.ExecuteSyncInSession(ctx, session, query, disableSpinner)
}

// ExecuteSyncInSession implements Client
// execute a query against this client and wait for the result
func (c *DbClient) ExecuteSyncInSession(ctx context.Context, session *sql.Conn, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error) {
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
func (c *DbClient) Execute(ctx context.Context, query string, disableSpinner bool) (res *queryresult.Result, err error) {
	// acquire a session
	session, err := c.AcquireSession(ctx)
	if err != nil {
		return
	}

	// define callback to close session when the async execution is complete
	closeSessionCallback := func() { session.Close() }
	return c.ExecuteInSession(ctx, session, query, closeSessionCallback, disableSpinner)
}

func (c *DbClient) ExecuteInSession(ctx context.Context, session *sql.Conn, query string, onComplete func(), disableSpinner bool) (res *queryresult.Result, err error) {
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
			// TODO change this to NOT close the session here - instead maybe call a passed-in OnComplete callback
			// close and release the db connection, if we have one
			if session != nil {
				session.Close()
			}
		}
	}()

	if !disableSpinner && cmdconfig.Viper().GetBool(constants.ConfigKeyShowInteractiveOutput) {
		// if `show-interactive-output` is false, the spinner gets created, but is never shown
		// so the s.Active() will always come back false . . .
		spinner = display.ShowSpinner("Loading results...")
	}

	// begin a transaction
	tx, err = c.createTransaction(ctx, session, true)
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
		tx.Commit()
		if onComplete != nil {
			onComplete()
		}
	}()

	return result, nil
}

func (c *DbClient) AcquireSession(ctx context.Context) (*sql.Conn, error) {
	c.sessionAcquireLock.Lock()
	defer c.sessionAcquireLock.Unlock()

	// get a database conneciton and query its backend pid
	// note - this will retry if the connection is bad
	session, backendPid, err := c.getSessionWithRetries(ctx)
	if err != nil {
		return nil, err
	}

	if c.ensureSessionFunc == nil {
		return session, nil
	}

	sessionStat, isInitialized := c.initializedSessions[backendPid]
	if !isInitialized {
		sessionStat = NewSessionStat()
		if err := c.ensureSessionFunc(ctx, session); err != nil {
			return nil, err
		}
		sessionStat.Initialized = time.Now()
	}

	// update required session search path if needed
	if strings.Join(sessionStat.SearchPath, ",") != strings.Join(c.requiredSessionSearchPath, ",") {
		if err := c.setSessionSearchPathToRequired(ctx, session); err != nil {
			return nil, err
		}
		sessionStat.SearchPath = c.requiredSessionSearchPath
	}

	sessionStat.LastUsed = time.Now()
	sessionStat.UsedCount++
	// now write back to the map
	c.initializedSessions[backendPid] = sessionStat

	return session, nil
}

func (c *DbClient) getSessionWithRetries(ctx context.Context) (*sql.Conn, int64, error) {
	backoff, err := retry.NewFibonacci(100 * time.Millisecond)
	if err != nil {
		return nil, 0, err
	}

	retries := 0
	var session *sql.Conn
	var backendPid int64
	const getSessionMaxRetries = 10
	err = retry.Do(ctx, retry.WithMaxRetries(getSessionMaxRetries, backoff), func(ctx context.Context) error {
		session, err = c.dbClient.Conn(ctx)
		if err != nil {
			retries++
			return retry.RetryableError(err)
		}
		backendPid, err = getBackendPid(ctx, session)
		if err != nil {
			session.Close()
			retries++
			return retry.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		log.Printf("[WARN] AcquireSession failed after 10 retries: %s", err)
		return nil, 0, err
	}

	if retries > 0 {
		log.Printf("[TRACE] AcquireSession succeeded after %d retries", retries)
	}
	return session, backendPid, nil
}

// get the unique postgres identifier for a database session
func getBackendPid(ctx context.Context, session *sql.Conn) (int64, error) {
	var pid int64
	rows, err := session.QueryContext(ctx, "select pg_backend_pid()")
	if err != nil {
		return pid, err
	}
	rows.Next()
	rows.Scan(&pid)
	rows.Close()
	return pid, nil
}

// createTransaction , with a timeout - this may be required if the db client becomes unresponsive
func (c *DbClient) createTransaction(ctx context.Context, session *sql.Conn, retryOnTimeout bool) (tx *sql.Tx, err error) {
	doneChan := make(chan bool)
	go func() {
		tx, err = session.BeginTx(ctx, nil)
		if err != nil {
			err = utils.PrefixError(err, "error creating transaction")
		}
		close(doneChan)
	}()

	select {
	case <-doneChan:
	case <-time.After(time.Second * 5):
		log.Printf("[TRACE] timed out creating a transaction")
		if retryOnTimeout {
			log.Printf("[TRACE] refresh the client and retry")
			// refresh the db client to try to fix the issue
			c.refreshDbClient(ctx)

			// recurse back into this function, passing 'retryOnTimeout=false' to prevent further recursion
			return c.createTransaction(ctx, session, false)
		}
		err = fmt.Errorf("error creating transaction - please restart Steampipe")
	}
	return
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
