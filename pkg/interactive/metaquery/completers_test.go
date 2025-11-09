package metaquery

import (
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/stretchr/testify/assert"
)

func TestComplete(t *testing.T) {
	tableSuggestions := []prompt.Suggest{
		{Text: "aws_ec2_instance", Description: "Table"},
		{Text: "aws_s3_bucket", Description: "Table"},
	}

	tests := map[string]struct {
		query               string
		tableSuggestions    []prompt.Suggest
		expectedCount       int
		shouldContain       []string
		shouldNotContain    []string
		checkExactSuggestion bool
	}{
		".inspect with table suggestions": {
			query:            ".inspect ",
			tableSuggestions: tableSuggestions,
			expectedCount:    2,
			shouldContain:    []string{"aws_ec2_instance", "aws_s3_bucket"},
		},
		".inspect without arguments": {
			query:            ".inspect",
			tableSuggestions: tableSuggestions,
			expectedCount:    2,
		},
		".header with on/off": {
			query:         ".header",
			expectedCount: 2,
			shouldContain: []string{"on", "off"},
		},
		".multi with on/off": {
			query:         ".multi",
			expectedCount: 2,
			shouldContain: []string{"on", "off"},
		},
		".timing with options": {
			query:         ".timing",
			expectedCount: 3,
			shouldContain: []string{"on", "off", "verbose"},
		},
		".output with formats": {
			query:         ".output",
			expectedCount: 4,
			shouldContain: []string{"json", "csv", "table", "line"},
		},
		".cache with options": {
			query:         ".cache",
			expectedCount: 3,
			shouldContain: []string{"on", "off", "clear"},
		},
		".autocomplete with on/off": {
			query:         ".autocomplete",
			expectedCount: 2,
			shouldContain: []string{"on", "off"},
		},
		"unknown command": {
			query:         ".unknown",
			expectedCount: 0,
		},
		".help has no completer": {
			query:         ".help",
			expectedCount: 0,
		},
		".exit has no completer": {
			query:         ".exit",
			expectedCount: 0,
		},
		"query with semicolon stripped": {
			query:            ".inspect aws;",
			tableSuggestions: tableSuggestions,
			expectedCount:    2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			input := &CompleterInput{
				Query:            tc.query,
				TableSuggestions: tc.tableSuggestions,
			}

			suggestions := Complete(input)

			if tc.expectedCount > 0 {
				assert.Len(t, suggestions, tc.expectedCount)
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

			for _, notExpected := range tc.shouldNotContain {
				for _, s := range suggestions {
					assert.NotEqual(t, notExpected, s.Text, "Should not contain suggestion: %s", notExpected)
				}
			}
		})
	}
}

func TestCompleterFromArgsOf(t *testing.T) {
	tests := map[string]struct {
		cmd           string
		expectedCount int
		shouldContain []string
	}{
		".header completer": {
			cmd:           ".header",
			expectedCount: 2,
			shouldContain: []string{"on", "off"},
		},
		".multi completer": {
			cmd:           ".multi",
			expectedCount: 2,
			shouldContain: []string{"on", "off"},
		},
		".timing completer": {
			cmd:           ".timing",
			expectedCount: 3,
			shouldContain: []string{"on", "off", "verbose"},
		},
		".output completer": {
			cmd:           ".output",
			expectedCount: 4,
			shouldContain: []string{"json", "csv", "table", "line"},
		},
		".cache completer": {
			cmd:           ".cache",
			expectedCount: 3,
			shouldContain: []string{"on", "off", "clear"},
		},
		".autocomplete completer": {
			cmd:           ".autocomplete",
			expectedCount: 2,
			shouldContain: []string{"on", "off"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			completer := completerFromArgsOf(tc.cmd)
			input := &CompleterInput{Query: tc.cmd}

			suggestions := completer(input)

			assert.Len(t, suggestions, tc.expectedCount)

			for _, expected := range tc.shouldContain {
				found := false
				for _, s := range suggestions {
					if s.Text == expected {
						found = true
						assert.NotEmpty(t, s.Description, "Suggestion should have description")
						assert.Equal(t, expected, s.Output)
						break
					}
				}
				assert.True(t, found, "Expected to find suggestion: %s", expected)
			}
		})
	}
}

func TestInspectCompleter(t *testing.T) {
	tests := map[string]struct {
		tableSuggestions []prompt.Suggest
		expectedCount    int
		shouldContain    []string
	}{
		"with table suggestions": {
			tableSuggestions: []prompt.Suggest{
				{Text: "aws_ec2_instance", Description: "Table"},
				{Text: "aws_s3_bucket", Description: "Table"},
				{Text: "gcp_compute_instance", Description: "Table"},
			},
			expectedCount: 3,
			shouldContain: []string{"aws_ec2_instance", "aws_s3_bucket", "gcp_compute_instance"},
		},
		"empty suggestions": {
			tableSuggestions: []prompt.Suggest{},
			expectedCount:    0,
		},
		"single suggestion": {
			tableSuggestions: []prompt.Suggest{
				{Text: "users", Description: "Table"},
			},
			expectedCount: 1,
			shouldContain: []string{"users"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			input := &CompleterInput{
				Query:            ".inspect",
				TableSuggestions: tc.tableSuggestions,
			}

			suggestions := inspectCompleter(input)

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

func TestCompleterInput(t *testing.T) {
	t.Run("CompleterInput struct", func(t *testing.T) {
		input := &CompleterInput{
			Query: ".inspect aws_ec2",
			TableSuggestions: []prompt.Suggest{
				{Text: "aws_ec2_instance", Description: "Table"},
			},
		}

		assert.Equal(t, ".inspect aws_ec2", input.Query)
		assert.Len(t, input.TableSuggestions, 1)
		assert.Equal(t, "aws_ec2_instance", input.TableSuggestions[0].Text)
	})
}
