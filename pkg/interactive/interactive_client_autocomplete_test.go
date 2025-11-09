package interactive

import (
	"testing"

	"github.com/c-bata/go-prompt"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/interactive/metaquery"
	"github.com/stretchr/testify/assert"
)

func TestGetFirstWordSuggestions(t *testing.T) {
	tests := map[string]struct {
		word              string
		setupSuggestions  func(*autoCompleteSuggestions)
		shouldContain     []string
		shouldNotContain  []string
		minSuggestionCount int
	}{
		"simple word returns all first word options": {
			word: "sel",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				// Default suggestions
			},
			shouldContain: []string{"select", "with"},
			minSuggestionCount: 2, // At least select, with, and metaqueries
		},
		"qualified mod query": {
			word: "mymod.",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.queriesByMod = map[string][]prompt.Suggest{
					"mymod": {
						{Text: "mymod.query1", Description: "Query"},
						{Text: "mymod.query2", Description: "Query"},
					},
				}
			},
			shouldContain: []string{"mymod.query1", "mymod.query2"},
		},
		"qualified unknown mod": {
			word: "unknown.",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.mods = []prompt.Suggest{
					{Text: "mod1", Description: "Mod"},
					{Text: "mod2", Description: "Mod"},
				}
				s.unqualifiedQueries = []prompt.Suggest{
					{Text: "query1", Description: "Query"},
				}
			},
			shouldContain: []string{"mod1", "mod2", "query1"},
		},
		"empty word returns all options": {
			word: "",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				// Default suggestions
			},
			shouldContain: []string{"select", "with"},
			minSuggestionCount: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := &InteractiveClient{
				suggestions: newAutocompleteSuggestions(),
			}
			tc.setupSuggestions(client.suggestions)

			suggestions := client.getFirstWordSuggestions(tc.word)

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

			for _, notExpected := range tc.shouldNotContain {
				for _, s := range suggestions {
					assert.NotEqual(t, notExpected, s.Text, "Should not contain: %s", notExpected)
				}
			}

			if tc.minSuggestionCount > 0 {
				assert.GreaterOrEqual(t, len(suggestions), tc.minSuggestionCount)
			}
		})
	}
}

func TestGetTableAndConnectionSuggestions(t *testing.T) {
	tests := map[string]struct {
		word             string
		setupSuggestions func(*autoCompleteSuggestions)
		expectedCount    int
		shouldContain    []string
	}{
		"unqualified returns schemas and unqualified tables": {
			word: "aws",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.schemas = []prompt.Suggest{
					{Text: "aws", Description: "Schema"},
					{Text: "gcp", Description: "Schema"},
				}
				s.unqualifiedTables = []prompt.Suggest{
					{Text: "aws_ec2_instance", Description: "Table"},
					{Text: "users", Description: "Table"},
				}
			},
			shouldContain: []string{"aws", "gcp", "aws_ec2_instance", "users"},
			expectedCount: 4,
		},
		"qualified returns tables from schema": {
			word: "aws.ec2",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.schemas = []prompt.Suggest{
					{Text: "aws", Description: "Schema"},
				}
				s.tablesBySchema = map[string][]prompt.Suggest{
					"aws": {
						{Text: "aws.ec2_instance", Description: "Table"},
						{Text: "aws.s3_bucket", Description: "Table"},
					},
				}
			},
			shouldContain: []string{"aws.ec2_instance", "aws.s3_bucket"},
			expectedCount: 2,
		},
		"qualified with unknown schema": {
			word: "unknown.table",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.schemas = []prompt.Suggest{
					{Text: "aws", Description: "Schema"},
				}
				s.unqualifiedTables = []prompt.Suggest{
					{Text: "users", Description: "Table"},
				}
			},
			expectedCount: 0,
		},
		"empty word": {
			word: "",
			setupSuggestions: func(s *autoCompleteSuggestions) {
				s.schemas = []prompt.Suggest{
					{Text: "aws", Description: "Schema"},
				}
				s.unqualifiedTables = []prompt.Suggest{
					{Text: "users", Description: "Table"},
				}
			},
			shouldContain: []string{"aws", "users"},
			expectedCount: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := &InteractiveClient{
				suggestions: newAutocompleteSuggestions(),
			}
			tc.setupSuggestions(client.suggestions)

			suggestions := client.getTableAndConnectionSuggestions(tc.word)

			assert.Len(t, suggestions, tc.expectedCount)

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

func TestQueryCompleter_EmptyInput(t *testing.T) {
	t.Run("returns nil for empty input when autocompleteOnEmpty is false", func(t *testing.T) {
		// Save original value
		originalValue := cmdconfig.Viper().GetBool(pconstants.ArgAutoComplete)
		defer cmdconfig.Viper().Set(pconstants.ArgAutoComplete, originalValue)

		// Enable autocomplete
		cmdconfig.Viper().Set(pconstants.ArgAutoComplete, true)

		client := &InteractiveClient{
			suggestions:            newAutocompleteSuggestions(),
			initialisationComplete: true,
			autocompleteOnEmpty:    false,
		}

		doc := prompt.Document{}
		doc.Text = ""

		suggestions := client.queryCompleter(doc)
		assert.Nil(t, suggestions)
	})

	t.Run("returns suggestions for empty input when autocompleteOnEmpty is true", func(t *testing.T) {
		// Save original value
		originalValue := cmdconfig.Viper().GetBool(pconstants.ArgAutoComplete)
		defer cmdconfig.Viper().Set(pconstants.ArgAutoComplete, originalValue)

		// Enable autocomplete
		cmdconfig.Viper().Set(pconstants.ArgAutoComplete, true)

		client := &InteractiveClient{
			suggestions:            newAutocompleteSuggestions(),
			initialisationComplete: true,
			autocompleteOnEmpty:    true,
		}

		doc := prompt.Document{}
		doc.Text = ""

		suggestions := client.queryCompleter(doc)
		// Should return some suggestions (select, with, metaqueries)
		assert.NotNil(t, suggestions)
	})
}

func TestQueryCompleter_FirstWord(t *testing.T) {
	t.Run("provides suggestions for first word", func(t *testing.T) {
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
		doc.Text = "sel"

		suggestions := client.queryCompleter(doc)
		assert.NotNil(t, suggestions)

		// Should contain select
		found := false
		for _, s := range suggestions {
			if s.Text == "select" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain 'select' suggestion")
	})
}

func TestQueryCompleter_Metaquery(t *testing.T) {
	t.Run("provides metaquery suggestions", func(t *testing.T) {
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
		doc.Text = ".he"

		suggestions := client.queryCompleter(doc)
		assert.NotNil(t, suggestions)

		// Should contain .help
		found := false
		for _, s := range suggestions {
			if s.Text == ".help" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain '.help' suggestion")
	})
}

func TestQueryCompleter_TableSuggestions(t *testing.T) {
	t.Run("provides table suggestions after FROM", func(t *testing.T) {
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
			{Text: "accounts", Description: "Table"},
		}
		client.suggestions.schemas = []prompt.Suggest{
			{Text: "public", Description: "Schema"},
		}

		doc := prompt.Document{}
		doc.Text = "select * from us"

		suggestions := client.queryCompleter(doc)
		assert.NotNil(t, suggestions)

		// Should contain users
		found := false
		for _, s := range suggestions {
			if s.Text == "users" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain 'users' suggestion")
	})
}

func TestInitialiseSchemaAndTableSuggestions(t *testing.T) {
	t.Run("handles nil schema metadata", func(t *testing.T) {
		client := &InteractiveClient{
			suggestions:    newAutocompleteSuggestions(),
			schemaMetadata: nil,
		}

		// Should not panic - the method returns early when schemaMetadata is nil
		assert.NotPanics(t, func() {
			client.initialiseSchemaAndTableSuggestions(nil)
		})

		// Suggestions should remain empty (or contain default connection tables)
		// Since the method returns early, schemas should be empty
		assert.Empty(t, client.suggestions.schemas)
	})

	t.Run("schema metadata structure", func(t *testing.T) {
		// Test that SchemaMetadata has the expected structure
		metadata := &db_common.SchemaMetadata{
			Schemas:             map[string]map[string]db_common.TableSchema{},
			TemporarySchemaName: "temp",
		}

		assert.NotNil(t, metadata.Schemas)
		assert.Equal(t, "temp", metadata.TemporarySchemaName)
	})
}

func TestIsMetaQuery(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"metaquery .help": {
			input:    ".help",
			expected: true,
		},
		"metaquery .tables": {
			input:    ".tables",
			expected: true,
		},
		"metaquery with space": {
			input:    ".inspect users",
			expected: true,
		},
		"not a metaquery": {
			input:    "select * from users",
			expected: false,
		},
		"empty string": {
			input:    "",
			expected: false,
		},
		"just dot": {
			input:    ".",
			expected: false, // A single dot is not a valid metaquery
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := metaquery.IsMetaQuery(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
