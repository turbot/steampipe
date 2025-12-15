package queryresult

import (
	"sync"

	"github.com/turbot/pipe-fittings/v2/queryresult"
)

// Result wraps queryresult.Result[TimingResultStream] with idempotent Close()
// and synchronization to prevent race between StreamRow and Close
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

// StreamRow wraps the underlying StreamRow with synchronization
func (r *Result) StreamRow(row []interface{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.closed {
		r.Result.StreamRow(row)
	}
}

// WrapResult wraps a pipe-fittings Result with our wrapper that has idempotent Close
func WrapResult(r *queryresult.Result[TimingResultStream]) *Result {
	if r == nil {
		return nil
	}
	return &Result{
		Result: r,
	}
}

// ResultStreamer is a type alias for queryresult.ResultStreamer[TimingResultStream]
type ResultStreamer = queryresult.ResultStreamer[TimingResultStream]

func NewResultStreamer() *ResultStreamer {
	return queryresult.NewResultStreamer[TimingResultStream]()
}
