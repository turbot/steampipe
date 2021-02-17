package constants

import "time"

// Display constants
const (
	// SpinnerShowTimeout :: duration after which spinner should be shown
	SpinnerShowTimeout = 1 * time.Second

	// Max Column Width
	MaxColumnWidth = 1024

	// what do we display for null column values
	NullString = "<null>"
)
