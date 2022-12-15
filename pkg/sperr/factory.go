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

func Wrap(err error, options ...ErrorOption) *Error {
	if err == nil {
		return nil
	}
	se := &Error{
		cause: err,
		stack: callers(),
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
