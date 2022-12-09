package sperr

type SteampipeErrorOption func(*SteampipeError)

func WithSeverity(severity Severity) SteampipeErrorOption {
	return func(se *SteampipeError) {
		se.severity = severity
	}
}

func WithDiagnostic(message string) SteampipeErrorOption {
	return func(se *SteampipeError) {
		se.diagnostic = message
	}
}

func WithMessage(message string) SteampipeErrorOption {
	return func(se *SteampipeError) {
		se.message = message
	}
}

func AsTerminal() SteampipeErrorOption {
	return func(se *SteampipeError) {
		se.terminate = true
	}
}
