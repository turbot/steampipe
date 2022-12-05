package sperr

import "fmt"

type SteampipeError struct {
	// make this compatible with the golang error interface
	error

	severity           Severity
	diagnosticMessage  string
	userMessage        string
	replaceUserMessage bool
	origErr            error
}

func NewSteampipeError(err error, options ...SteampipeErrorOption) *SteampipeError {
	if err == nil {
		return nil
	}
	se := &SteampipeError{
		origErr:  err,
		severity: Warning,
	}
	for _, o := range options {
		o(se)
	}
	return se
}

func (e *SteampipeError) Error() string {
	return fmt.Sprintf("%s - %s", e.diagnosticMessage, e.origErr.Error())
}
