package sperr

type SteampipeErrorOption func(*SteampipeError)

func WithSeverity(severity Severity) SteampipeErrorOption {
	return func(se *SteampipeError) {
		se.severity = severity
	}
}

func WithDiagnostic(message string) SteampipeErrorOption {
	return func(se *SteampipeError) {
		se.diagnosticMessage = message
	}
}

func WithUserMessage(message string) SteampipeErrorOption {
	return func(se *SteampipeError) {
		se.userMessage = message
	}
}

func WithUserMessageReplaced(message string) SteampipeErrorOption {
	return func(se *SteampipeError) {
		se.userMessage = message
		se.replaceUserMessage = true
	}
}
