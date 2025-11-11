package queryresult

import (
	"sync"

	"github.com/turbot/pipe-fittings/v2/queryresult"
)

// Result wraps queryresult.Result[TimingResultStream] with idempotent Close()
type Result struct {
	*queryresult.Result[TimingResultStream]
	closeOnce sync.Once
	mu        sync.RWMutex
	closed    bool
}

func NewResult(cols []*queryresult.ColumnDef) *Result {
	return &Result{
		Result: queryresult.NewResult[TimingResultStream](cols, NewTimingResultStream()),
	}
}

// Close closes the row channel in an idempotent manner
func (r *Result) Close() {
	r.closeOnce.Do(func() {
		r.mu.Lock()
		r.closed = true
		r.mu.Unlock()
		r.Result.Close()
	})
}

// StreamRow safely sends a row to the RowChan, checking if it's closed first
func (r *Result) StreamRow(rowResult []interface{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.closed {
		r.Result.StreamRow(rowResult)
	}
}

// StreamError safely sends an error to the RowChan, checking if it's closed first
func (r *Result) StreamError(err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.closed {
		r.Result.StreamError(err)
	}
}

// WrapResult wraps a pipe-fittings Result with our wrapper that has idempotent Close
func WrapResult(r *queryresult.Result[TimingResultStream]) *Result {
	return &Result{
		Result: r,
	}
}

// ResultStreamer is a type alias for queryresult.ResultStreamer[TimingResultStream]
type ResultStreamer = queryresult.ResultStreamer[TimingResultStream]

func NewResultStreamer() *ResultStreamer {
	return queryresult.NewResultStreamer[TimingResultStream]()
}
