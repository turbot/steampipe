package sperr

import (
	"fmt"
)

// WithMessage wraps Error and sets the provided message in the new error
func (e *Error) WithMessage(format string, args ...interface{}) *Error {
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

// WithDetail wraps Error and sets the provided detail message in the new error
func (e *Error) WithDetail(format string, args ...interface{}) *Error {
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

// AsRootMessage sets this error as the root error in the error stack.
// When an Error is set as root, all child errors are hidden from display
func (e *Error) AsRootMessage() *Error {
	if e == nil {
		return nil
	}
	e.isRootMessage = true
	return e
}
