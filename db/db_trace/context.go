package db_trace

import "context"

// unique type to prevent assignment.
type dbTraceContextKetType struct{}

func DBTraceFromContext(ctx context.Context) *DBTrace {
	v := ctx.Value(dbTraceContextKetType{})
	if v == nil {
		return noOpTrace
	}
	trace, _ := v.(*DBTrace)
	return trace
}

// WithDBTrace returns a new context based on the provided parent
// ctx. Service operations made with the returned context will use
// the provided trace hooks, in addition to any previous hooks
// registered with ctx. Any hooks defined in the provided trace will
// be called first.
func WithDBTrace(ctx context.Context, trace *DBTrace) context.Context {
	if trace == nil {
		panic("nil trace")
	}
	trace.fill()
	old := DBTraceFromContext(ctx)
	trace.compose(old)
	ctx = context.WithValue(ctx, dbTraceContextKetType{}, trace)
	return ctx
}
