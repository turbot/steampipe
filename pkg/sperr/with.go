package sperr

import "fmt"

// WithDiagnostic updates Error and sets the provided message in the error
func (e *Error) WithDetail(format string, args ...interface{}) *Error {
	e.detail = fmt.Sprintf(format, args...)
	return e
}

// HideChildErrors forces child errors to be hidden from the message output
func (e *Error) HideChildErrors() *Error {
	e.hideChildErrors = true
	return e
}
