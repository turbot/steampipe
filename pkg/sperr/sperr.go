package sperr

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
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

func (e Error) Unwrap() error { return e.cause }
func (e Error) Is(target error) bool {
	_, ok := target.(*Error)
	return ok
}

func (e Error) As(target any) bool {
	if err, ok := target.(*Error); ok {
		target = err
		return true
	}
	return errors.As(e.RootCause(), target)
}

func (e Error) Error() string {
	return strings.Join(e.messageChain(), ":")
}

// TODO recurse down to all the details
func (e Error) Detail() string {
	return strings.Join(e.detailChain(), "\n|-")
}

// All error values returned from this package implement fmt.Formatter and can
// be formatted by the fmt package. The following verbs are supported:
//
//		%s    print the error. If the error has a Cause it will be
//		      printed recursively.
//		%v    see %s
//		%+v   extended format. Each Frame of the error's StackTrace will
//		      be printed in detail.
//	  %q		a double-quoted string safely escaped with Go syntax
//
// TODO: add Details for +
func (e Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, e.Detail())
			e.Stack().Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

func (e Error) messageChain() []string {
	res := []string{}
	if len(e.msg) > 0 {
		res = []string{e.msg}
	}
	if e.isRoot || e.cause == nil {
		return res
	}

	type hasMsgChain interface {
		messageChain() []string
	}

	// is the underlying a sperr
	if under, ok := e.cause.(hasMsgChain); ok {
		res = append(res, under.messageChain()...)
	} else {
		res = append(res, e.cause.Error())
	}
	return res
}

func (e Error) detailChain() []string {
	res := []string{}
	if len(e.detail) > 0 {
		res = []string{e.detail}
	}
	if e.cause == nil {
		return res
	}

	type hasDetailChain interface {
		detailChain() []string
	}
	// is the underlying a sperr
	if under, ok := e.cause.(hasDetailChain); ok {
		res = append(res, under.detailChain()...)
	} else {
		res = append(res, e.cause.Error())
	}
	return res
}
