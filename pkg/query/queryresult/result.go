package queryresult

import (
	"sync"
	"sync/atomic"

	"github.com/turbot/pipe-fittings/v2/queryresult"
)

// Result wraps queryresult.Result[TimingResultStream] with idempotent Close()
// and synchronization to prevent race between StreamRow and Close
type Result struct {
	*queryresult.Result[TimingResultStream]
	closeOnce sync.Once
	closed    atomic.Bool
}

func NewResult(cols []*queryresult.ColumnDef) *Result {
	return &Result{
		Result: queryresult.NewResult[TimingResultStream](cols, NewTimingResultStream()),
	}
}

// Close closes the row channel in an idempotent manner
func (r *Result) Close() {
	r.closeOnce.Do(func() {
		r.closed.Store(true)
		r.Result.Close()
	})
}

// StreamRow wraps the underlying StreamRow with synchronization to prevent panic on closed channel
func (r *Result) StreamRow(row []interface{}) {
	// Check if already closed - if so, silently drop the row
	if r.closed.Load() {
		return
	}

	// Use recover to gracefully handle the race where channel closes between check and send
	defer func() {
		if r := recover(); r != nil {
			// Channel was closed between our check and the send - this is okay, just drop the row
		}
	}()

	r.Result.StreamRow(row)
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
