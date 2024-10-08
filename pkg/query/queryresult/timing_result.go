package queryresult

type TimingResultStream struct {
	Stream chan *TimingResult
}

// GetTiming implements TimingContainer
func (t TimingResultStream) GetTiming() any {
	return <-t.Stream
}

func (t TimingResultStream) SetTiming(result *TimingResult) {
	t.Stream <- result
}

func NewTimingResultStream() TimingResultStream {
	return TimingResultStream{
		Stream: make(chan *TimingResult, 1),
	}
}

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

// GetTiming implements TimingContainer
func (t TimingResult) GetTiming() any {
	// just return ourselves
	return t
}
