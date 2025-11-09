package metaquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptSuggestions(t *testing.T) {
	t.Run("returns all metaquery suggestions", func(t *testing.T) {
		suggestions := PromptSuggestions()

		assert.NotEmpty(t, suggestions, "Should return non-empty suggestions")

		// Should contain common metaqueries
		expectedMetaqueries := []string{
			".help",
			".exit",
			".quit",
			".tables",
			".inspect",
			".connections",
			".header",
			".timing",
			".output",
			".cache",
			".search_path",
			".autocomplete",
		}

		for _, expected := range expectedMetaqueries {
			found := false
			for _, s := range suggestions {
				if s.Text == expected {
					found = true
					assert.NotEmpty(t, s.Description, "Metaquery %s should have description", expected)
					assert.Equal(t, expected, s.Output, "Metaquery %s output should match text", expected)
					break
				}
			}
			assert.True(t, found, "Expected to find metaquery: %s", expected)
		}
	})

	t.Run("suggestions are sorted alphabetically", func(t *testing.T) {
		suggestions := PromptSuggestions()

		// Check that suggestions are sorted
		for i := 1; i < len(suggestions); i++ {
			assert.True(t, suggestions[i-1].Text <= suggestions[i].Text,
				"Suggestions should be sorted: %s should come before or equal to %s",
				suggestions[i-1].Text, suggestions[i].Text)
		}
	})

	t.Run("all suggestions have required fields", func(t *testing.T) {
		suggestions := PromptSuggestions()

		for _, s := range suggestions {
			assert.NotEmpty(t, s.Text, "Suggestion should have text")
			assert.NotEmpty(t, s.Description, "Suggestion should have description: %s", s.Text)
			assert.NotEmpty(t, s.Output, "Suggestion should have output: %s", s.Text)
			assert.Equal(t, s.Text, s.Output, "For metaqueries, text and output should match: %s", s.Text)
		}
	})

	t.Run("count of metaqueries matches definitions", func(t *testing.T) {
		suggestions := PromptSuggestions()

		// Should have a reasonable number of metaqueries (at least the core ones)
		assert.GreaterOrEqual(t, len(suggestions), 12, "Should have at least 12 metaqueries")
	})

	t.Run("no duplicate suggestions", func(t *testing.T) {
		suggestions := PromptSuggestions()

		seen := make(map[string]bool)
		for _, s := range suggestions {
			assert.False(t, seen[s.Text], "Duplicate suggestion found: %s", s.Text)
			seen[s.Text] = true
		}
	})
}
