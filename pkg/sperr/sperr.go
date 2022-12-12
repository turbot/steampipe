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

type Error struct {
	callContext     *CallContext
	originalError   error
	diagnostic      string
	message         string
	hideChildErrors bool
}

func (e *Error) Error() string {
	// if there's an underlying error
	if e.originalError != nil {
		// and if it is a sperr.Error
		if sperr, ok := e.originalError.(*Error); ok {
			// and if it wants to hide it's children
			if sperr.hideChildErrors {
				// return the error message itself
				return e.message
			}
		}

		// return the message with the message from it's children
		return fmt.Sprintf("%s:%s", e.message, e.originalError.Error())
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

func New(format string, args ...interface{}) *Error {
	sperr := &Error{
		message:     fmt.Sprintf(format, args...),
		callContext: getCallContext(),
	}
	return sperr
}

func Wrap(err error) *Error {
	if err == nil {
		return nil
	}
	se := &Error{
		originalError: err,
		callContext:   getCallContext(),
	}
	return se
}
