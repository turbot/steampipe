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

// TestGetQueryInfo_FromDetection tests that getQueryInfo correctly detects
// when the user is editing a table name after typing "from ".
//
// This is important for autocomplete - when a user types "from " (with a space),
// the system should recognize they are about to enter a table name and enable
// table suggestions.
//
// Bug: #4810
func TestGetQueryInfo_FromDetection(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedTable     string
		expectedEditTable bool
	}{
		{
			name:              "just_from",
			input:             "from ",
			expectedTable:     "",
			expectedEditTable: true, // Should be true - user is about to enter table name
		},
		{
			name:              "from_with_table",
			input:             "from my_table",
			expectedTable:     "my_table",
			expectedEditTable: false, // Not editing, already entered
		},
		{
			name:              "from_keyword_only",
			input:             "from",
			expectedTable:     "",
			expectedEditTable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getQueryInfo(tt.input)

			if result.Table != tt.expectedTable {
				t.Errorf("getQueryInfo(%q).Table = %q, expected %q", tt.input, result.Table, tt.expectedTable)
			}

			if result.EditingTable != tt.expectedEditTable {
				t.Errorf("getQueryInfo(%q).EditingTable = %v, expected %v", tt.input, result.EditingTable, tt.expectedEditTable)
			}
		})
	}
}
