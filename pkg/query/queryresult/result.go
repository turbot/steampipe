package queryresult

type TimingResult struct {
	DurationMs          int64              `json:"duration_ms"`
	Scans               []*ScanMetadataRow `json:"scans"`
	ScanCount           int64              `json:"scan_count,omitempty"`
	RowsReturned        int64              `json:"rows_returned"`
	UncachedRowsFetched int64              `json:"uncached_rows_fetched"`
	CachedRowsFetched   int64              `json:"cached_rows_fetched"`
	HydrateCalls        int64              `json:"hydrate_calls"`
	ConnectionCount     int64              `json:"connection_count"`
}

func (r *TimingResult) Initialise(summary *QueryRowSummary, scans []*ScanMetadataRow) {
	r.ScanCount = summary.ScanCount
	r.ConnectionCount = summary.ConnectionCount
	r.UncachedRowsFetched = summary.UncachedRowsFetched
	r.CachedRowsFetched = summary.CachedRowsFetched
	r.HydrateCalls = summary.HydrateCalls
	// populate scans - note this may not be all scans
	r.Scans = scans
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
