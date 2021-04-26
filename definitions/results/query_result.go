/**
	This package is for all interfaces that are imported in multiple packages in the
	code base

	This package MUST never import any other `steampipe` package
**/
package results

import (
	"context"
	"database/sql"
	"time"
)

type RowResult struct {
	Data  []interface{}
	Error error
}
type QueryResult struct {
	RowChan  *chan *RowResult
	ColTypes []*sql.ColumnType
	Duration chan time.Duration
}

// close the channels
func (r QueryResult) Close() {
	close(*r.RowChan)
}

func (r QueryResult) StreamRow(rowResult []interface{}) {
	*r.RowChan <- &RowResult{Data: rowResult}
}

func (r QueryResult) StreamError(err error) {
	*r.RowChan <- &RowResult{Error: err}
}

func NewQueryResult(colTypes []*sql.ColumnType, ctx *context.Context) *QueryResult {
	rowChan := make(chan *RowResult)
	return &QueryResult{
		RowChan:  &rowChan,
		QueryCtx: ctx,
		ColTypes: colTypes,
		Duration: make(chan time.Duration, 1),
	}
}

type SyncQueryResult struct {
	Rows     []interface{}
	ColTypes []*sql.ColumnType
	Duration time.Duration
}
