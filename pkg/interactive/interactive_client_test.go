package interactive

import (
	"testing"

	"github.com/c-bata/go-prompt"
)

// TestGetTableAndConnectionSuggestionsEdgeCases tests that
// getTableAndConnectionSuggestions returns an empty slice instead of nil
// when the schema is not found in the tablesBySchema map.
//
// This is important for proper API contract - functions that return slices
// should return empty slices rather than nil to avoid unexpected nil pointer
// issues in calling code.
//
// Bug: #4800
func TestGetTableAndConnectionSuggestionsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		expected bool // true if we expect non-nil result
	}{
		{
			name:     "unknown schema should return non-nil empty slice",
			word:     "unknown_schema.table_name",
			expected: true,
		},
		{
			name:     "qualified name with unknown connection",
			word:     "nonexistent.some_table",
			expected: true,
		},
		{
			name:     "partial qualified name",
			word:     "missing.",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal InteractiveClient with empty suggestions
			c := &InteractiveClient{
				suggestions: &autoCompleteSuggestions{
					schemas:           []prompt.Suggest{},
					unqualifiedTables: []prompt.Suggest{},
					tablesBySchema:    make(map[string][]prompt.Suggest),
				},
			}

			result := c.getTableAndConnectionSuggestions(tt.word)

			if tt.expected && result == nil {
				t.Errorf("getTableAndConnectionSuggestions(%q) returned nil, expected non-nil empty slice", tt.word)
			}

			// Additional check: should be empty in these test cases
			if result != nil && len(result) != 0 {
				t.Errorf("getTableAndConnectionSuggestions(%q) returned non-empty slice %v, expected empty slice", tt.word, result)
			}
		})
	}
}
