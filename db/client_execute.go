package db

import (
	"fmt"
	"time"

	"github.com/turbot/steampipe/constants"

	"github.com/briandowns/spinner"

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

	spinner := utils.CreateSpinner("Executing query...")
	// channel to flag to spinner goroutine that the query has run
	queryDone := make(chan bool, 1)
	// start spinner after a short delay
	go startSpinnerAfterDelay(spinner, queryDone)

	rows, err := c.dbClient.Query(query)
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
	cols, err := rows.Columns()

	rowChan := make(chan []interface{})

	result := newQueryResult(&rowChan, colTypes)

	rowCount := 0

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
			rowChan <- result

			// update the spinner message with the count of rows that have already been fetched
			utils.UpdateSpinnerMessage(spinner, fmt.Sprintf("Waiting for results... Fetched: %5d", rowCount))
			rowCount++
		}

		// we are done fetching results. time for display. remove the spinner
		utils.StopSpinner(spinner)

		// set the time that it took for this one to execute
		result.Duration <- time.Since(start)
	}()

	return result, nil
}

func startSpinnerAfterDelay(s *spinner.Spinner, queryDone chan bool) {
	{
		for {
			select {
			case <-queryDone:
				return
			case <-time.After(constants.SpinnerShowTimeout):
				s.Start()
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}
