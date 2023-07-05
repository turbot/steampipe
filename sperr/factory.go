package sperr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
)

// New creates a new sperr.Error with a stack
// It is recommended that `New` be called from the place where the actual
// error occurs and not do define errors.
// This is because sperr.Error is a stateful construct which also
// encapsulates the stacktrace at the time of creation.
//
// For converting generic errors to sperr.Error the recommended usage pattern
// is sperr.Wrap or sperr.Wrapf
func New(format string, args ...interface{}) error {
	sperr := &Error{
		message: fmt.Sprintf(format, args...),
		stack:   callers(), // always has a stack
	}
	return sperr
}

// Wrap creates a new sperr.Error if the `error` that is being wrapped
// is not an sperr.Error
//
// When wrapping an error this also adds a stacktrace into the new `Error` object
func Wrap(err error, options ...Option) error {
	if err == nil {
		return nil
	}
	res := wrap(err, options...)
	// if the error we wrapped was not an sperr,
	// we need to set the stack
	if _, ok := err.(*Error); !ok {
		res.stack = callers()
	}
	return res
}

func WrapWithMessage(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	res := wrap(err, WithMessage(format, args...))
	// if the error we wrapped was not an sperr,
	// we need to set the stack
	if _, ok := err.(*Error); !ok {
		res.stack = callers()
	}
	return res
}

func WrapWithRootMessage(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	res := wrap(err, WithRootMessage(format, args...))
	// if the error we wrapped was not an sperr,
	// we need to set the stack
	if _, ok := err.(*Error); !ok {
		res.stack = callers()
	}
	return res
}

func ToError(val interface{}, options ...Option) error {
	// we need to do a IsNil, since the value of the interface may be nil
	// and not the interface itself
	if helpers.IsNil(val) {
		return nil
	}
	var res *Error

	if e, ok := val.(error); ok {
		res = Wrap(e, options...).(*Error)
	} else {
		res = New("%v", val).(*Error)
	}
	for _, opt := range options {
		res = opt(res)
	}
	// if the error we wrapped was not an sperr,
	// we need to set the stack
	if _, ok := val.(*Error); !ok {
		res.stack = callers()
	}
	return res
}

// err MUST be non-nil
func wrap(err error, options ...Option) *Error {
	// we know err will always be non-nil - callers need to make sure of that
	var e *Error
	if x, ok := err.(*Error); ok {
		e = x
	} else {
		msg := inferMessageFromError(err)
		// hide the child error if we could infer a message from the error
		isRoot := len(msg) > 0
		e = &Error{
			cause:         err,
			message:       msg,
			isRootMessage: isRoot,
			// do not set the stack here. We need rely on the calling function
			// to set the stack so that the call stack remains clean
		}
	}
	for _, opt := range options {
		e = opt(e)
	}
	return e
}

func inferMessageFromError(err error) string {
	constructedMsg := ""
	// if this is one of the errors from the SQL stdlib
	if errors.Is(err, sql.ErrConnDone) ||
		errors.Is(err, sql.ErrNoRows) ||
		errors.Is(err, sql.ErrTxDone) {
		// all of these errors are errors.New with a string argument all prefixed with 'sql: '
		constructedMsg = strings.TrimPrefix(err.Error(), "sql:")
	} else if errors.Is(err, context.DeadlineExceeded) {
		constructedMsg = "exceeded allowed timeout"
	} else if errors.Is(err, context.Canceled) {
		constructedMsg = "operation cancelled"
	}

	// Errors we need to wrap around and produce beautiful messages
	// plugin.ErrChecksumsDoNotMatch
	// plugin.ErrProcessNotFound
	// plugin.ErrSecureConfigAndReattach
	// plugin.ErrSecureConfigNoChecksum
	// plugin.ErrSecureConfigNoHash

	return strings.TrimSpace(constructedMsg)
}
