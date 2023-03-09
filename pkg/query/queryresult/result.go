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
	Cols         []*ColumnDef
	TimingResult chan *TimingResult
}

func NewResult(cols []*ColumnDef) *Result {

	rowChan := make(chan *RowResult)
	return &Result{
		RowChan:      &rowChan,
		Cols:         cols,
		TimingResult: make(chan *TimingResult, 1),
	}
}

// IsExportSourceData implements ExportSourceData
func (*Result) IsExportSourceData() {}

// Close closes the row channel
func (r *Result) Close() {
	close(*r.RowChan)
}

func (r *Result) StreamRow(rowResult []interface{}) {
	*r.RowChan <- &RowResult{Data: rowResult}
}
func (r *Result) StreamError(err error) {
	*r.RowChan <- &RowResult{Error: err}
}

type SyncQueryResult struct {
	Rows         []interface{}
	Cols         []*ColumnDef
	TimingResult *TimingResult
}
