package sperr

import "fmt"

// WithMessage wraps Error and sets the provided message in the new error
func (e *Error) WithMessage(format string, args ...interface{}) *Error {
	newErr := &Error{
		msg: fmt.Sprintf(format, args...),
	}
	return newErr
}

// WithRootMessage wraps Error and sets the provided message as the final message in the new error.
// When a root message is set, all child errors are hidden from display
func (e *Error) WithRootMessage(format string, args ...interface{}) *Error {
	newErr := &Error{
		msg: fmt.Sprintf(format, args...),
	}
	return newErr
}

// WithDetail wraps Error and sets the provided detail message in the new error
func (e *Error) WithDetail(format string, args ...interface{}) *Error {
	newErr := &Error{
		detail: fmt.Sprintf(format, args...),
	}
	return newErr
}
