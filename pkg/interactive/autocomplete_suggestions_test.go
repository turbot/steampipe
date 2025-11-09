package interactive

import (
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/stretchr/testify/assert"
)

func TestNewAutocompleteSuggestions(t *testing.T) {
	t.Run("creates new suggestions with initialized maps", func(t *testing.T) {
		suggestions := newAutocompleteSuggestions()

		assert.NotNil(t, suggestions)
		assert.NotNil(t, suggestions.tablesBySchema)
		assert.NotNil(t, suggestions.queriesByMod)
		assert.Empty(t, suggestions.schemas)
		assert.Empty(t, suggestions.unqualifiedTables)
		assert.Empty(t, suggestions.unqualifiedQueries)
		assert.Empty(t, suggestions.mods)
	})
}

func TestAutocompleteSuggestions_Sort(t *testing.T) {
	t.Run("sorts schemas", func(t *testing.T) {
		suggestions := newAutocompleteSuggestions()
		suggestions.schemas = []prompt.Suggest{
			{Text: "zebra", Description: "Schema"},
			{Text: "alpha", Description: "Schema"},
			{Text: "middle", Description: "Schema"},
		}

		suggestions.sort()

		assert.Equal(t, "alpha", suggestions.schemas[0].Text)
		assert.Equal(t, "middle", suggestions.schemas[1].Text)
		assert.Equal(t, "zebra", suggestions.schemas[2].Text)
	})

	t.Run("sorts unqualified tables", func(t *testing.T) {
		suggestions := newAutocompleteSuggestions()
		suggestions.unqualifiedTables = []prompt.Suggest{
			{Text: "users", Description: "Table"},
			{Text: "accounts", Description: "Table"},
			{Text: "orders", Description: "Table"},
		}

		suggestions.sort()

		assert.Equal(t, "accounts", suggestions.unqualifiedTables[0].Text)
		assert.Equal(t, "orders", suggestions.unqualifiedTables[1].Text)
		assert.Equal(t, "users", suggestions.unqualifiedTables[2].Text)
	})

	t.Run("sorts unqualified queries", func(t *testing.T) {
		suggestions := newAutocompleteSuggestions()
		suggestions.unqualifiedQueries = []prompt.Suggest{
			{Text: "query_z", Description: "Query"},
			{Text: "query_a", Description: "Query"},
			{Text: "query_m", Description: "Query"},
		}

		suggestions.sort()

		assert.Equal(t, "query_a", suggestions.unqualifiedQueries[0].Text)
		assert.Equal(t, "query_m", suggestions.unqualifiedQueries[1].Text)
		assert.Equal(t, "query_z", suggestions.unqualifiedQueries[2].Text)
	})

	t.Run("sorts tables by schema", func(t *testing.T) {
		suggestions := newAutocompleteSuggestions()
		suggestions.tablesBySchema = map[string][]prompt.Suggest{
			"aws": {
				{Text: "aws.vpc", Description: "Table"},
				{Text: "aws.ec2", Description: "Table"},
				{Text: "aws.s3", Description: "Table"},
			},
		}

		suggestions.sort()

		tables := suggestions.tablesBySchema["aws"]
		assert.Equal(t, "aws.ec2", tables[0].Text)
		assert.Equal(t, "aws.s3", tables[1].Text)
		assert.Equal(t, "aws.vpc", tables[2].Text)
	})

	t.Run("sorts queries by mod", func(t *testing.T) {
		suggestions := newAutocompleteSuggestions()
		suggestions.queriesByMod = map[string][]prompt.Suggest{
			"mymod": {
				{Text: "mymod.query_z", Description: "Query"},
				{Text: "mymod.query_a", Description: "Query"},
				{Text: "mymod.query_m", Description: "Query"},
			},
		}

		suggestions.sort()

		queries := suggestions.queriesByMod["mymod"]
		assert.Equal(t, "mymod.query_a", queries[0].Text)
		assert.Equal(t, "mymod.query_m", queries[1].Text)
		assert.Equal(t, "mymod.query_z", queries[2].Text)
	})

	t.Run("handles empty suggestions", func(t *testing.T) {
		suggestions := newAutocompleteSuggestions()

		// Should not panic with empty slices
		assert.NotPanics(t, func() {
			suggestions.sort()
		})
	})

	t.Run("sorts all types together", func(t *testing.T) {
		suggestions := newAutocompleteSuggestions()
		suggestions.schemas = []prompt.Suggest{
			{Text: "z", Description: "Schema"},
			{Text: "a", Description: "Schema"},
		}
		suggestions.unqualifiedTables = []prompt.Suggest{
			{Text: "table_z", Description: "Table"},
			{Text: "table_a", Description: "Table"},
		}
		suggestions.unqualifiedQueries = []prompt.Suggest{
			{Text: "query_z", Description: "Query"},
			{Text: "query_a", Description: "Query"},
		}
		suggestions.tablesBySchema = map[string][]prompt.Suggest{
			"conn1": {
				{Text: "z", Description: "Table"},
				{Text: "a", Description: "Table"},
			},
		}
		suggestions.queriesByMod = map[string][]prompt.Suggest{
			"mod1": {
				{Text: "z", Description: "Query"},
				{Text: "a", Description: "Query"},
			},
		}

		suggestions.sort()

		// Verify all are sorted
		assert.Equal(t, "a", suggestions.schemas[0].Text)
		assert.Equal(t, "table_a", suggestions.unqualifiedTables[0].Text)
		assert.Equal(t, "query_a", suggestions.unqualifiedQueries[0].Text)
		assert.Equal(t, "a", suggestions.tablesBySchema["conn1"][0].Text)
		assert.Equal(t, "a", suggestions.queriesByMod["mod1"][0].Text)
	})
}
