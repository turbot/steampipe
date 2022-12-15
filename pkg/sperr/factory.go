package sperr

import (
	"errors"
	"fmt"
)

func New(format string, args ...interface{}) *Error {
	sperr := &Error{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(), // always has a stack
	}
	return sperr
}

func Wrap(err error, options ...ErrorOption) *Error {
	if err == nil {
		return nil
	}
	se := &Error{
		cause: err,
	}

	// does the wrapped error have an sperr.Error down in it's stack?
	// this works because Is repeatedly calls Unwrap and runs Is (if available)
	// in the result. Since Error implements Is - any Error in the stack will
	// return true
	if !errors.Is(err, Error{}) {
		// if this is not an Error, then there's no Stack in the underlying error
		// add it
		se.stack = callers()
	}
	for _, o := range options {
		se = o(se)
	}
	return se
}

func Wrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return Wrap(err, WithMessage(format, args...))
}

func ToError(val interface{}, options ...ErrorOption) *Error {
	if val == nil {
		return nil
	}
	var err error
	if e, ok := val.(error); ok {
		err = Wrap(e)
	} else {
		err = New("%v", val)
	}
	sperr := err.(*Error)
	for _, o := range options {
		err = o(sperr)
	}
	return sperr
}
