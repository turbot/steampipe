package sperr

import (
	"fmt"
	"io"
	"strings"
)

type Error struct {
	stack         *stack
	cause         error
	detail        string
	message       string
	isRootMessage bool
}

// RootCause will retrieves the underlying root error in the error stack
// RootCause will recursively retrieve
// the topmost error that does not have a cause, which is assumed to be
// the original cause.
func (e *Error) RootCause() error {
	if e == nil {
		return nil
	}
	type hasCause interface {
		Cause() error
	}
	if e.cause == nil {
		// return self if we don't have a cause
		// I was created with New
		return e
	}
	if cause, ok := e.cause.(hasCause); ok {
		return cause.Cause()
	}
	return e.cause
}

// Cause returns the underlying cause of this error. Maybe <nil> if this was created with New
func (e *Error) Cause() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Stack retrieves the stack trace of the absolute underlying sperr.Error
func (e *Error) Stack() StackTrace {
	if e == nil {
		return nil
	}
	type hasStack interface {
		Stack() StackTrace
	}
	if cause, ok := e.cause.(hasStack); ok {
		return cause.Stack()
	}
	if e.stack == nil {
		return StackTrace{}
	}
	return e.stack.StackTrace()
}

// Unwrap returns the immediately underlying error
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	res := []string{}
	if len(e.message) > 0 {
		res = append(res, e.message)
	}
	if e.isRootMessage || e.cause == nil {
		return e.message
	}
	if e.cause != nil && len(e.cause.Error()) > 0 {
		res = append(res, e.cause.Error())
	}
	return strings.Join(res, " : ")
}

func (e *Error) Detail() string {
	if e == nil {
		return ""
	}
	type hasDetail interface {
		Detail() string
	}
	res := []string{}
	if len(e.detail) > 0 {
		// if this is available - the underlying error will always be a sperr
		res = append(res, fmt.Sprintf("%s :: %s", e.message, e.detail))
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
func (e *Error) Format(s fmt.State, verb rune) {
	if e == nil {
		return
	}
	switch verb {
	case 'v':
		io.WriteString(s, e.Error())
		io.WriteString(s, "\n")

		printStack := s.Flag('#')
		printDetail := printStack || s.Flag('+')

		if printDetail {
			io.WriteString(s, "\nDetails:\n")
			io.WriteString(s, e.Detail())
			io.WriteString(s, "\n")
		}

		if printStack {
			io.WriteString(s, "\nStack:")
			io.WriteString(s, fmt.Sprintf("%+v", e.Stack()))
			io.WriteString(s, "\n")
		}
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		// fallback to the standard %q for consistent escaping
		fmt.Fprintf(s, "%q", e.Error())
	}
}
