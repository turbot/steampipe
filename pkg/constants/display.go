package constants

import "time"

// Display constants
const (
	// SpinnerShowTimeout is the duration after which spinner should be shown
	SpinnerShowTimeout = 1 * time.Second

	MaxColumnWidth = 1024

	// NullString is the string which is displayed for null column values
	NullString = "<null>"
)
