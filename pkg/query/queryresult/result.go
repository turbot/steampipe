package queryresult

import (
	"time"
)

type TimingMetadata struct {
	RowsFetched       int64
	CachedRowsFetched int64
	HydrateCalls      int64
}

type TimingResult struct {
	Duration time.Duration
	Metadata *TimingMetadata
}
type RowResult struct {
	Data  []interface{}
	Error error
}
type Result struct {
	RowChan      *chan *RowResult
	ColNames     []string
	ColTypes     []string
	TimingResult chan *TimingResult
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

func NewQueryResult(colNames, colTypes []string) *Result {
	rowChan := make(chan *RowResult)
	return &Result{
		RowChan:      &rowChan,
		ColNames:     colNames,
		ColTypes:     colTypes,
		TimingResult: make(chan *TimingResult, 1),
	}
}

type SyncQueryResult struct {
	Rows         []interface{}
	ColNames     []string
	ColTypes     []string
	TimingResult *TimingResult
}
