package queryresult

import (
	"database/sql"
	"time"
)

type RowResult struct {
	Data  []interface{}
	Error error
}
type Result struct {
	RowChan  *chan *RowResult
	ColTypes []*sql.ColumnType
	Duration chan time.Duration
}

// Close closes the row channel
func (r Result) Close() {
	close(*r.RowChan)
}

func (r Result) StreamRow(rowResult []interface{}) {
	*r.RowChan <- &RowResult{Data: rowResult}
}

func (r Result) StreamError(err error) {
	*r.RowChan <- &RowResult{Error: err}
}

func NewQueryResult(colTypes []*sql.ColumnType) *Result {
	rowChan := make(chan *RowResult)
	return &Result{
		RowChan:  &rowChan,
		ColTypes: colTypes,
		Duration: make(chan time.Duration, 1),
	}
}

type SyncQueryResult struct {
	Rows     []interface{}
	ColTypes []*sql.ColumnType
	Duration time.Duration
}
