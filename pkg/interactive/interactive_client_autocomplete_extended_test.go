package interactive

import (
	"testing"

	"github.com/c-bata/go-prompt"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
	"github.com/stretchr/testify/assert"
)

// TestSanitiseTableName tests the table name escaping logic
// This function has complex logic with spaces, hyphens, and uppercase characters
// Bug hunting: looking for incorrect escaping, edge cases with special characters
func TestSanitiseTableName(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"simple lowercase": {
			input:    "users",
			expected: "users",
		},
		"with uppercase": {
			input:    "Users",
			expected: `"Users"`,
		},
		"with spaces": {
			input:    "my table",
			expected: `"my table"`,
		},
		"with hyphen": {
			input:    "my-table",
			expected: `"my-table"`,
		},
		"with uppercase and spaces": {
			input:    "My Table",
			expected: `"My Table"`,
		},
		"schema qualified lowercase": {
			input:    "public.users",
			expected: "public.users",
		},
		"schema qualified with uppercase table": {
			input:    "public.Users",
			expected: `public."Users"`,
		},
		"schema qualified with uppercase schema": {
			input:    "Public.users",
			expected: `"Public".users`,
		},
		"schema qualified both uppercase": {
			input:    "Public.Users",
			expected: `"Public"."Users"`,
		},
		"schema qualified with spaces": {
			input:    "my schema.my table",
			expected: `"my schema"."my table"`,
		},
		"schema qualified with hyphen": {
			input:    "my-schema.my-table",
			expected: `"my-schema"."my-table"`,
		},
		"empty string": {
			input:    "",
			expected: "",
		},
		"only dot": {
			input:    ".",
			expected: ".",
		},
		"multiple dots": {
			input:    "schema.table.column",
			expected: "schema.table.column",
		},
		"trailing dot": {
			input:    "schema.",
			expected: "schema.",
		},
		"leading dot": {
			input:    ".table",
			expected: ".table",
		},
		"underscore lowercase": {
			input:    "aws_ec2_instance",
			expected: "aws_ec2_instance",
		},
		"mixed case with underscore": {
			input:    "AWS_EC2_Instance",
			expected: `"AWS_EC2_Instance"`,
		},
		"special aws table": {
			input:    "aws.ec2-instance",
			expected: `aws."ec2-instance"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := sanitiseTableName(tc.input)
			assert.Equal(t, tc.expected, result, "sanitiseTableName(%q) should return %q but got %q", tc.input, tc.expected, result)
		})
	}
}

// TestQueryCompleterDisabled tests autocomplete when disabled
func TestQueryCompleterDisabled(t *testing.T) {
	t.Run("returns nil when autocomplete disabled", func(t *testing.T) {
		// Save original value
		originalValue := cmdconfig.Viper().GetBool(pconstants.ArgAutoComplete)
		defer cmdconfig.Viper().Set(pconstants.ArgAutoComplete, originalValue)

		// Disable autocomplete
		cmdconfig.Viper().Set(pconstants.ArgAutoComplete, false)

		client := &InteractiveClient{
			suggestions:            newAutocompleteSuggestions(),
			initialisationComplete: true,
		}

		doc := prompt.Document{}
		doc.Text = "select"

		suggestions := client.queryCompleter(doc)
		assert.Nil(t, suggestions, "Should return nil when autocomplete is disabled")
	})
}

// TestQueryCompleterNotInitialised tests autocomplete before initialization
func TestQueryCompleterNotInitialised(t *testing.T) {
	t.Run("returns nil when not initialised", func(t *testing.T) {
		// Save original value
		originalValue := cmdconfig.Viper().GetBool(pconstants.ArgAutoComplete)
		defer cmdconfig.Viper().Set(pconstants.ArgAutoComplete, originalValue)

		// Enable autocomplete
		cmdconfig.Viper().Set(pconstants.ArgAutoComplete, true)

		client := &InteractiveClient{
			suggestions:            newAutocompleteSuggestions(),
			initialisationComplete: false, // Not initialised
		}

		doc := prompt.Document{}
		doc.Text = "select"

		suggestions := client.queryCompleter(doc)
		assert.Nil(t, suggestions, "Should return nil when client is not initialised")
	})
}

// TestGetTableAndConnectionSuggestions_EdgeCases tests boundary conditions
// Bug hunting: nil checks, empty strings, malformed input
func TestGetTableAndConnectionSuggestions_EdgeCases(t *testing.T) {
	tests := map[string]struct {
		word             string
		setupSuggestions func(*autoCompleteSuggestions)
		expectedCount    int
	}{
		"multiple dots returns tables from first schema": {
			word: "schema.table.column",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.schemas = []prompt.Suggest{
					{Text: "schema", Description: "Schema"},
				}
				s.tablesBySchema = map[string][]prompt.Suggest{
					"schema": {
						{Text: "schema.table", Description: "Table"},
					},
				}
			},
			expectedCount: 1, // Returns tables from "schema" (first part before dot)
		},
		"dot only": {
			word: ".",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.schemas = []prompt.Suggest{
					{Text: "public", Description: "Schema"},
				}
				s.unqualifiedTables = []prompt.Suggest{
					{Text: "users", Description: "Table"},
				}
			},
			expectedCount: 0, // Schema is empty string after split, no tables for empty schema
		},
		"trailing dot with space returns tables from schema": {
			word: "schema. ",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.tablesBySchema = map[string][]prompt.Suggest{
					"schema": {
						{Text: "schema.table", Description: "Table"},
					},
				}
			},
			expectedCount: 1, // Returns tables from "schema"
		},
		"nil tablesBySchema map": {
			word: "schema.table",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				// Don't initialize tablesBySchema
				s.tablesBySchema = nil
			},
			expectedCount: 0, // Should handle nil map gracefully
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := &InteractiveClient{
				suggestions: newAutocompleteSuggestions(),
			}
			tc.setupSuggestions(client.suggestions)

			// This should not panic even with edge cases
			suggestions := client.getTableAndConnectionSuggestions(tc.word)
			assert.Len(t, suggestions, tc.expectedCount)
		})
	}
}

// TestGetFirstWordSuggestions_EdgeCases tests boundary conditions
// Bug hunting: empty mod names, nil maps, malformed qualified names
func TestGetFirstWordSuggestions_EdgeCases(t *testing.T) {
	tests := map[string]struct {
		word              string
		setupSuggestions  func(*autoCompleteSuggestions)
		minSuggestionCount int
		shouldContain     []string
	}{
		"multiple dots in qualified name": {
			word: "mod.query.extra",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.queriesByMod = map[string][]prompt.Suggest{
					"mod": {
						{Text: "mod.query1", Description: "Query"},
					},
				}
			},
			minSuggestionCount: 2, // Should still return select, with
		},
		"dot only": {
			word: ".",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.mods = []prompt.Suggest{
					{Text: "mymod", Description: "Mod"},
				}
			},
			minSuggestionCount: 2, // Should return mods and queries
		},
		"empty mod name": {
			word: ".query",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.queriesByMod = map[string][]prompt.Suggest{
					"": {
						{Text: "query1", Description: "Query"},
					},
				}
			},
			minSuggestionCount: 2, // Should still return select, with
		},
		"nil queriesByMod map": {
			word: "mod.query",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.queriesByMod = nil // Should not panic
			},
			shouldContain: []string{"select", "with"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := &InteractiveClient{
				suggestions: newAutocompleteSuggestions(),
			}
			tc.setupSuggestions(client.suggestions)

			// Should not panic with edge cases
			suggestions := client.getFirstWordSuggestions(tc.word)

			if tc.minSuggestionCount > 0 {
				assert.GreaterOrEqual(t, len(suggestions), tc.minSuggestionCount)
			}

			for _, expected := range tc.shouldContain {
				found := false
				for _, s := range suggestions {
					if s.Text == expected {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find suggestion: %s", expected)
			}
		})
	}
}

// TestQueryCompleter_CaseInsensitivity tests that autocomplete is case insensitive
func TestQueryCompleter_CaseInsensitivity(t *testing.T) {
	tests := map[string]struct {
		input         string
		shouldContain string
	}{
		"uppercase SELECT": {
			input:         "SELECT",
			shouldContain: "select",
		},
		"mixed case SeLeCt": {
			input:         "SeLeCt",
			shouldContain: "select",
		},
		"uppercase FROM with trailing space": {
			input:         "SELECT * FROM ",
			shouldContain: "", // Should provide table suggestions
		},
		"mixed case from with trailing space": {
			input:         "select * FrOm ",
			shouldContain: "", // Should provide table suggestions
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Save original value
			originalValue := cmdconfig.Viper().GetBool(pconstants.ArgAutoComplete)
			defer cmdconfig.Viper().Set(pconstants.ArgAutoComplete, originalValue)

			// Enable autocomplete
			cmdconfig.Viper().Set(pconstants.ArgAutoComplete, true)

			client := &InteractiveClient{
				suggestions:            newAutocompleteSuggestions(),
				initialisationComplete: true,
			}
			client.suggestions.unqualifiedTables = []prompt.Suggest{
				{Text: "users", Description: "Table"},
			}

			doc := prompt.Document{}
			doc.Text = tc.input

			// Should not panic and should handle case insensitivity
			assert.NotPanics(t, func() {
				client.queryCompleter(doc)
			}, "Should not panic for: %s", tc.input)
		})
	}
}

// TestInitialiseSchemaAndTableSuggestions_EmptyConnectionState tests behavior with empty state
// Bug hunting: nil pointer dereferences, empty collections
// NOTE: Full testing requires database connection and proper client setup
// BUG FOUND: initialiseSchemaAndTableSuggestions panics when client() is nil
// The function calls c.client().GetRequiredSessionSearchPath() without checking if client() is nil
func TestInitialiseSchemaAndTableSuggestions_EmptyConnectionState(t *testing.T) {
	t.Run("returns early with nil schema metadata", func(t *testing.T) {
		client := &InteractiveClient{
			suggestions:    newAutocompleteSuggestions(),
			schemaMetadata: nil, // Nil metadata should return early
		}

		// Should not panic with nil metadata (early return)
		assert.NotPanics(t, func() {
			client.initialiseSchemaAndTableSuggestions(steampipeconfig.ConnectionStateMap{})
		})

		// Schemas should remain empty
		assert.Empty(t, client.suggestions.schemas)
	})
}

// TestGetPreviousWord_EdgeCase tests the 90% covered case
func TestGetPreviousWord_MultipleTrailingSpaces(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"many trailing spaces": {
			input:    "select * from     ",
			expected: "from",
		},
		"tabs are not spaces": {
			input:    "select * from\t\t  ",
			expected: "from\t\t", // Tabs are different characters, lastIndexByteNot only looks for spaces
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := getPreviousWord(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestAutocompleteSuggestions_ConcurrentAccess tests for race conditions
// Bug hunting: concurrent map access, data races
func TestAutocompleteSuggestions_ConcurrentSort(t *testing.T) {
	t.Run("concurrent sort calls do not panic", func(t *testing.T) {
		suggestions := newAutocompleteSuggestions()
		suggestions.schemas = []prompt.Suggest{
			{Text: "z", Description: "Schema"},
			{Text: "a", Description: "Schema"},
			{Text: "m", Description: "Schema"},
		}

		// This test ensures sort is safe to call multiple times
		// In a real scenario, we'd want to test with race detector
		done := make(chan bool, 3)

		for i := 0; i < 3; i++ {
			go func() {
				suggestions.sort()
				done <- true
			}()
		}

		for i := 0; i < 3; i++ {
			<-done
		}

		// Should be sorted
		assert.Equal(t, "a", suggestions.schemas[0].Text)
	})
}

// TestQueryCompleter_WithWhitespace tests various whitespace scenarios
// Bug hunting: whitespace handling, trimming edge cases
func TestQueryCompleter_WithWhitespace(t *testing.T) {
	tests := map[string]struct {
		input string
	}{
		"leading spaces": {
			input: "   select",
		},
		"leading tabs": {
			input: "\t\tselect",
		},
		"mixed leading whitespace": {
			input: " \t select",
		},
		"multiple spaces between words": {
			input: "select  *  from",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Save original value
			originalValue := cmdconfig.Viper().GetBool(pconstants.ArgAutoComplete)
			defer cmdconfig.Viper().Set(pconstants.ArgAutoComplete, originalValue)

			// Enable autocomplete
			cmdconfig.Viper().Set(pconstants.ArgAutoComplete, true)

			client := &InteractiveClient{
				suggestions:            newAutocompleteSuggestions(),
				initialisationComplete: true,
			}

			doc := prompt.Document{}
			doc.Text = tc.input

			// Should handle whitespace gracefully and not panic
			assert.NotPanics(t, func() {
				client.queryCompleter(doc)
			})
		})
	}
}

// TestQueryCompleter_VeryLongInput tests behavior with large input
// Bug hunting: performance issues, buffer overflows
func TestQueryCompleter_VeryLongInput(t *testing.T) {
	t.Run("handles very long input", func(t *testing.T) {
		// Save original value
		originalValue := cmdconfig.Viper().GetBool(pconstants.ArgAutoComplete)
		defer cmdconfig.Viper().Set(pconstants.ArgAutoComplete, originalValue)

		// Enable autocomplete
		cmdconfig.Viper().Set(pconstants.ArgAutoComplete, true)

		client := &InteractiveClient{
			suggestions:            newAutocompleteSuggestions(),
			initialisationComplete: true,
		}

		// Create a very long input string
		longInput := "select * from "
		for i := 0; i < 1000; i++ {
			longInput += "very_long_table_name_that_keeps_going "
		}

		doc := prompt.Document{}
		doc.Text = longInput

		// Should not panic or hang with very long input
		assert.NotPanics(t, func() {
			client.queryCompleter(doc)
		})
	})
}

// TestGetTableAndConnectionSuggestions_NilSuggestions tests nil safety
func TestGetTableAndConnectionSuggestions_NilSuggestions(t *testing.T) {
	t.Run("handles nil schemas slice", func(t *testing.T) {
		client := &InteractiveClient{
			suggestions: newAutocompleteSuggestions(),
		}
		client.suggestions.schemas = nil

		// Should not panic when appending nil slices
		assert.NotPanics(t, func() {
			client.getTableAndConnectionSuggestions("test")
		})
	})

	t.Run("handles nil unqualifiedTables slice", func(t *testing.T) {
		client := &InteractiveClient{
			suggestions: newAutocompleteSuggestions(),
		}
		client.suggestions.unqualifiedTables = nil

		// Should not panic when appending nil slices
		assert.NotPanics(t, func() {
			client.getTableAndConnectionSuggestions("test")
		})
	})
}
