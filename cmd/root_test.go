package cmd

import (
	"sync"
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

// TestAddCommands_Concurrent tests that AddCommands is thread-safe
// Bug #4708: AddCommands/ResetCommands not thread-safe (data races detected)
func TestAddCommands_Concurrent(t *testing.T) {
	var wg sync.WaitGroup

	// Run AddCommands concurrently to expose race conditions
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ResetCommands()
			AddCommands()
		}()
	}

	wg.Wait()
}
