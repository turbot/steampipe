package sperr

import "fmt"

// WithDiagnostic adds the provided diagnostic message in the error
func (e *Error) WithDiagnostic(format string, args ...interface{}) *Error {
	e.diagnostic = fmt.Sprintf(format, args...)
	return e
}

// WithMessage  the provided message in the error
func (e *Error) WithMessage(format string, args ...interface{}) *Error {
	e.message = fmt.Sprintf(format, args...)
	return e
}

// HideChildErrors forces child errors to be hidden from the message output
func (e *Error) HideChildErrors() *Error {
	e.hideChildErrors = true
	return e
}

/**

We use chaining patterns instead of Options functions, since
we want to support both `New` and `Wrapping` using the same interface

**/
