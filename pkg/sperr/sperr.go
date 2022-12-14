package sperr

import (
	"fmt"
	"io"
)

type Error struct {
	stack           *stack
	cause           error
	diagnostic      string
	msg             string
	hideChildErrors bool
}

// Cause will recursively retrieve
// the topmost error that does not have a cause, which is assumed to be
// the original cause.
func (e *Error) Cause() error {
	type hasCause interface {
		Cause() error
	}
	if cause, ok := e.cause.(hasCause); ok {
		return cause.Cause()
	}
	return e.cause

}

func (e *Error) Error() string {
	if e.cause != nil && !e.hideChildErrors {
		return e.msg + ": " + e.cause.Error()
	}
	return e.msg
}

// implement the fmt.Formatter interface
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, e.msg)
			e.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.msg)
	case 'q':
		fmt.Fprintf(s, "%q", e.msg)
	}
}
