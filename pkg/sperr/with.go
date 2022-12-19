package sperr

import "fmt"

// WithMessage wraps Error and sets the provided message in the new error
func (e *Error) WithMessage(format string, args ...interface{}) *Error {
	if e == nil {
		return nil
	}
	newErr := &Error{
		msg:   fmt.Sprintf(format, args...),
		cause: e,
	}
	return newErr
}

// AsRoot wraps Error and sets the provided message as the final message in the new error.
// When an Error is set as root, all child errors are hidden from display
func (e *Error) AsRoot(format string, args ...interface{}) *Error {
	newErr := &Error{
		isRoot: true,
		cause:  e,
	}
	return newErr
}

// WithDetail wraps Error and sets the provided detail message in the new error
func (e *Error) WithDetail(format string, args ...interface{}) *Error {
	newErr := &Error{
		detail: fmt.Sprintf(format, args...),
		cause:  e,
	}
	return newErr
}
