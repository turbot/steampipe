package sperr

import (
	"fmt"
	"io"
)

type Error struct {
	stack           *stack
	cause           error
	detail          string
	msg             string
	hideChildErrors bool
}

// Cause will retrieves the underlying root error in the error stack
// Cause will recursively retrieve
// the topmost error that does not have a cause, which is assumed to be
// the original cause.
//
// TODO: also have a global Cause which calls into this if it is an sperr.Error
func (e Error) Cause() error {
	type hasCause interface {
		Cause() error
	}
	if cause, ok := e.cause.(hasCause); ok {
		return cause.Cause()
	}
	return e.cause
}

func (e Error) Stack() StackTrace {
	type hasStack interface {
		Stack() StackTrace
	}
	if cause, ok := e.cause.(hasStack); ok {
		return cause.Stack()
	}
	return e.stack.StackTrace()
}

func (e Error) Unwrap() error { return e.cause }
func (e Error) Is(target error) bool {
	_, ok := target.(Error)
	return ok
}

func (e Error) Error() string {
	res := e.msg
	if e.cause != nil && !e.hideChildErrors {
		res = e.msg + ": " + e.cause.Error()
	}
	return res
}

// All error values returned from this package implement fmt.Formatter and can
// be formatted by the fmt package. The following verbs are supported:
//
//	%s    print the error. If the error has a Cause it will be
//	      printed recursively.
//	%v    see %s
//	%+v   extended format. Each Frame of the error's StackTrace will
//	      be printed in detail.
func (e Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, e.msg)
			e.Stack().Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.msg)
	case 'q':
		fmt.Fprintf(s, "%q", e.msg)
	}
}
