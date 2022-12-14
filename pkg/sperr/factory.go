package sperr

import "fmt"

func New(format string, args ...interface{}) *Error {
	sperr := &Error{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(),
	}
	return sperr
}

func Wrap(err error) *Error {
	if err == nil {
		return nil
	}
	se := &Error{
		cause: err,
		stack: callers(),
	}
	return se
}

func ToError(val interface{}) *Error {
	if e, ok := val.(error); ok {
		return Wrap(e)
	} else {
		return New("%v", val)
	}
}
