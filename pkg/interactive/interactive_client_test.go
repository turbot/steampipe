package interactive

import (
	"context"
	"sync"
	"testing"

	"github.com/c-bata/go-prompt"
)

// TestGetTableAndConnectionSuggestions_ReturnsEmptySliceNotNil tests that
// getTableAndConnectionSuggestions returns an empty slice instead of nil
// when no matching connection is found in the schema.
//
// This is important for proper API contract - functions that return slices
// should return empty slices rather than nil to avoid unexpected nil pointer
// issues in calling code.
//
// Bug: #4710
// PR: #4734
func TestGetTableAndConnectionSuggestions_ReturnsEmptySliceNotNil(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		expected bool // true if we expect non-nil result
	}{
		{
			name:     "empty word should return non-nil",
			word:     "",
			expected: true,
		},
		{
			name:     "unqualified table should return non-nil",
			word:     "table",
			expected: true,
		},
		{
			name:     "non-existent connection should return non-nil",
			word:     "nonexistent.table",
			expected: true,
		},
		{
			name:     "qualified table with dot should return non-nil",
			word:     "aws.instances",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal InteractiveClient with empty suggestions
			c := &InteractiveClient{
				suggestions: &autoCompleteSuggestions{
					schemas:          []prompt.Suggest{},
					unqualifiedTables: []prompt.Suggest{},
					tablesBySchema:    make(map[string][]prompt.Suggest),
				},
			}

			result := c.getTableAndConnectionSuggestions(tt.word)

			if tt.expected && result == nil {
				t.Errorf("getTableAndConnectionSuggestions(%q) returned nil, expected non-nil empty slice", tt.word)
			}

			// Additional check: even if not nil, should be empty in these test cases
			if result != nil && len(result) != 0 {
				t.Errorf("getTableAndConnectionSuggestions(%q) returned non-empty slice %v, expected empty slice", tt.word, result)
			}
		})
	}
}

// TestInitialisationComplete_RaceCondition tests that concurrent access to
// the initialisationComplete flag does not cause data races.
//
// This test simulates the real-world scenario where:
// - One goroutine (init goroutine) writes to initialisationComplete
// - Other goroutines (query executor, notification handler) read from it
//
// Bug: #4803
func TestInitialisationComplete_RaceCondition(t *testing.T) {
	c := &InteractiveClient{
		initialisationComplete: false,
	}

	var wg sync.WaitGroup
	ctx := context.Background()

	// Simulate initialization goroutine writing to the flag
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			c.initialisationComplete = true
			c.initialisationComplete = false
		}
	}()

	// Simulate query executor reading the flag
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = c.isInitialised()
		}
	}()

	// Simulate notification handler reading the flag
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			c.handleConnectionUpdateNotification(ctx)
		}
	}()

	wg.Wait()
}
