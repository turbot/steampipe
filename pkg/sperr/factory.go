package sperr

import (
	"fmt"
)

func New(format string, args ...interface{}) *Error {
	sperr := &Error{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(), // always has a stack
	}
	return sperr
}

func Wrap(err error) *Error {
	if err == nil {
		return nil
	}
	if castedErr, ok := err.(*Error); ok {
		return castedErr
	}
	se := &Error{
		cause: err,
		stack: callers(),
	}
	return se
}

func Wrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return Wrap(err).WithMessage(format, args...)
}

func ToError(val interface{}) *Error {
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
	return sperr
}
