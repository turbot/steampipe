package queryresult

import (
	"sync"

	"github.com/turbot/pipe-fittings/v2/queryresult"
)

// Result wraps queryresult.Result[TimingResultStream] with idempotent Close()
type Result struct {
	*queryresult.Result[TimingResultStream]
	closeOnce sync.Once
}

func NewResult(cols []*queryresult.ColumnDef) *Result {
	return &Result{
		Result: queryresult.NewResult[TimingResultStream](cols, NewTimingResultStream()),
	}
}

// Close closes the row channel in an idempotent manner
func (r *Result) Close() {
	r.closeOnce.Do(func() {
		r.Result.Close()
	})
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
