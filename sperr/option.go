package sperr

import "fmt"

type Option func(e *Error) *Error

func WithDetail(format string, args ...interface{}) Option {
	return func(e *Error) *Error {
		if e == nil {
			return nil
		}
		res := e
		// if there's a detail, wrap this error and set the detail the new error
		if len(e.detail) > 0 {
			res = &Error{
				cause: e,
			}
		}
		res.detail = fmt.Sprintf(format, args...)
		return res
	}
}

func WithMessage(format string, args ...interface{}) Option {
	return func(e *Error) *Error {
		if e == nil {
			return nil
		}
		res := e
		// if there's a message, wrap this error and set the message on the new error
		if len(e.message) > 0 {
			res = &Error{
				cause: e,
			}
		}
		res.message = fmt.Sprintf(format, args...)
		return res
	}
}

func WithRootMessage(format string, args ...interface{}) Option {
	return func(e *Error) *Error {
		e = WithMessage(format, args...)(e)
		e.isRootMessage = true
		return e
	}
}
