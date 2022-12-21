package sperr

import (
	"fmt"
	"io"
	"strings"
)

type Error struct {
	stack  *stack
	cause  error
	detail string
	msg    string
	isRoot bool
}

// RootCause will retrieves the underlying root error in the error stack
// RootCause will recursively retrieve
// the topmost error that does not have a cause, which is assumed to be
// the original cause.
func (e Error) RootCause() error {
	type hasCause interface {
		Cause() error
	}
	if cause, ok := e.cause.(hasCause); ok {
		return cause.Cause()
	}
	return e.cause
}

// Stack retrieves the stack trace of the absolute underlying sperr.Error
func (e Error) Stack() StackTrace {
	type hasStack interface {
		Stack() StackTrace
	}
	if cause, ok := e.cause.(hasStack); ok {
		return cause.Stack()
	}
	if e.stack == nil {
		panic("sperr: stack cannot be nil")
	}
	return e.stack.StackTrace()
}

// Unwrap returns the immediately underlying error
func (e Error) Unwrap() error { return e.cause }

func (e Error) Error() (str string) {
	res := []string{}
	if len(e.msg) > 0 {
		res = append(res, e.msg)
	}
	if e.isRoot || e.cause == nil {
		return e.msg
	}
	if e.cause != nil && len(e.cause.Error()) > 0 {
		res = append(res, e.cause.Error())
	}
	return strings.Join(res, ":")
}

func (e Error) Detail() string {
	type hasDetail interface {
		Detail() string
	}
	res := []string{}
	if len(e.detail) > 0 {
		// if this is available - the underlying error will always be a sperr
		res = append(res, fmt.Sprintf("%s :: %s", e.msg, e.detail))
	}
	if e.cause != nil && len(e.cause.Error()) > 0 {
		if asD, ok := e.cause.(hasDetail); ok {
			res = append(res, asD.Detail())
		} else {
			if len(e.Error()) > 0 {
				res = append(res, e.Error())
			}
		}
	}
	return strings.Join(res, "\n|-- ")
}

// All error values returned from this package implement fmt.Formatter and can
// be formatted by the fmt package. The following verbs are supported:
//
//			%s    print the error. If the error has a Cause it will be
//			      printed recursively.
//			%v    see %s
//			%+v   detailed format - includes messages and detail.
//	    %#v   Each Frame of the error's StackTrace will be printed in detail.
//		  %q		a double-quoted string safely escaped with Go syntax
//
// TODO: add Details for +
func (e Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		io.WriteString(s, e.Error())
		io.WriteString(s, "\n")
		if s.Flag('+') {
			io.WriteString(s, "\nDetails:\n")
			io.WriteString(s, e.Detail())
			io.WriteString(s, "\n")
		}
		if s.Flag('#') {
			io.WriteString(s, "\nStack:")
			e.Stack().Format(s, verb)
			io.WriteString(s, "\n")
		}
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}
