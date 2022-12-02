package sperr

type Severity rune

//go:generate go run golang.org/x/tools/cmd/stringer -type=Severity
const (
	Error   Severity = 'E'
	Warning Severity = 'W'
)
