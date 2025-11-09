package queryresult

import (
	"sync"

	"github.com/turbot/pipe-fittings/v2/queryresult"
)

// Result wraps queryresult.Result[TimingResultStream] to provide safe Close() behavior
type Result struct {
	*queryresult.Result[TimingResultStream]
	closeOnce sync.Once
	closed    bool
	closeMu   sync.RWMutex
}

func NewResult(cols []*queryresult.ColumnDef) *Result {
	return &Result{
		Result: queryresult.NewResult[TimingResultStream](cols, NewTimingResultStream()),
	}
}

// Close closes the row channel safely, ensuring it can be called multiple times
// without panicking. Subsequent calls after the first are no-ops.
func (r *Result) Close() {
	r.closeOnce.Do(func() {
		r.closeMu.Lock()
		r.closed = true
		r.closeMu.Unlock()
		r.Result.Close()
	})
}

// IsClosed returns true if Close() has been called
func (r *Result) IsClosed() bool {
	r.closeMu.RLock()
	defer r.closeMu.RUnlock()
	return r.closed
}

// ResultStreamer is a type alias for queryresult.ResultStreamer[TimingResultStream]
type ResultStreamer = queryresult.ResultStreamer[TimingResultStream]

func NewResultStreamer() *ResultStreamer {
	return queryresult.NewResultStreamer[TimingResultStream]()
}
