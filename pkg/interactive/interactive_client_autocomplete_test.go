package interactive

import (
	"testing"

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
