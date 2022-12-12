package sperr

import (
	"fmt"
	"runtime"
)

type CallContext struct {
	ProgramCounter uintptr
	Function       *runtime.Func
	File           string
	Line           int
}

type SteampipeError struct {
	callContext     *CallContext
	originalError   error
	diagnostic      string
	message         string
	hideChildErrors bool
}

func (e *SteampipeError) Error() string {
	if sperr, ok := e.originalError.(*SteampipeError); ok && !sperr.hideChildErrors {
		return fmt.Sprintf("%s:%s", e.message, sperr.Error())
	}
	return e.message
}

func getCallContext() *CallContext {
	callContext := &CallContext{
		File: "unknown",
		Line: -1,
	}
	if pc, file, line, ok := runtime.Caller(2 /*skip this function and the function in this package*/); ok {
		fn := runtime.FuncForPC(pc)
		callContext = &CallContext{
			ProgramCounter: pc,
			Function:       fn,
			File:           file,
			Line:           line,
		}
	}
	return callContext
}

func New(format string, args ...interface{}) *SteampipeError {
	sperr := &SteampipeError{
		message:     fmt.Sprintf(format, args...),
		callContext: getCallContext(),
	}
	return sperr
}

func Wrap(err error) *SteampipeError {
	if err == nil {
		return nil
	}
	se := &SteampipeError{
		originalError: err,
		callContext:   getCallContext(),
	}
	return se
}
