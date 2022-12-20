package sperr

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func New(format string, args ...interface{}) *Error {
	sperr := &Error{
		msg:    fmt.Sprintf(format, args...),
		stack:  callers(), // always has a stack
		isRoot: true,
	}
	return sperr
}

func Wrap(err error) *Error {
	if err == nil {
		return nil
	}
	if castedErr, ok := err.(*Error); ok {
		return castedErr
	}

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

	// default to a blank error string
	msg := ""
	// and let the underlying error bubble
	setRoot := false

	// if we were able to deduce a message
	if len(constructedMsg) > 0 {
		// use it
		msg = constructedMsg
		// and hide the underlying error
		setRoot = true
	}

	return &Error{
		cause:  err,
		msg:    msg,
		isRoot: setRoot, // set this to false if the underlying error is unknown
		stack:  callers(),
	}
}

func Wrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return Wrap(err).WithMessage(format, args...)
}

func ToError(val interface{}) *Error {
	if val == nil {
		return nil
	}
	var err error
	if e, ok := val.(error); ok {
		err = Wrap(e)
	} else {
		err = New("%v", val)
	}
	sperr := err.(*Error)
	return sperr
}
