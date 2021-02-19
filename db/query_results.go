package db

import (
	"database/sql"
	"time"
)

type ResultStreamer struct {
	Results      chan *QueryResult
	displayReady chan string
}

func newQueryResults() *ResultStreamer {
	return &ResultStreamer{
		// make buffered channel  so we can always stream a single result
		Results:      make(chan *QueryResult, 1),
		displayReady: make(chan string, 1),
	}
}

func (q *ResultStreamer) streamResult(result *QueryResult) {
	q.Results <- result
}

func (q *ResultStreamer) streamSingleResult(result *QueryResult, onComplete func()) {
	q.Results <- result
	q.Wait()
	onComplete()
	close(q.Results)
}

func (q *ResultStreamer) close() {
	close(q.Results)
}

// Done :: signals that the next QueryResult has been processed
func (q *ResultStreamer) Done() {
	q.displayReady <- ""
}

// Wait :: waits for the next QueryResult to get processed
func (q *ResultStreamer) Wait() {
	<-q.displayReady
}

type QueryResult struct {
	RowChan   *chan []interface{}
	ErrorChan *chan error
	ColTypes  []*sql.ColumnType
	Duration  chan time.Duration
}

// close the channels
func (r QueryResult) Close() {
	close(*r.RowChan)
	close(*r.ErrorChan)

}

func (r QueryResult) StreamRow(rowResult []interface{}) {
	*r.RowChan <- rowResult
}

func (r QueryResult) StreamError(err error) {
	*r.ErrorChan <- err
}

func newQueryResult(colTypes []*sql.ColumnType) *QueryResult {
	rowChan := make(chan []interface{})
	errChan := make(chan error, 1)
	return &QueryResult{

		RowChan:   &rowChan,
		ErrorChan: &errChan,
		ColTypes:  colTypes,
		Duration:  make(chan time.Duration, 1),
	}
}

type SyncQueryResult struct {
	Rows     []interface{}
	ColTypes []*sql.ColumnType
	Duration time.Duration
}
