package interactive

import (
	"fmt"
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

// TestAutocompleteSuggestionsMemoryUsage tests that autocomplete suggestion maps
// have reasonable size limits to prevent unbounded memory growth.
//
// Bug: #4812 - Without limits, large schemas cause excessive memory consumption
func TestAutocompleteSuggestionsMemoryUsage(t *testing.T) {
	// Create a client with suggestions
	c := &InteractiveClient{
		suggestions: newAutocompleteSuggestions(),
	}

	// Simulate a large schema with many tables
	// This represents a real-world scenario with cloud providers having 100+ connections
	// each with 50+ tables
	const numSchemas = 150
	const tablesPerSchema = 75

	for schema := 0; schema < numSchemas; schema++ {
		schemaName := fmt.Sprintf("connection_%d", schema)
		tables := make([]prompt.Suggest, tablesPerSchema)
		for i := 0; i < tablesPerSchema; i++ {
			tables[i] = prompt.Suggest{
				Text:        fmt.Sprintf("table_%d", i),
				Description: "Description for table",
			}
		}
		c.suggestions.tablesBySchema[schemaName] = tables
	}

	// The fix should enforce reasonable limits on map sizes
	// Max 100 schemas in the suggestions map
	const maxSchemas = 100
	if len(c.suggestions.tablesBySchema) > maxSchemas {
		t.Errorf("tablesBySchema map should be limited to %d schemas, got %d",
			maxSchemas, len(c.suggestions.tablesBySchema))
	}

	// Max 500 tables per schema
	const maxTablesPerSchema = 500
	for schema, tables := range c.suggestions.tablesBySchema {
		if len(tables) > maxTablesPerSchema {
			t.Errorf("tablesBySchema[%s] should be limited to %d tables, got %d",
				schema, maxTablesPerSchema, len(tables))
		}
	}
}
