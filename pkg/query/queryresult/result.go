package queryresult

type TimingResult struct {
	DurationMs   int64              `json:"duration_ms"`
	Scans        []*ScanMetadataRow `json:"scans"`
	RowsReturned int64              `json:"rows_returned"`
	RowsFetched  int64              `json:"rows_fetched"`
	HydrateCalls int64              `json:"hydrate_calls"`
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
