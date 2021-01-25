package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/briandowns/spinner"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// ExecuteSync :: execute a query against this client and wait for the result
func (c *Client) ExecuteSync(query string) (*SyncQueryResult, error) {
	start := time.Now()
	// https://github.com/golang/go/wiki/CodeReviewComments#indent-error-flow
	result, err := c.executeQuery(query)
	if err != nil {
		return nil, err
	}
	syncResult := &SyncQueryResult{ColTypes: result.ColTypes}
	for row := range *result.RowChan {
		syncResult.Rows = append(syncResult.Rows, row)
	}
	syncResult.Duration = time.Since(start)
	return syncResult, nil
}

func (c *Client) executeQuery(query string) (*QueryResult, error) {
	if query == "" {
		return &QueryResult{}, nil
	}

	start := time.Now()

	var rows *sql.Rows
	var err error
	var s *spinner.Spinner

	queryDone := false

	go func() {
		rows, err = c.dbClient.Query(query)
		queryDone = true
	}()

	for {
		if queryDone {
			break
		}
		if time.Since(start) > constants.SpinnerShowTimeout && !s.Active() {
			s = utils.ShowSpinner("Executing query...")
		}
		time.Sleep(50 * time.Millisecond)
	}

	if err != nil {
		if s.Active() {
			// in case the query takes a long time to fail
			utils.StopSpinner(s)
		}
		return nil, err
	}

	if s.Active() {
		utils.UpdateSpinnerMessage(s, "Waiting for results...")
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("error reading columns from query: %v", err)
	}
	cols, err := rows.Columns()

	rowChan := make(chan []interface{})

	result := newQueryResult(&rowChan, colTypes)

	// read the rows in a go routine
	go func() {
		// defer this, so that these get cleaned up even if there is an unforeseen error
		defer func() {
			// close the channel
			close(rowChan)
			// close the rows object
			rows.Close()
		}()

		for rows.Next() {
			// slice of interfaces to receive the row data
			columnValues := make([]interface{}, len(cols))
			// make a slice of pointers to the result to pass to scan
			resultPtrs := make([]interface{}, len(cols)) // A temporary interface{} slice
			for i := range columnValues {
				resultPtrs[i] = &columnValues[i]
			}
			err = rows.Scan(resultPtrs...)
			if err != nil {
				utils.ShowErrorWithMessage(err, "Failed to scan row")
				return
			}
			// populate row data - handle special case types
			result := populateRow(columnValues, colTypes)

			// we have started populating results
			if s.Active() {
				utils.StopSpinner(s)
			}
			rowChan <- result
		}
		// set the time that it took for this one to execute
		result.Duration <- time.Since(start)
	}()

	return result, nil
}
