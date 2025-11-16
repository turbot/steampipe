package interactive

import (
	"sync"
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// TestInitialiseSchemaAndTableSuggestions_NilClient tests that initialiseSchemaAndTableSuggestions
// handles a nil client gracefully without panicking.
// This is a regression test for bug #4713.
func TestInitialiseSchemaAndTableSuggestions_NilClient(t *testing.T) {
	// Create an InteractiveClient with nil initData, which causes client() to return nil
	c := &InteractiveClient{
		initData:   nil, // This will cause client() to return nil
		suggestions: newAutocompleteSuggestions(),
		// Set schemaMetadata to non-nil so we get past the early return on line 43
		schemaMetadata: &db_common.SchemaMetadata{
			Schemas:             make(map[string]map[string]db_common.TableSchema),
			TemporarySchemaName: "temp",
		},
	}

	// Create an empty connection state map
	connectionStateMap := steampipeconfig.ConnectionStateMap{}

	// This should not panic - the function should handle nil client gracefully
	assert.NotPanics(t, func() {
		c.initialiseSchemaAndTableSuggestions(connectionStateMap)
	})
}

// TestAutocompleteSuggestions_ConcurrentSort tests that the sort() method
// can be called concurrently without data races.
// This is a regression test for bug #4711.
func TestAutocompleteSuggestions_ConcurrentSort(t *testing.T) {
	// Create suggestions with some data to sort
	suggestions := newAutocompleteSuggestions()
	suggestions.schemas = []prompt.Suggest{
		{Text: "schema2", Description: "Schema"},
		{Text: "schema1", Description: "Schema"},
		{Text: "schema3", Description: "Schema"},
	}
	suggestions.unqualifiedTables = []prompt.Suggest{
		{Text: "table2", Description: "Table"},
		{Text: "table1", Description: "Table"},
		{Text: "table3", Description: "Table"},
	}
	suggestions.tablesBySchema = map[string][]prompt.Suggest{
		"schema1": {
			{Text: "schema1.table2", Description: "Table"},
			{Text: "schema1.table1", Description: "Table"},
		},
		"schema2": {
			{Text: "schema2.table2", Description: "Table"},
			{Text: "schema2.table1", Description: "Table"},
		},
	}
	suggestions.queriesByMod = map[string][]prompt.Suggest{
		"mod1": {
			{Text: "query2", Description: "Query"},
			{Text: "query1", Description: "Query"},
		},
		"mod2": {
			{Text: "query2", Description: "Query"},
			{Text: "query1", Description: "Query"},
		},
	}

	// Call sort() concurrently from multiple goroutines
	// This should trigger the race detector if there's no synchronization
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

	// Verify that data is still valid (not corrupted)
	assert.Len(t, suggestions.schemas, 3)
	assert.Len(t, suggestions.unqualifiedTables, 3)
	assert.Len(t, suggestions.tablesBySchema, 2)
	assert.Len(t, suggestions.queriesByMod, 2)
}
