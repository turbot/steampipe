package sperr

import (
	"fmt"
	"runtime"
)

type CallFn struct {
	Function *runtime.Func
	File     string
	Line     int
}

type SteampipeError struct {
	// make this compatible with the golang error interface
	error

	callContext *CallFn
	severity    Severity
	diagnostic  string
	message     string
	terminate   bool
}

func (e *SteampipeError) With(options ...SteampipeErrorOption) {
	for _, o := range options {
		o(e)
	}
}

func New(format string, args ...interface{}) *SteampipeError {
	sperr := &SteampipeError{
		diagnostic: fmt.Sprintf(format, args...),
		severity:   Error,
	}
	if pc, file, line, ok := runtime.Caller(0); ok {
		fn := runtime.FuncForPC(pc)
		c := &CallFn{
			Function: fn,
			File:     file,
			Line:     line,
		}
		sperr.callContext = c
	}
	return sperr
}

func Wrap(err error, options ...SteampipeErrorOption) *SteampipeError {
	if err == nil {
		return nil
	}
	se := &SteampipeError{
		error:    err,
		severity: Error,
	}
	if pc, file, line, ok := runtime.Caller(0); ok {
		fn := runtime.FuncForPC(pc)
		c := &CallFn{
			Function: fn,
			File:     file,
			Line:     line,
		}
		se.callContext = c
	}

	for _, o := range options {
		o(se)
	}
	return se
}

func (e *SteampipeError) Error() string {
	if e.error != nil {
		return e.error.Error()
	}
	return e.diagnostic
}
