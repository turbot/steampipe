package sperr

import "fmt"

type ErrorOption = func(*Error) *Error

// WithMessage wraps an Error produces a new Error with the given message set
func WithMessage(format string, args ...interface{}) ErrorOption {
	return func(e *Error) *Error {
		n := Wrap(e)
		n.msg = fmt.Sprintf(format, args...)
		return n
	}
}
