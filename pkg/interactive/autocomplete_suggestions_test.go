package interactive

import (
	"sync"
	"testing"

	"github.com/c-bata/go-prompt"
)

// TestAutoCompleteSuggestions_ConcurrentSort tests that sort() can be called
// concurrently without triggering data races.
// This test reproduces the race condition reported in issue #4716.
func TestAutoCompleteSuggestions_ConcurrentSort(t *testing.T) {
	// Create a populated autoCompleteSuggestions instance
	suggestions := newAutocompleteSuggestions()

	// Populate with test data
	suggestions.schemas = []prompt.Suggest{
		{Text: "public"},
		{Text: "aws"},
		{Text: "github"},
	}

	suggestions.unqualifiedTables = []prompt.Suggest{
		{Text: "table1"},
		{Text: "table2"},
		{Text: "table3"},
	}

	suggestions.unqualifiedQueries = []prompt.Suggest{
		{Text: "query1"},
		{Text: "query2"},
		{Text: "query3"},
	}

	suggestions.tablesBySchema["public"] = []prompt.Suggest{
		{Text: "users"},
		{Text: "accounts"},
	}

	suggestions.queriesByMod["aws"] = []prompt.Suggest{
		{Text: "aws_query1"},
		{Text: "aws_query2"},
	}

	// Call sort() concurrently from multiple goroutines
	// This should trigger a race condition if the sort() method is not thread-safe
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			suggestions.sort()
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// If we get here without panicking or race detector errors, the test passes
	// Note: This test will fail when run with -race flag if sort() is not thread-safe
}
