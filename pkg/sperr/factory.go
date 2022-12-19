package sperr

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func New(format string, args ...interface{}) *Error {
	sperr := &Error{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(), // always has a stack
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

	msg := err.Error()

	// if this is one of the errors from the SQL stdlib
	if errors.Is(err, sql.ErrConnDone) || errors.Is(err, sql.ErrNoRows) || errors.Is(err, sql.ErrTxDone) {
		msg = strings.TrimPrefix(err.Error(), "sql:")
	}

	return &Error{
		cause:  err,
		msg:    msg,
		isRoot: true, // set this to false if the underlying error is unknown
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
