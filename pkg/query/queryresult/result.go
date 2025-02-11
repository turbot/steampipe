package queryresult

import "github.com/turbot/pipe-fittings/v2/queryresult"

// Result is a type alias for queryresult.Result[TimingResultStream]
type Result = queryresult.Result[TimingResultStream]

func NewResult(cols []*queryresult.ColumnDef) *Result {
	return queryresult.NewResult[TimingResultStream](cols, NewTimingResultStream())
}

// ResultStreamer is a type alias for queryresult.ResultStreamer[TimingResultStream]
type ResultStreamer = queryresult.ResultStreamer[TimingResultStream]

func NewResultStreamer() *ResultStreamer {
	return queryresult.NewResultStreamer[TimingResultStream]()
}
