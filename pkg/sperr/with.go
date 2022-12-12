package sperr

import "fmt"

// WithDiagnostic adds the provided diagnostic message in the error
func (e *SteampipeError) WithDiagnostic(format string, args ...interface{}) *SteampipeError {
	e.diagnostic = fmt.Sprintf(format, args...)
	return e
}

// WithMessage  the provided message in the error
func (e *SteampipeError) WithMessage(format string, args ...interface{}) *SteampipeError {
	e.message = fmt.Sprintf(format, args...)
	return e
}

// HideChildErrors forces child errors to be hidden from the message output
func (e *SteampipeError) HideChildErrors() *SteampipeError {
	e.hideChildErrors = true
	return e
}

/**

We use chaining patterns instead of Options functions, since
we want to support both `New` and `Wrapping` using the same interface

**/
