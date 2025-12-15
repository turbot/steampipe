package interactive

import (
	"testing"

	"github.com/c-bata/go-prompt"
)

// TestNewAutocompleteSuggestions tests the creation of autocomplete suggestions
func TestNewAutocompleteSuggestions(t *testing.T) {
	s := newAutocompleteSuggestions()

	if s == nil {
		t.Fatal("newAutocompleteSuggestions returned nil")
	}

	if s.tablesBySchema == nil {
		t.Error("tablesBySchema map is nil")
	}

	if s.queriesByMod == nil {
		t.Error("queriesByMod map is nil")
	}

	// Note: slices are not initialized (nil is valid for slices in Go)
	// We just verify the struct itself is created
}

// TestAutocompleteSuggestionsSort tests the sorting of suggestions
func TestAutocompleteSuggestionsSort(t *testing.T) {
	s := newAutocompleteSuggestions()

	// Add unsorted suggestions
	s.schemas = []prompt.Suggest{
		{Text: "zebra", Description: "Schema"},
		{Text: "apple", Description: "Schema"},
		{Text: "mango", Description: "Schema"},
	}

	s.unqualifiedTables = []prompt.Suggest{
		{Text: "users", Description: "Table"},
		{Text: "accounts", Description: "Table"},
		{Text: "posts", Description: "Table"},
	}

	s.tablesBySchema["test"] = []prompt.Suggest{
		{Text: "z_table", Description: "Table"},
		{Text: "a_table", Description: "Table"},
	}

	// Sort
	s.sort()

	// Verify schemas are sorted
	if len(s.schemas) > 1 {
		for i := 1; i < len(s.schemas); i++ {
			if s.schemas[i-1].Text > s.schemas[i].Text {
				t.Errorf("schemas not sorted: %s > %s", s.schemas[i-1].Text, s.schemas[i].Text)
			}
		}
	}

	// Verify tables are sorted
	if len(s.unqualifiedTables) > 1 {
		for i := 1; i < len(s.unqualifiedTables); i++ {
			if s.unqualifiedTables[i-1].Text > s.unqualifiedTables[i].Text {
				t.Errorf("unqualifiedTables not sorted: %s > %s", s.unqualifiedTables[i-1].Text, s.unqualifiedTables[i].Text)
			}
		}
	}

	// Verify tablesBySchema are sorted
	tables := s.tablesBySchema["test"]
	if len(tables) > 1 {
		for i := 1; i < len(tables); i++ {
			if tables[i-1].Text > tables[i].Text {
				t.Errorf("tablesBySchema not sorted: %s > %s", tables[i-1].Text, tables[i].Text)
			}
		}
	}
}

// TestAutocompleteSuggestionsEmptySort tests sorting with empty suggestions
func TestAutocompleteSuggestionsEmptySort(t *testing.T) {
	s := newAutocompleteSuggestions()

	// Should not panic with empty suggestions
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("sort() panicked with empty suggestions: %v", r)
		}
	}()

	s.sort()
}

// TestAutocompleteSuggestionsSortWithDuplicates tests sorting with duplicate entries
func TestAutocompleteSuggestionsSortWithDuplicates(t *testing.T) {
	s := newAutocompleteSuggestions()

	// Add duplicate suggestions
	s.schemas = []prompt.Suggest{
		{Text: "apple", Description: "Schema"},
		{Text: "apple", Description: "Schema"},
		{Text: "banana", Description: "Schema"},
	}

	// Should not panic with duplicates
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("sort() panicked with duplicates: %v", r)
		}
	}()

	s.sort()

	// Verify duplicates are preserved (not removed)
	if len(s.schemas) != 3 {
		t.Errorf("sort() removed duplicates, got %d entries, want 3", len(s.schemas))
	}
}

// TestAutocompleteSuggestionsWithUnicode tests suggestions with unicode characters
func TestAutocompleteSuggestionsWithUnicode(t *testing.T) {
	s := newAutocompleteSuggestions()

	s.schemas = []prompt.Suggest{
		{Text: "用户", Description: "Schema"},
		{Text: "数据库", Description: "Schema"},
		{Text: "🔥", Description: "Schema"},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("sort() panicked with unicode: %v", r)
		}
	}()

	s.sort()

	// Just verify it doesn't crash
	if len(s.schemas) != 3 {
		t.Errorf("sort() lost unicode entries, got %d entries, want 3", len(s.schemas))
	}
}

// TestAutocompleteSuggestionsLargeDataset tests with a large number of suggestions
func TestAutocompleteSuggestionsLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	s := newAutocompleteSuggestions()

	// Add 10,000 schemas
	for i := 0; i < 10000; i++ {
		s.schemas = append(s.schemas, prompt.Suggest{
			Text:        "schema_" + string(rune(i)),
			Description: "Schema",
		})
	}

	// Add 10,000 tables
	for i := 0; i < 10000; i++ {
		s.unqualifiedTables = append(s.unqualifiedTables, prompt.Suggest{
			Text:        "table_" + string(rune(i)),
			Description: "Table",
		})
	}

	// Should not hang or crash
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("sort() panicked with large dataset: %v", r)
		}
	}()

	s.sort()
}

// TestAutocompleteSuggestionsMemoryUsage tests memory usage with many suggestions
func TestAutocompleteSuggestionsMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	// Create 100 suggestion sets
	suggestions := make([]*autoCompleteSuggestions, 100)

	for i := 0; i < 100; i++ {
		s := newAutocompleteSuggestions()

		// Add many suggestions
		for j := 0; j < 1000; j++ {
			s.schemas = append(s.schemas, prompt.Suggest{
				Text:        "schema",
				Description: "Schema",
			})
		}

		suggestions[i] = s
	}

	// If we get here without OOM, the test passes
	// Clear suggestions to allow GC
	suggestions = nil
}

// TestAutocompleteSuggestionsSizeLimits tests that suggestion maps are bounded
// This test verifies the fix for #4812: autocomplete suggestions should have size limits
func TestAutocompleteSuggestionsSizeLimits(t *testing.T) {
	s := newAutocompleteSuggestions()

	// Test setTablesForSchema enforces schema count limit
	t.Run("schema count limit", func(t *testing.T) {
		// Add more schemas than the limit
		for i := 0; i < 150; i++ {
			tables := []prompt.Suggest{
				{Text: "table1", Description: "Table"},
			}
			s.setTablesForSchema("schema_"+string(rune(i)), tables)
		}

		// Should not exceed maxSchemasInSuggestions (100)
		if len(s.tablesBySchema) > 100 {
			t.Errorf("tablesBySchema size %d exceeds limit of 100", len(s.tablesBySchema))
		}
	})

	// Test setTablesForSchema enforces per-schema table limit
	t.Run("tables per schema limit", func(t *testing.T) {
		s2 := newAutocompleteSuggestions()

		// Create more tables than the limit
		manyTables := make([]prompt.Suggest, 600)
		for i := 0; i < 600; i++ {
			manyTables[i] = prompt.Suggest{
				Text:        "table_" + string(rune(i)),
				Description: "Table",
			}
		}

		s2.setTablesForSchema("test_schema", manyTables)

		// Should not exceed maxTablesPerSchema (500)
		if len(s2.tablesBySchema["test_schema"]) > 500 {
			t.Errorf("tables per schema %d exceeds limit of 500", len(s2.tablesBySchema["test_schema"]))
		}
	})

	// Test setQueriesForMod enforces mod count limit
	t.Run("mod count limit", func(t *testing.T) {
		s3 := newAutocompleteSuggestions()

		// Add more mods than the limit
		for i := 0; i < 150; i++ {
			queries := []prompt.Suggest{
				{Text: "query1", Description: "Query"},
			}
			s3.setQueriesForMod("mod_"+string(rune(i)), queries)
		}

		// Should not exceed maxSchemasInSuggestions (100)
		if len(s3.queriesByMod) > 100 {
			t.Errorf("queriesByMod size %d exceeds limit of 100", len(s3.queriesByMod))
		}
	})

	// Test setQueriesForMod enforces per-mod query limit
	t.Run("queries per mod limit", func(t *testing.T) {
		s4 := newAutocompleteSuggestions()

		// Create more queries than the limit
		manyQueries := make([]prompt.Suggest, 600)
		for i := 0; i < 600; i++ {
			manyQueries[i] = prompt.Suggest{
				Text:        "query_" + string(rune(i)),
				Description: "Query",
			}
		}

		s4.setQueriesForMod("test_mod", manyQueries)

		// Should not exceed maxQueriesPerMod (500)
		if len(s4.queriesByMod["test_mod"]) > 500 {
			t.Errorf("queries per mod %d exceeds limit of 500", len(s4.queriesByMod["test_mod"]))
		}
	})
}

// TestAutocompleteSuggestionsEdgeCases tests various edge cases
func TestAutocompleteSuggestionsEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "empty text suggestion",
			test: func(t *testing.T) {
				s := newAutocompleteSuggestions()
				s.schemas = []prompt.Suggest{
					{Text: "", Description: "Empty"},
				}
				s.sort() // Should not panic
			},
		},
		{
			name: "very long text suggestion",
			test: func(t *testing.T) {
				s := newAutocompleteSuggestions()
				longText := make([]byte, 10000)
				for i := range longText {
					longText[i] = 'a'
				}
				s.schemas = []prompt.Suggest{
					{Text: string(longText), Description: "Long"},
				}
				s.sort() // Should not panic
			},
		},
		{
			name: "null bytes in text",
			test: func(t *testing.T) {
				s := newAutocompleteSuggestions()
				s.schemas = []prompt.Suggest{
					{Text: "schema\x00name", Description: "Null"},
				}
				s.sort() // Should not panic
			},
		},
		{
			name: "special characters in text",
			test: func(t *testing.T) {
				s := newAutocompleteSuggestions()
				s.schemas = []prompt.Suggest{
					{Text: "schema!@#$%^&*()", Description: "Special"},
				}
				s.sort() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Test panicked: %v", r)
				}
			}()
			tt.test(t)
		})
	}
}
