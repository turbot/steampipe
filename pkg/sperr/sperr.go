package sperr

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
	se := new(SteampipeError)
	se.origErr = err
	for _, o := range options {
		o(se)
	}
	return se
}
