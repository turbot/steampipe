package sperr

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// New creates a new sperr.Error with a stack
// It is recommended that `New` be called from the place where the actual
// error occurs and not do define errors.
// This is because sperr.Error is a stateful construct which also
// encapsulates the stacktrace at the time of creation.
//
// For converting generic errors to sperr.Error the recommended usage pattern
// is sperr.Wrap or sperr.Wrapf
func New(format string, args ...interface{}) *Error {
	sperr := &Error{
		message: fmt.Sprintf(format, args...),
		stack:   callers(), // always has a stack
		isRoot:  true,
	}
	return sperr
}

// Wrap creates a new sperr.Error if the `error` that is being wrapped
// is not an sperr.Error
//
// When wrapping an `error` this also adds a stacktrace into the new `Error` object
func Wrap(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}

	msg := inferMessageFromError(err)
	// hide the child error if we could infer a message from the error
	setRoot := len(msg) > 0

	return &Error{
		cause:   err,
		message: msg,
		isRoot:  setRoot,
		stack:   callers(),
	}
}

func Wrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		cause:   err,
		message: fmt.Sprintf(format, args...),
		stack:   callers(),
	}
}

func ToError(val interface{}) *Error {
	if val == nil {
		return nil
	}
	var err *Error
	if e, ok := val.(error); ok {
		err = Wrap(e)
	} else {
		err = New("%v", val)
	}
	// overwrite the stack to the stack for this function call
	// so that ToError doesn't end up in the stack
	err.stack = callers()
	return err
}

func inferMessageFromError(err error) string {
	constructedMsg := ""
	// if this is one of the errors from the SQL stdlib
	if errors.Is(err, sql.ErrConnDone) ||
		errors.Is(err, sql.ErrNoRows) ||
		errors.Is(err, sql.ErrTxDone) {
		// all of these errors are errors.New with a string argument all prefixed with 'sql: '
		constructedMsg = strings.TrimPrefix(err.Error(), "sql:")
	}

	// Errors we need to wrap around and produce beautiful messages
	// context.DeadlineExceeded
	// context.TimeoutExceeded
	// sql.ErrConnDone
	// sql.ErrNoRows
	// sql.ErrTxDone
	// plugin.ErrChecksumsDoNotMatch
	// plugin.ErrProcessNotFound
	// plugin.ErrSecureConfigAndReattach
	// plugin.ErrSecureConfigNoChecksum
	// plugin.ErrSecureConfigNoHash

	return strings.TrimSpace(constructedMsg)
}
