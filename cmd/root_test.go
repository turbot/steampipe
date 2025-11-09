package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHideRootFlags_NonExistentFlag tests that hideRootFlags handles non-existent flags gracefully
// Bug #4707: hideRootFlags panics when called with a flag that doesn't exist
func TestHideRootFlags_NonExistentFlag(t *testing.T) {
	// Initialize the root command
	InitCmd()

	// Test that calling hideRootFlags with a non-existent flag should NOT panic
	assert.NotPanics(t, func() {
		hideRootFlags("non-existent-flag")
	}, "hideRootFlags should handle non-existent flags without panicking")
}
