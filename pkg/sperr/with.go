package sperr

import (
	"fmt"
	"log"
)

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

// WithDetail wraps Error and sets the provided detail message in the new error
func (e *Error) WithDetail(format string, args ...interface{}) *Error {
	newErr := &Error{
		detail: fmt.Sprintf(format, args...),
		cause:  e,
	}
	return newErr
}

// SetDetail sets the given message as the detail in the Error
// It is not suggested, but is legal, to call SetDetail on an error which already
// has a detail. In such a case, SetDetail will raise a warning, but will set the
// detail none-the-less
func (e *Error) SetDetail(format string, args ...interface{}) *Error {
	if len(e.detail) > 0 {
		log.Println("[WARN] setting detail on an Error which already has a detail set to", e.detail)
	}
	e.detail = fmt.Sprintf(format, args...)
	return e
}

// SetAsRoot sets this error as the root error in the error stack.
// When an Error is set as root, all child errors are hidden from display
func (e *Error) SetAsRoot() *Error {
	newErr := &Error{
		isRoot: true,
		cause:  e,
	}
	return newErr
}
