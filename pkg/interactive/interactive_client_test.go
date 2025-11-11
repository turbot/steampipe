package interactive

import (
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

// TestAutocompleteSuggestionsNilMaps tests that sort() handles nil maps gracefully.
//
// When autoCompleteSuggestions is initialized without using the constructor
// (e.g., through struct literal or improper initialization), the tablesBySchema
// and queriesByMod maps will be nil. The sort() method should handle this
// gracefully with explicit nil checks for defensive programming.
//
// Issue: #4813
func TestAutocompleteSuggestionsNilMaps(t *testing.T) {
	// Initialize struct without using constructor - maps will be nil
	suggestions := &autoCompleteSuggestions{
		schemas:            []prompt.Suggest{{Text: "zebra"}, {Text: "apple"}},
		unqualifiedTables:  []prompt.Suggest{{Text: "users"}, {Text: "accounts"}},
		unqualifiedQueries: []prompt.Suggest{{Text: "query2"}, {Text: "query1"}},
		// tablesBySchema and queriesByMod are nil (not initialized)
	}

	// This should not panic even with nil maps
	// The sort() method should handle nil maps gracefully
	suggestions.sort()

	// Verify the slices were sorted correctly
	if len(suggestions.schemas) != 2 || suggestions.schemas[0].Text != "apple" {
		t.Errorf("schemas not sorted correctly, got %v", suggestions.schemas)
	}
	if len(suggestions.unqualifiedTables) != 2 || suggestions.unqualifiedTables[0].Text != "accounts" {
		t.Errorf("unqualifiedTables not sorted correctly, got %v", suggestions.unqualifiedTables)
	}
}
